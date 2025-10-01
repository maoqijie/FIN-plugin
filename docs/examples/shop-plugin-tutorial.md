# 示例插件教程：商店插件

本教程将手把手教你如何编写一个完整的商店插件，实现玩家通过积分购买/出售物品的功能。

## 目标功能

- 玩家输入 `购买` 或 `buy` 打开购买菜单
- 玩家输入 `收购` 或 `sell` 打开出售菜单
- 使用记分板作为货币系统
- 交互式选择物品和数量
- 自动拦截商店交互消息，不转发到 QQ 群

## 第一步：创建插件基础结构

首先在 `cmd/main/Plugin/` 下创建插件目录 `shop`：

```bash
mkdir -p cmd/main/Plugin/shop
cd cmd/main/Plugin/shop
```

创建 `plugin.yaml` 配置文件：

```yaml
name: shop
displayName: 商店插件
author: 你的名字
version: 1.0.0
description: 玩家商店系统，支持购买和出售物品
entry: main.so
```

创建 `go.mod` 文件：

```go
module shop-plugin

go 1.25.1

require github.com/maoqijie/FIN-plugin v0.0.0

replace github.com/maoqijie/FIN-plugin => ../../../../PluginFramework
```

## 第二步：定义插件结构

创建 `main.go`，首先定义插件的数据结构：

```go
package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	sdk "github.com/maoqijie/FIN-plugin/sdk"
)

// ShopPlugin 商店插件主结构
type ShopPlugin struct {
	ctx             *sdk.Context
	scoreboardName  string                     // 货币计分板名称
	sellItems       map[string]*ShopItem       // 可购买的物品
	buybackItems    map[string]*ShopItem       // 可收购的物品
	waitingForInput map[string]*PlayerInputState // 等待玩家输入的状态
}

// ShopItem 商品信息
type ShopItem struct {
	ID    string // Minecraft 物品 ID
	Price int    // 价格
}

// PlayerInputState 玩家输入状态
type PlayerInputState struct {
	Action     string    // "buy" 或 "sell"
	ItemName   string    // 物品名称
	Item       *ShopItem // 物品信息
	WaitingFor string    // "item" 或 "amount"
	Timestamp  time.Time // 创建时间，用于超时清理
}
```

## 第三步：实现插件生命周期方法

### Init 方法：初始化插件

```go
func (p *ShopPlugin) Init(ctx *sdk.Context) error {
	p.ctx = ctx
	p.waitingForInput = make(map[string]*PlayerInputState)
	p.scoreboardName = "money" // 货币计分板名称

	// 初始化可购买商品列表
	p.sellItems = map[string]*ShopItem{
		"钻石":  {ID: "minecraft:diamond", Price: 100},
		"铁锭":  {ID: "minecraft:iron_ingot", Price: 10},
		"金锭":  {ID: "minecraft:gold_ingot", Price: 50},
		"绿宝石": {ID: "minecraft:emerald", Price: 200},
	}

	// 初始化可收购商品列表（价格通常低于出售价）
	p.buybackItems = map[string]*ShopItem{
		"钻石":  {ID: "minecraft:diamond", Price: 80},
		"铁锭":  {ID: "minecraft:iron_ingot", Price: 8},
		"金锭":  {ID: "minecraft:gold_ingot", Price: 40},
		"绿宝石": {ID: "minecraft:emerald", Price: 160},
	}

	// 监听聊天消息
	ctx.ListenChat(p.onChat)

	ctx.Logf("商店插件已加载")
	return nil
}
```

### Start 方法：启动插件

```go
func (p *ShopPlugin) Start() error {
	p.ctx.Logf("商店插件已启动")

	// 启动定时清理任务（使用 goroutine，不阻塞）
	go p.startCleanupTimer()

	return nil
}
```

### Stop 方法：停止插件

```go
func (p *ShopPlugin) Stop() error {
	p.ctx.Logf("商店插件已停止")
	return nil
}
```

## 第四步：实现聊天事件处理

```go
func (p *ShopPlugin) onChat(event *sdk.ChatEvent) {
	// 只处理聊天类型的消息
	if event.TextType != 1 {
		return
	}

	playerName := event.Sender
	message := strings.TrimSpace(event.Message)

	// 检查是否有等待输入的状态
	if state, exists := p.waitingForInput[playerName]; exists {
		if state.WaitingFor == "amount" {
			p.handleAmountInput(playerName, message, state)
			// 拦截此消息，不转发到 QQ
			event.Cancelled = true
			return
		} else if state.WaitingFor == "item" {
			p.handleItemSelection(playerName, message, state)
			// 拦截此消息，不转发到 QQ
			event.Cancelled = true
			return
		}
	}

	// 处理商店命令
	switch message {
	case "购买", "buy":
		p.showShop(playerName, "buy")
		// 拦截此消息，不转发到 QQ
		event.Cancelled = true
	case "收购", "sell":
		p.showShop(playerName, "sell")
		// 拦截此消息，不转发到 QQ
		event.Cancelled = true
	}
}
```

**关键点说明：**
- `event.TextType != 1` 过滤非聊天消息
- `event.Cancelled = true` 拦截消息，防止转发到 QQ 群
- 使用状态机管理多步骤交互

## 第五步：实现商店菜单显示

```go
func (p *ShopPlugin) showShop(playerName string, action string) {
	var items map[string]*ShopItem
	var title string

	if action == "buy" {
		items = p.sellItems
		title = "§a=== 商店购买列表 ==="
	} else {
		items = p.buybackItems
		title = "§e=== 商店收购列表 ==="
	}

	// 向玩家显示标题
	p.ctx.GameUtils().SayTo(playerName, title)

	// 显示所有商品
	index := 1
	for itemName, item := range items {
		msg := fmt.Sprintf("§f%d. §b%s §f- §6%d §f积分", index, itemName, item.Price)
		p.ctx.GameUtils().SayTo(playerName, msg)
		index++
	}

	p.ctx.GameUtils().SayTo(playerName, "§7请输入物品名称进行交易")

	// 记录玩家状态：等待物品名称输入
	p.waitingForInput[playerName] = &PlayerInputState{
		Action:     action,
		WaitingFor: "item",
		Timestamp:  time.Now(),
	}
}
```

**颜色代码说明：**
- `§a` 绿色
- `§e` 黄色
- `§f` 白色
- `§b` 青色
- `§6` 金色
- `§7` 灰色

## 第六步：处理物品选择

```go
func (p *ShopPlugin) handleItemSelection(playerName string, input string, state *PlayerInputState) {
	var items map[string]*ShopItem
	if state.Action == "buy" {
		items = p.sellItems
	} else {
		items = p.buybackItems
	}

	// 检查物品是否存在
	item, exists := items[input]
	if !exists {
		p.ctx.GameUtils().SayTo(playerName, "§c物品不存在，请重新输入")
		return
	}

	// 更新状态：等待数量输入
	state.ItemName = input
	state.Item = item
	state.WaitingFor = "amount"
	state.Timestamp = time.Now()

	if state.Action == "buy" {
		p.ctx.GameUtils().SayTo(playerName, fmt.Sprintf("§a你要购买多少个 §b%s §a? (单价: §6%d §a积分)", input, item.Price))
	} else {
		p.ctx.GameUtils().SayTo(playerName, fmt.Sprintf("§e你要出售多少个 §b%s §e? (单价: §6%d §e积分)", input, item.Price))
	}
}
```

## 第七步：处理数量输入

```go
func (p *ShopPlugin) handleAmountInput(playerName string, input string, state *PlayerInputState) {
	// 解析数量
	amount, err := strconv.Atoi(input)
	if err != nil || amount <= 0 {
		p.ctx.GameUtils().SayTo(playerName, "§c请输入有效的数量（正整数）")
		return
	}

	// 执行交易
	if state.Action == "buy" {
		p.executeBuy(playerName, state.ItemName, state.Item, amount)
	} else {
		p.executeSell(playerName, state.ItemName, state.Item, amount)
	}

	// 清除玩家状态
	delete(p.waitingForInput, playerName)
}
```

## 第八步：实现购买逻辑

```go
func (p *ShopPlugin) executeBuy(playerName string, itemName string, item *ShopItem, amount int) {
	totalPrice := item.Price * amount

	// 获取玩家余额
	balance, err := p.ctx.GameUtils().GetScore(p.scoreboardName, playerName, 5.0)
	if err != nil {
		p.ctx.GameUtils().SayTo(playerName, "§c无法获取你的余额")
		return
	}

	// 检查余额是否足够
	if balance < totalPrice {
		p.ctx.GameUtils().SayTo(playerName, fmt.Sprintf("§c余额不足！需要 §6%d §c积分，你只有 §6%d §c积分", totalPrice, balance))
		return
	}

	// 扣除积分
	cmd := fmt.Sprintf("scoreboard players remove \"%s\" %s %d", playerName, p.scoreboardName, totalPrice)
	p.ctx.GameUtils().SendCommand(cmd)

	// 给予物品
	giveCmd := fmt.Sprintf("give \"%s\" %s %d", playerName, item.ID, amount)
	p.ctx.GameUtils().SendCommand(giveCmd)

	// 显示成功消息
	successMsg := fmt.Sprintf("§a购买成功！花费 §6%d §a积分购买了 §b%d §a个 §b%s", totalPrice, amount, itemName)
	p.ctx.GameUtils().SayTo(playerName, successMsg)

	newBalance := balance - totalPrice
	p.ctx.GameUtils().SayTo(playerName, fmt.Sprintf("§7剩余余额: §6%d §7积分", newBalance))
}
```

## 第九步：实现出售逻辑

```go
func (p *ShopPlugin) executeSell(playerName string, itemName string, item *ShopItem, amount int) {
	totalPrice := item.Price * amount

	// 获取玩家物品数量
	itemCount, err := p.ctx.GameUtils().GetItem(playerName, item.ID, 0)
	if err != nil {
		p.ctx.GameUtils().SayTo(playerName, "§c无法获取你的物品数量")
		return
	}

	// 检查物品数量是否足够
	if itemCount < amount {
		p.ctx.GameUtils().SayTo(playerName, fmt.Sprintf("§c物品不足！需要 §b%d §c个，你只有 §b%d §c个", amount, itemCount))
		return
	}

	// 清除物品
	clearCmd := fmt.Sprintf("clear \"%s\" %s 0 %d", playerName, item.ID, amount)
	p.ctx.GameUtils().SendCommand(clearCmd)

	// 增加积分
	addCmd := fmt.Sprintf("scoreboard players add \"%s\" %s %d", playerName, p.scoreboardName, totalPrice)
	p.ctx.GameUtils().SendCommand(addCmd)

	// 显示成功消息
	successMsg := fmt.Sprintf("§a出售成功！出售了 §b%d §a个 §b%s§a，获得 §6%d §a积分", amount, itemName, totalPrice)
	p.ctx.GameUtils().SayTo(playerName, successMsg)

	// 等待计分板更新后显示余额
	time.Sleep(500 * time.Millisecond)
	balance, err := p.ctx.GameUtils().GetScore(p.scoreboardName, playerName, 5.0)
	if err == nil {
		p.ctx.GameUtils().SayTo(playerName, fmt.Sprintf("§7当前余额: §6%d §7积分", balance))
	}
}
```

## 第十步：实现定时清理

```go
func (p *ShopPlugin) startCleanupTimer() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		p.cleanupExpiredStates()
	}
}

func (p *ShopPlugin) cleanupExpiredStates() {
	now := time.Now()
	for playerName, state := range p.waitingForInput {
		// 60 秒未操作则超时
		if now.Sub(state.Timestamp) > 60*time.Second {
			delete(p.waitingForInput, playerName)
			p.ctx.GameUtils().SayTo(playerName, "§c商店操作已超时，请重新开始")
		}
	}
}
```

## 第十一步：导出插件

```go
func NewPlugin() sdk.Plugin {
	return &ShopPlugin{}
}
```

## 完整代码

将以上所有代码片段组合到 `main.go` 中即可。完整代码约 250 行。

## 编译与测试

### 1. 编译插件

在主程序运行时，输入控制台命令：

```
reload
```

系统会自动检测到 `.go` 源文件并编译为 `.so` 插件。

### 2. 测试功能

1. 在游戏中输入 `购买` 或 `buy`
2. 系统显示可购买物品列表
3. 输入物品名称，如 `钻石`
4. 输入购买数量，如 `5`
5. 系统执行交易并显示结果

### 3. 前置条件

使用前需要创建货币计分板：

```
/scoreboard objectives add money dummy 积分
/scoreboard players set @a money 1000
```

## 扩展功能建议

### 1. 添加配置文件支持

```go
func (p *ShopPlugin) Init(ctx *sdk.Context) error {
	// 加载配置
	config := ctx.Config()
	defaultConfig := map[string]interface{}{
		"scoreboard": "money",
		"items": map[string]interface{}{
			"diamond": map[string]interface{}{
				"buy_price":  100,
				"sell_price": 80,
			},
		},
	}
	cfg := config.GetConfig(defaultConfig)

	// 解析配置
	if scoreboard, ok := cfg["scoreboard"].(string); ok {
		p.scoreboardName = scoreboard
	}

	// ... 其他初始化逻辑
}
```

### 2. 添加控制台命令

```go
func (p *ShopPlugin) Init(ctx *sdk.Context) error {
	// ... 其他初始化

	// 注册控制台命令
	ctx.RegisterConsoleCommand(sdk.ConsoleCommand{
		Triggers:    []string{"shop", "商店"},
		Usage:       "shop reload - 重新加载商店配置",
		Description: "商店插件管理命令",
		Handler:     p.handleConsoleCommand,
	})

	return nil
}

func (p *ShopPlugin) handleConsoleCommand(args []string) error {
	if len(args) > 0 && args[0] == "reload" {
		p.ctx.Logf("重新加载商店配置...")
		// 重新加载逻辑
		return nil
	}
	return nil
}
```

### 3. 添加交易记录

```go
type TransactionRecord struct {
	PlayerName string
	Action     string // "buy" 或 "sell"
	ItemName   string
	Amount     int
	Price      int
	Timestamp  time.Time
}

func (p *ShopPlugin) recordTransaction(record TransactionRecord) {
	// 保存到文件
	dataPath := p.ctx.FormatDataPath("transactions.json")
	// ... 序列化并保存
}
```

## 常见问题

### Q1: 为什么玩家输入后没有反应？

**A:** 检查以下几点：
1. 确保 `event.TextType == 1`（聊天消息）
2. 确认玩家名称没有空格或特殊字符
3. 查看控制台是否有错误日志

### Q2: 如何修改商品价格？

**A:** 修改 `Init()` 方法中的 `sellItems` 和 `buybackItems` 映射：

```go
p.sellItems = map[string]*ShopItem{
	"钻石": {ID: "minecraft:diamond", Price: 150}, // 修改价格
}
```

### Q3: 如何添加新商品？

**A:** 在商品映射中添加新条目：

```go
p.sellItems["下界之星"] = &ShopItem{
	ID:    "minecraft:nether_star",
	Price: 5000,
}
```

### Q4: 消息为什么转发到了 QQ 群？

**A:** 确保在处理商店消息时设置了 `event.Cancelled = true`。

## 总结

通过本教程，你学会了：

1. ✅ 创建插件基础结构
2. ✅ 实现插件生命周期方法
3. ✅ 监听和处理聊天事件
4. ✅ 使用状态机管理多步交互
5. ✅ 调用游戏命令和 API
6. ✅ 实现消息拦截机制
7. ✅ 定时任务和状态清理

这些技能可以应用到其他类型的插件开发中，如任务系统、传送系统、权限管理等。

## 下一步学习

- [插件 API 参考文档](../api/overview.md)
- [游戏工具 API](../api/game-utils.md)
- [数据持久化](../api/data-management.md)
- [插件间通信](../api/plugin-api.md)
