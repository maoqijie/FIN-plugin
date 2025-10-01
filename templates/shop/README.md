# 商店插件示例

这是一个完整的商店系统示例插件，参考 ToolDelta 的商店插件设计。

## 功能特性

### 1. 双向交易系统
- **购买物品**：玩家使用积分购买商店物品
- **出售物品**：玩家出售物品获得积分

### 2. 交互式菜单
- 显示可交易物品列表
- 实时显示物品价格
- 引导式交易流程

### 3. 完整的交易流程
1. 玩家输入命令（购买/收购）
2. 显示物品列表
3. 玩家选择物品
4. 玩家输入数量
5. 系统验证余额/物品
6. 执行交易并反馈结果

### 4. 安全机制
- 余额检查（购买时）
- 物品数量检查（出售时）
- 输入超时自动清理（60秒）
- 输入验证（数量必须为正整数）

## 使用方法

### 购买物品

1. 在游戏中输入：`购买` 或 `buy`
2. 查看物品列表，输入物品名称（如：`钻石`）
3. 输入购买数量（如：`10`）
4. 系统自动扣除积分并给予物品

**示例对话**：
```
玩家: 购买
系统: === 商店购买列表 ===
      1. 钻石 - 100 积分
      2. 铁锭 - 10 积分
      3. 金锭 - 50 积分
      4. 绿宝石 - 200 积分
      请输入物品名称进行交易
玩家: 钻石
系统: 你要购买多少个 钻石? (单价: 100 积分)
玩家: 5
系统: 购买成功！花费 500 积分购买了 5 个 钻石
      剩余余额: 1500 积分
```

### 出售物品

1. 在游戏中输入：`收购` 或 `sell`
2. 查看收购列表，输入物品名称（如：`铁锭`）
3. 输入出售数量（如：`64`）
4. 系统自动清除物品并增加积分

**示例对话**：
```
玩家: 收购
系统: === 商店收购列表 ===
      1. 钻石 - 80 积分
      2. 铁锭 - 8 积分
      3. 金锭 - 40 积分
      4. 绿宝石 - 160 积分
      请输入物品名称进行交易
玩家: 铁锭
系统: 你要出售多少个 铁锭? (单价: 8 积分)
玩家: 64
系统: 出售成功！出售了 64 个 铁锭，获得 512 积分
      当前余额: 2012 积分
```

## 配置说明

配置文件位于 `plugins/shop/config.json`：

```json
{
  "货币计分板名": "money",
  "出售": {
    "钻石": {
      "ID": "minecraft:diamond",
      "价格": 100
    },
    "铁锭": {
      "ID": "minecraft:iron_ingot",
      "价格": 10
    },
    "金锭": {
      "ID": "minecraft:gold_ingot",
      "价格": 50
    },
    "绿宝石": {
      "ID": "minecraft:emerald",
      "价格": 200
    }
  },
  "收购": {
    "钻石": {
      "ID": "minecraft:diamond",
      "价格": 80
    },
    "铁锭": {
      "ID": "minecraft:iron_ingot",
      "价格": 8
    },
    "金锭": {
      "ID": "minecraft:gold_ingot",
      "价格": 40
    },
    "绿宝石": {
      "ID": "minecraft:emerald",
      "价格": 160
    }
  }
}
```

### 配置项说明

#### 货币计分板名
指定用于存储玩家积分的计分板名称。

```json
"货币计分板名": "money"
```

**服务器设置**：
```mcfunction
# 创建货币计分板
scoreboard objectives add money dummy "积分"

# 给玩家初始积分
scoreboard players set @a money 1000
```

#### 出售配置
商店出售给玩家的物品列表。

```json
"出售": {
  "物品显示名": {
    "ID": "minecraft:物品ID",
    "价格": 购买价格
  }
}
```

#### 收购配置
商店从玩家收购的物品列表。

```json
"收购": {
  "物品显示名": {
    "ID": "minecraft:物品ID",
    "价格": 收购价格
  }
}
```

### 添加新物品

```json
{
  "出售": {
    "下界之星": {
      "ID": "minecraft:nether_star",
      "价格": 5000
    }
  },
  "收购": {
    "下界之星": {
      "ID": "minecraft:nether_star",
      "价格": 4000
    }
  }
}
```

### 物品 ID 参考

常用物品 ID：
- 钻石：`minecraft:diamond`
- 铁锭：`minecraft:iron_ingot`
- 金锭：`minecraft:gold_ingot`
- 绿宝石：`minecraft:emerald`
- 下界之星：`minecraft:nether_star`
- 附魔金苹果：`minecraft:enchanted_golden_apple`
- 钻石剑：`minecraft:diamond_sword`
- 钻石镐：`minecraft:diamond_pickaxe`

## 技术实现

### 核心数据结构

```go
// ShopPlugin 商店插件主结构
type ShopPlugin struct {
    ctx             *sdk.Context
    config          *ShopConfig
    waitingForInput map[string]*PlayerInputState
}

// ShopConfig 商店配置
type ShopConfig struct {
    ScoreboardName string                `json:"货币计分板名"`
    SellItems      map[string]*ShopItem  `json:"出售"`
    BuybackItems   map[string]*ShopItem  `json:"收购"`
}

// ShopItem 商店物品
type ShopItem struct {
    ID    string `json:"ID"`
    Price int    `json:"价格"`
}

// PlayerInputState 玩家输入状态
type PlayerInputState struct {
    Action     string    // "buy" 或 "sell"
    ItemName   string    // 物品名称
    Item       *ShopItem // 物品信息
    WaitingFor string    // "item" 或 "amount"
    Timestamp  time.Time // 超时控制
}
```

### 关键功能模块

#### 1. 配置管理
使用 `Config().GetConfig()` 加载配置，支持默认配置。

#### 2. 聊天监听
使用 `ListenChat()` 监听玩家命令和输入。

#### 3. 状态管理
使用 `map` 存储每个玩家的交易状态，支持多玩家并发交易。

#### 4. 交易执行
- **购买**：检查余额 → 扣除积分 → 给予物品
- **出售**：检查物品 → 清除物品 → 增加积分

#### 5. 超时清理
使用定时器每 30 秒清理超过 60 秒未完成的交易。

### 使用的 API

| API 方法 | 用途 |
|---------|------|
| `Config().GetConfig()` | 加载配置文件 |
| `ListenChat()` | 监听聊天消息 |
| `GameUtils().GetScore()` | 获取玩家积分 |
| `GameUtils().GetItem()` | 获取物品数量 |
| `GameUtils().SendCommand()` | 发送游戏命令 |
| `GameUtils().SayTo()` | 向玩家发送消息 |
| `Utils().NewTimer()` | 创建定时器 |
| `Utils().Sleep()` | 延时等待 |

## 运行测试

### 1. 编译插件

```bash
cd PluginFramework/templates/shop
go build -buildmode=plugin -o main.so main.go
```

### 2. 安装插件

将 `main.so` 复制到 `Plugin/shop/` 目录。

### 3. 准备服务器

在 Minecraft 服务器中执行：

```mcfunction
# 创建货币计分板
scoreboard objectives add money dummy "积分"

# 给自己初始积分
scoreboard players set @s money 2000

# 查看余额
scoreboard players list @s
```

### 4. 测试购买

```
玩家输入: 购买
玩家输入: 钻石
玩家输入: 5
```

### 5. 测试出售

```
玩家输入: 收购
玩家输入: 铁锭
玩家输入: 64
```

## 扩展功能建议

### 1. VIP 折扣系统

```go
// 检查玩家是否为 VIP
isVIP, _ := p.ctx.GameUtils().IsOp(playerName)
if isVIP {
    totalPrice = totalPrice * 90 / 100 // 9折
}
```

### 2. 交易历史记录

```go
// 使用 TempJSON 保存交易记录
type Transaction struct {
    PlayerName string
    Action     string
    ItemName   string
    Amount     int
    Price      int
    Timestamp  time.Time
}

// 保存交易
tj := p.ctx.TempJSON(p.ctx.DataPath())
transactions := []Transaction{}
tj.LoadAndRead("transactions.json", &transactions, false, nil, 0)
transactions = append(transactions, newTransaction)
tj.LoadAndWrite("transactions.json", transactions, true, 60.0)
```

### 3. 每日限购

```go
// 记录每日购买量
type DailyLimit struct {
    Date      string
    Purchases map[string]int // playerName -> amount
}

// 检查限购
if daily.Purchases[playerName] >= maxDaily {
    p.ctx.GameUtils().SayTo(playerName, "今日购买已达上限")
    return
}
```

### 4. 动态价格

```go
// 根据供需调整价格
type PriceData struct {
    BasePrice    int
    SellCount    int
    CurrentPrice int
}

// 价格计算
currentPrice = basePrice + (sellCount / 100) * 10
```

## 常见问题

### Q: 计分板不存在怎么办？
A: 在服务器中执行 `/scoreboard objectives add money dummy "积分"` 创建计分板。

### Q: 如何修改物品价格？
A: 编辑 `plugins/shop/config.json`，修改对应物品的 `价格` 字段。

### Q: 如何添加新物品？
A: 在配置文件的 `出售` 和 `收购` 中添加新的物品条目。

### Q: 交易超时时间能修改吗？
A: 在代码中搜索 `60*time.Second`，修改为你需要的时间。

### Q: 如何设置不同的购买和收购价格？
A: 在配置中，`出售` 价格通常高于 `收购` 价格，形成差价。

## 与 ToolDelta 对比

| 特性 | ToolDelta (Python) | FunInterwork (Go) |
|------|-------------------|-------------------|
| 配置格式 | JSON | JSON |
| 命令触发 | "购买"/"收购" | "购买"/"buy"/"收购"/"sell" |
| 状态管理 | 字典 | Map + 结构体 |
| 超时清理 | 无 | 自动清理（60秒） |
| 并发支持 | 有限 | 完全支持 |
| 类型安全 | 动态 | 静态类型 |

## 最佳实践

1. **合理定价**：收购价应低于出售价，形成经济循环
2. **余额检查**：交易前始终验证玩家余额和物品
3. **清晰反馈**：每步操作都给玩家明确的消息提示
4. **超时处理**：自动清理长时间未完成的交易状态
5. **日志记录**：记录所有交易操作便于审计

## 许可证

本示例插件遵循主仓库的许可证条款。
