# 示例插件教程：商店插件（跨平台版）

本教程将手把手教你如何编写一个完整的跨平台商店插件，实现玩家通过积分购买/出售物品的功能。

## 目标功能

- 玩家输入 `购买` 或 `buy` 打开购买菜单
- 玩家输入 `收购` 或 `sell` 打开出售菜单
- 使用记分板作为货币系统
- 交互式选择物品和数量
- 自动拦截商店交互消息，不转发到 QQ 群
- **跨平台支持**：Windows、Linux、macOS、Android

## 第一步：使用 create 命令创建插件

在主程序控制台直接创建插件：

```bash
# 在控制台输入
create

# 按提示输入：
插件名字: shop
显示名: 商店系统
描述: 玩家商店系统，支持购买和出售物品
```

系统会自动：
1. 生成插件目录结构
2. 创建 `plugin.yaml` 和 `go.mod`
3. 生成基础的 `main.go` 模板
4. 下载依赖并构建当前平台的可执行文件
5. 自动加载插件

## 第二步：定义插件结构

修改生成的 `main.go`，添加商店相关的数据结构：

```go
package main

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/go-plugin"
	sdk "github.com/maoqijie/FIN-plugin/sdk"
)

// ShopPlugin 商店插件主结构
type ShopPlugin struct {
	ctx             *sdk.Context
	scoreboardName  string                       // 货币计分板名称
	sellItems       map[string]*ShopItem         // 可购买的物品
	buybackItems    map[string]*ShopItem         // 可收购的物品
	waitingForInput map[string]*PlayerInputState // 等待玩家输入的状态
	mu              sync.RWMutex                 // 并发安全锁
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

	// 注册控制台命令
	ctx.RegisterConsoleCommand(sdk.ConsoleCommand{
		Name:        "shop",
		Triggers:    []string{"shop", "商店"},
		Usage:       "shop reload - 重新加载商店配置",
		Description: "商店插件管理命令",
		Handler: func(args []string) error {
			if len(args) > 0 && args[0] == "reload" {
				ctx.LogInfo("重新加载商店配置...")
			}
			return nil
		},
	})

	ctx.LogInfo("商店插件已加载")
	return nil
}
```

### Start 方法：启动插件

```go
func (p *ShopPlugin) Start() error {
	p.ctx.LogSuccess("商店插件已启动")

	// 启动定时清理任务（使用 goroutine，不阻塞）
	go p.startCleanupTimer()

	return nil
}
```

### Stop 方法：停止插件

```go
func (p *ShopPlugin) Stop() error {
	p.ctx.LogWarning("商店插件已停止")
	return nil
}
```

### GetInfo 方法：返回插件信息

```go
func (p *ShopPlugin) GetInfo() sdk.PluginInfo {
	return sdk.PluginInfo{
		Name:        "shop",
		DisplayName: "商店系统",
		Version:     "1.0.0",
		Description: "玩家商店系统，支持购买和出售物品",
		Author:      "猫七街",
	}
}
```

## 第四步：实现聊天事件处理

```go
func (p *ShopPlugin) onChat(event *sdk.ChatEvent) {
	playerName := event.Sender
	message := strings.TrimSpace(event.Message)

	// 使用锁保护并发访问
	p.mu.RLock()
	state, exists := p.waitingForInput[playerName]
	p.mu.RUnlock()

	// 检查是否有等待输入的状态
	if exists {
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
- `event.Cancelled = true` 拦截消息，防止转发到 QQ 群
- 使用 `sync.RWMutex` 保证并发安全
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

	gu := p.ctx.GameUtils()
	if gu == nil {
		return
	}

	// 向玩家显示标题
	gu.SayTo(playerName, title)

	// 显示所有商品
	index := 1
	for itemName, item := range items {
		msg := fmt.Sprintf("§f%d. §b%s §f- §6%d §f积分", index, itemName, item.Price)
		gu.SayTo(playerName, msg)
		index++
	}

	gu.SayTo(playerName, "§7请输入物品名称进行交易")

	// 记录玩家状态：等待物品名称输入
	p.mu.Lock()
	p.waitingForInput[playerName] = &PlayerInputState{
		Action:     action,
		WaitingFor: "item",
		Timestamp:  time.Now(),
	}
	p.mu.Unlock()
}
```

**颜色代码说明：**
- `§a` 绿色
- `§e` 黄色
- `§f` 白色
- `§b` 青色
- `§6` 金色
- `§7` 灰色
- `§c` 红色

## 第六步：处理物品选择

```go
func (p *ShopPlugin) handleItemSelection(playerName string, input string, state *PlayerInputState) {
	var items map[string]*ShopItem
	if state.Action == "buy" {
		items = p.sellItems
	} else {
		items = p.buybackItems
	}

	gu := p.ctx.GameUtils()
	if gu == nil {
		return
	}

	// 检查物品是否存在
	item, exists := items[input]
	if !exists {
		gu.SayTo(playerName, "§c物品不存在，请重新输入")
		return
	}

	// 更新状态：等待数量输入
	p.mu.Lock()
	state.ItemName = input
	state.Item = item
	state.WaitingFor = "amount"
	state.Timestamp = time.Now()
	p.mu.Unlock()

	if state.Action == "buy" {
		gu.SayTo(playerName, fmt.Sprintf("§a你要购买多少个 §b%s §a? (单价: §6%d §a积分)", input, item.Price))
	} else {
		gu.SayTo(playerName, fmt.Sprintf("§e你要出售多少个 §b%s §e? (单价: §6%d §e积分)", input, item.Price))
	}
}
```

## 第七步：处理数量输入

```go
func (p *ShopPlugin) handleAmountInput(playerName string, input string, state *PlayerInputState) {
	gu := p.ctx.GameUtils()
	if gu == nil {
		return
	}

	// 解析数量
	amount, err := strconv.Atoi(input)
	if err != nil || amount <= 0 {
		gu.SayTo(playerName, "§c请输入有效的数量（正整数）")
		return
	}

	// 执行交易
	if state.Action == "buy" {
		p.executeBuy(playerName, state.ItemName, state.Item, amount)
	} else {
		p.executeSell(playerName, state.ItemName, state.Item, amount)
	}

	// 清除玩家状态
	p.mu.Lock()
	delete(p.waitingForInput, playerName)
	p.mu.Unlock()
}
```

## 第八步：实现购买逻辑

```go
func (p *ShopPlugin) executeBuy(playerName string, itemName string, item *ShopItem, amount int) {
	totalPrice := item.Price * amount
	gu := p.ctx.GameUtils()
	if gu == nil {
		return
	}

	// 获取玩家余额
	balance, err := gu.GetScore(p.scoreboardName, playerName, 5.0)
	if err != nil {
		gu.SayTo(playerName, "§c无法获取你的余额")
		p.ctx.LogError("获取玩家 %s 余额失败: %v", playerName, err)
		return
	}

	// 检查余额是否足够
	if balance < totalPrice {
		gu.SayTo(playerName, fmt.Sprintf("§c余额不足！需要 §6%d §c积分，你只有 §6%d §c积分", totalPrice, balance))
		return
	}

	// 扣除积分
	cmd := fmt.Sprintf("scoreboard players remove \"%s\" %s %d", playerName, p.scoreboardName, totalPrice)
	gu.SendCommand(cmd)

	// 给予物品
	giveCmd := fmt.Sprintf("give \"%s\" %s %d", playerName, item.ID, amount)
	gu.SendCommand(giveCmd)

	// 显示成功消息
	successMsg := fmt.Sprintf("§a购买成功！花费 §6%d §a积分购买了 §b%d §a个 §b%s", totalPrice, amount, itemName)
	gu.SayTo(playerName, successMsg)

	newBalance := balance - totalPrice
	gu.SayTo(playerName, fmt.Sprintf("§7剩余余额: §6%d §7积分", newBalance))
}
```

## 第九步：实现出售逻辑

```go
func (p *ShopPlugin) executeSell(playerName string, itemName string, item *ShopItem, amount int) {
	totalPrice := item.Price * amount
	gu := p.ctx.GameUtils()
	if gu == nil {
		return
	}

	// 获取玩家物品数量
	itemCount, err := gu.GetItem(playerName, item.ID, 0)
	if err != nil {
		gu.SayTo(playerName, "§c无法获取你的物品数量")
		p.ctx.LogError("获取玩家 %s 物品数量失败: %v", playerName, err)
		return
	}

	// 检查物品数量是否足够
	if itemCount < amount {
		gu.SayTo(playerName, fmt.Sprintf("§c物品不足！需要 §b%d §c个，你只有 §b%d §c个", amount, itemCount))
		return
	}

	// 清除物品
	clearCmd := fmt.Sprintf("clear \"%s\" %s 0 %d", playerName, item.ID, amount)
	gu.SendCommand(clearCmd)

	// 增加积分
	addCmd := fmt.Sprintf("scoreboard players add \"%s\" %s %d", playerName, p.scoreboardName, totalPrice)
	gu.SendCommand(addCmd)

	// 显示成功消息
	successMsg := fmt.Sprintf("§a出售成功！出售了 §b%d §a个 §b%s§a，获得 §6%d §a积分", amount, itemName, totalPrice)
	gu.SayTo(playerName, successMsg)

	// 等待计分板更新后显示余额
	time.Sleep(500 * time.Millisecond)
	balance, err := gu.GetScore(p.scoreboardName, playerName, 5.0)
	if err == nil {
		gu.SayTo(playerName, fmt.Sprintf("§7当前余额: §6%d §7积分", balance))
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
	gu := p.ctx.GameUtils()

	p.mu.Lock()
	defer p.mu.Unlock()

	for playerName, state := range p.waitingForInput {
		// 60 秒未操作则超时
		if now.Sub(state.Timestamp) > 60*time.Second {
			delete(p.waitingForInput, playerName)
			if gu != nil {
				gu.SayTo(playerName, "§c商店操作已超时，请重新开始")
			}
		}
	}
}
```

## 第十一步：导出插件（gRPC）

```go
func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: sdk.HandshakeConfig,
		Plugins: map[string]plugin.Plugin{
			"plugin": &sdk.PluginGRPC{Impl: &ShopPlugin{}},
		},
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
```

## 完整的 plugin.yaml

```yaml
name: shop
displayName: 商店系统
version: 1.0.0
description: 玩家商店系统，支持购买和出售物品
author: 猫七街
source: local
sdkVersion: 0.1.0

# 跨平台配置
platform:
  windows_amd64: shop.exe
  windows_arm64: shop.exe
  linux_amd64: shop
  linux_arm64: shop
  darwin_amd64: shop
  darwin_arm64: shop
  android_arm64: shop
```

## 编译与测试

### 1. 自动构建（推荐）

插件创建时已自动构建，修改代码后：

```bash
# 在控制台输入
reload
```

系统会自动检测到源码变化并重新编译。

### 2. 手动跨平台构建

使用生成的 `build.sh` 脚本：

```bash
# 构建所有平台
./build.sh

# 产物：
# - shop.exe (Windows)
# - shop (Linux/macOS/Android)
```

### 3. 测试功能

1. 在游戏中输入 `购买` 或 `buy`
2. 系统显示可购买物品列表
3. 输入物品名称，如 `钻石`
4. 输入购买数量，如 `5`
5. 系统执行交易并显示结果

### 4. 前置条件

使用前需要创建货币计分板：

```
/scoreboard objectives add money dummy 积分
/scoreboard players set @a money 1000
```

## 扩展功能建议

### 1. 添加数据持久化

```go
import (
	"encoding/json"
	"os"
	"path/filepath"
)

type ShopConfig struct {
	Scoreboard string               `json:"scoreboard"`
	SellItems  map[string]*ShopItem `json:"sell_items"`
	BuyItems   map[string]*ShopItem `json:"buy_items"`
}

func (p *ShopPlugin) loadConfig() error {
	pluginDir := p.ctx.GetPluginDir()
	configFile := filepath.Join(pluginDir, "config.json")

	data, err := os.ReadFile(configFile)
	if err != nil {
		// 使用默认配置
		return p.saveDefaultConfig()
	}

	var config ShopConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return err
	}

	p.scoreboardName = config.Scoreboard
	p.sellItems = config.SellItems
	p.buybackItems = config.BuyItems
	return nil
}

func (p *ShopPlugin) saveDefaultConfig() error {
	pluginDir := p.ctx.GetPluginDir()
	configFile := filepath.Join(pluginDir, "config.json")

	config := ShopConfig{
		Scoreboard: "money",
		SellItems:  p.sellItems,
		BuyItems:   p.buybackItems,
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configFile, data, 0644)
}
```

### 2. 添加交易记录

```go
type TransactionRecord struct {
	PlayerName string    `json:"player_name"`
	Action     string    `json:"action"` // "buy" 或 "sell"
	ItemName   string    `json:"item_name"`
	Amount     int       `json:"amount"`
	Price      int       `json:"price"`
	Timestamp  time.Time `json:"timestamp"`
}

func (p *ShopPlugin) recordTransaction(record TransactionRecord) {
	pluginDir := p.ctx.GetPluginDir()
	recordFile := filepath.Join(pluginDir, "transactions.json")

	// 读取现有记录
	var records []TransactionRecord
	if data, err := os.ReadFile(recordFile); err == nil {
		json.Unmarshal(data, &records)
	}

	// 添加新记录
	records = append(records, record)

	// 保存
	data, _ := json.MarshalIndent(records, "", "  ")
	os.WriteFile(recordFile, data, 0644)
}
```

### 3. 添加 QQ 群通知

```go
func (p *ShopPlugin) executeBuy(playerName string, itemName string, item *ShopItem, amount int) {
	// ... 原有购买逻辑 ...

	// 发送到 QQ 群
	qq := p.ctx.QQUtils()
	if qq != nil {
		msg := fmt.Sprintf("玩家 %s 购买了 %d 个 %s，花费 %d 积分",
			playerName, amount, itemName, totalPrice)
		qq.SendGroupMessage(12345678, msg) // 替换为你的群号
	}
}
```

## 常见问题

### Q1: 为什么跨平台插件比传统插件慢？

**A:** 跨平台插件基于 gRPC，有轻微的通信延迟（~0.1-0.5ms），但换来了全平台支持和进程隔离。对于商店这类交互频率较低的插件，性能影响可忽略。

### Q2: 如何修改商品价格？

**A:** 方法 1：直接修改 `Init()` 中的代码后 `reload`

方法 2：使用配置文件（见扩展功能）

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

### Q5: 并发安全问题？

**A:** 本教程已使用 `sync.RWMutex` 保证并发安全，多个玩家同时操作不会冲突。

## 跨平台部署

### Windows 部署

```powershell
# 构建
$env:GOOS="windows"; $env:GOARCH="amd64"; go build -o shop.exe

# 部署
copy shop.exe plugins\shop\
```

### Linux/macOS 部署

```bash
# 构建
GOOS=linux GOARCH=amd64 go build -o shop

# 部署
chmod +x shop
cp shop plugins/shop/
```

### Android 部署（Termux）

```bash
# 构建
GOOS=android GOARCH=arm64 go build -o shop

# 部署到 Android
adb push shop /sdcard/plugins/shop/
```

## 总结

通过本教程，你学会了：

1. ✅ 使用 `create` 命令快速创建跨平台插件
2. ✅ 实现插件生命周期方法（Init/Start/Stop/GetInfo）
3. ✅ 监听和处理聊天事件
4. ✅ 使用状态机管理多步交互
5. ✅ 调用游戏命令和 API
6. ✅ 实现消息拦截机制
7. ✅ 定时任务和状态清理
8. ✅ 并发安全编程
9. ✅ 跨平台构建和部署

这些技能可以应用到其他类型的插件开发中，如任务系统、传送系统、权限管理等。

## 下一步学习

- [跨平台插件开发指南](../../CROSS_PLATFORM_PLUGIN_GUIDE.md)
- [插件 API 参考文档](../api/overview.md)
- [游戏工具 API](../api/game-utils.md)
- [数据持久化](../api/data-management.md)
- [插件市场发布](../../PLUGIN_MARKET_README.md)
