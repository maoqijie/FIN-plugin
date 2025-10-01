## PlayerManager - 玩家信息管理

`PlayerManager` 提供玩家信息查询和交互功能，类似 ToolDelta 的 PlayerInfoMaintainer。

### 获取 PlayerManager 实例

```go
func (p *plugin) Init(ctx *sdk.Context) error {
    p.ctx = ctx

    // 获取玩家管理器
    pm := ctx.PlayerManager()
    if pm == nil {
        return fmt.Errorf("PlayerManager 未初始化")
    }

    return nil
}
```

### PlayerManager 方法

#### GetAllPlayers - 获取所有在线玩家

```go
func (p *plugin) Start() error {
    pm := p.ctx.PlayerManager()

    // 获取所有在线玩家
    players := pm.GetAllPlayers()
    p.ctx.Logf("当前在线玩家数: %d", len(players))

    for _, player := range players {
        p.ctx.Logf("玩家: %s (UUID: %s)", player.Name, player.UUID)
    }

    return nil
}
```

#### GetBotInfo - 获取机器人信息

```go
// 获取机器人自己的玩家信息
bot := pm.GetBotInfo()
if bot != nil {
    p.ctx.Logf("机器人名称: %s", bot.Name)
    p.ctx.Logf("机器人 XUID: %s", bot.XUID)
}
```

#### GetPlayerByName - 根据名称查找玩家

```go
// 查找指定玩家
player := pm.GetPlayerByName("Steve")
if player != nil {
    p.ctx.Logf("找到玩家: %s", player.Name)
} else {
    p.ctx.Logf("玩家不在线")
}
```

#### GetPlayerByUUID / GetPlayerByUniqueID

```go
// 根据 UUID 查找
player := pm.GetPlayerByUUID("123e4567-e89b-12d3-a456-426614174000")

// 根据实体唯一 ID 查找
player := pm.GetPlayerByUniqueID(1234567890)
```

#### GetPlayerCount - 获取在线人数

```go
count := pm.GetPlayerCount()
p.ctx.Logf("在线玩家: %d 人", count)
```

### Player 对象属性

```go
type Player struct {
    Name            string // 玩家名称
    UUID            string // 玩家 UUID
    XUID            string // 玩家 XUID
    EntityUniqueID  int64  // 实体唯一 ID
    EntityRuntimeID uint64 // 实体运行时 ID
    Online          bool   // 是否在线
}
```

### Player 交互方法

#### Show - 发送聊天消息

```go
player := pm.GetPlayerByName("Steve")
if player != nil {
    // 发送普通消息
    player.Show("欢迎来到服务器！")

    // 发送带颜色的消息
    player.Show("§a你的分数: §e100")
}
```

#### SetTitle - 显示标题

```go
// 显示主标题和副标题
player.SetTitle("欢迎回来", "Welcome Back")

// 只显示主标题
player.SetTitle("游戏开始")
```

#### SetActionBar - 显示 ActionBar

```go
// 在 ActionBar 显示信息
player.SetActionBar("当前血量: 20/20")
```

### Player 查询方法

#### GetPos / GetPosXYZ - 获取玩家坐标

```go
// 获取详细坐标信息
pos, err := player.GetPos()
if err == nil {
    p.ctx.Logf("玩家位置: %.2f, %.2f, %.2f", pos.X, pos.Y, pos.Z)
    p.ctx.Logf("维度: %d, 旋转角度: %.2f", pos.Dimension, pos.YRot)
}

// 获取简单 XYZ 坐标
x, y, z, err := player.GetPosXYZ()
if err == nil {
    p.ctx.Logf("坐标: %.2f, %.2f, %.2f", x, y, z)
}
```

#### GetScore - 获取计分板分数

```go
// 获取玩家在指定计分板的分数
score, err := player.GetScore("money", 10.0) // 超时 10 秒
if err == nil {
    p.ctx.Logf("玩家金币: %d", score)
}
```

#### GetItemCount - 获取物品数量

```go
// 获取背包中钻石的数量
count, err := player.GetItemCount("diamond", 0)
if err == nil {
    p.ctx.Logf("玩家拥有 %d 个钻石", count)
}
```

#### IsOp - 检查管理员权限

```go
isOp, err := player.IsOp()
if err == nil && isOp {
    p.ctx.Logf("%s 是管理员", player.Name)
}
```

### Player 操作方法

#### Teleport - 传送玩家

```go
// 传送到指定坐标
player.Teleport(100, 64, 200)

// 传送到另一个玩家
player.TeleportTo("Alex")
```

#### SetGameMode - 设置游戏模式

```go
// 设置为创造模式
player.SetGameMode(1)

// 游戏模式: 0=生存, 1=创造, 2=冒险, 3=旁观
```

#### GiveItem - 给予物品

```go
// 给予 64 个钻石
player.GiveItem("diamond", 64, 0)

// 给予附魔金苹果
player.GiveItem("golden_apple", 1, 1)
```

#### ClearItem - 清除物品

```go
// 清除 10 个钻石
player.ClearItem("diamond", 10)

// 清除所有钻石
player.ClearItem("diamond", -1)

// 清除所有物品
player.ClearItem("", -1)
```

#### AddEffect - 给予药水效果

```go
// 给予速度 II 效果，持续 30 秒
player.AddEffect("speed", 30, 1, false)

// 给予再生效果，隐藏粒子
player.AddEffect("regeneration", 60, 0, true)

// 常用效果:
// - speed (速度)
// - slowness (缓慢)
// - haste (急迫)
// - strength (力量)
// - regeneration (再生)
// - resistance (抗性提升)
// - fire_resistance (抗火)
// - water_breathing (水下呼吸)
// - night_vision (夜视)
// - invisibility (隐身)
```

#### ClearEffects - 清除所有效果

```go
player.ClearEffects()
```

#### Kill - 杀死玩家

```go
player.Kill()
```

#### Kick - 踢出玩家

```go
// 踢出并显示原因
player.Kick("违反服务器规则")

// 踢出（无原因）
player.Kick()
```

### 完整示例

#### 示例 1：欢迎新玩家

```go
func (p *plugin) Init(ctx *sdk.Context) error {
    p.ctx = ctx

    // 监听玩家加入事件
    ctx.ListenPlayerJoin(func(event sdk.PlayerEvent) {
        pm := ctx.PlayerManager()
        player := pm.GetPlayerByName(event.Name)
        if player == nil {
            return
        }

        // 欢迎消息
        player.SetTitle("欢迎", event.Name)
        player.Show("§a欢迎来到服务器！")

        // 给予新手礼包
        player.GiveItem("diamond_sword", 1, 0)
        player.GiveItem("bread", 64, 0)

        // 传送到出生点
        player.Teleport(0, 100, 0)

        // 给予速度效果
        player.AddEffect("speed", 300, 0, false)
    })

    return nil
}
```

#### 示例 2：排行榜系统

```go
func (p *plugin) ShowRanking() error {
    pm := p.ctx.PlayerManager()
    players := pm.GetAllPlayers()

    // 获取所有玩家分数
    type PlayerScore struct {
        Name  string
        Score int
    }
    scores := make([]PlayerScore, 0, len(players))

    for _, player := range players {
        score, err := player.GetScore("money", 5.0)
        if err == nil {
            scores = append(scores, PlayerScore{
                Name:  player.Name,
                Score: score,
            })
        }
    }

    // 排序（这里简化处理）
    // ... 排序逻辑 ...

    // 向所有玩家显示排行榜
    for _, player := range players {
        player.Show("§6===== 金币排行榜 =====")
        for i, ps := range scores {
            player.Show(fmt.Sprintf("§e%d. %s: §a%d", i+1, ps.Name, ps.Score))
        }
    }

    return nil
}
```

#### 示例 3：管理员工具

```go
func (p *plugin) Init(ctx *sdk.Context) error {
    p.ctx = ctx

    // 注册控制台命令
    ctx.RegisterConsoleCommand(sdk.ConsoleCommand{
        Name:    "heal",
        Triggers: []string{"heal", "治疗"},
        Usage:   "治疗指定玩家",
        ArgumentHint: "<玩家名>",
        Handler: func(args []string) error {
            if len(args) < 1 {
                return fmt.Errorf("用法: heal <玩家名>")
            }

            pm := ctx.PlayerManager()
            player := pm.GetPlayerByName(args[0])
            if player == nil {
                return fmt.Errorf("玩家 %s 不在线", args[0])
            }

            // 清除负面效果
            player.ClearEffects()

            // 给予治疗效果
            player.AddEffect("regeneration", 10, 4, false)
            player.AddEffect("instant_health", 1, 4, false)

            // 提示消息
            player.Show("§a你已被治疗！")
            ctx.Logf("已治疗玩家: %s", player.Name)

            return nil
        },
    })

    return nil
}
```

#### 示例 4：自动踢出 AFK 玩家

```go
type plugin struct {
    ctx       *sdk.Context
    lastPos   map[string][3]float32
    afkTime   map[string]time.Time
}

func (p *plugin) Start() error {
    p.lastPos = make(map[string][3]float32)
    p.afkTime = make(map[string]time.Time)

    // 定期检查 AFK 玩家
    utils := p.ctx.Utils()
    timer := utils.NewTimer(60.0, func() { // 每 60 秒检查一次
        p.checkAFKPlayers()
    })
    timer.Start()

    return nil
}

func (p *plugin) checkAFKPlayers() {
    pm := p.ctx.PlayerManager()
    players := pm.GetAllPlayers()

    for _, player := range players {
        x, y, z, err := player.GetPosXYZ()
        if err != nil {
            continue
        }

        lastPos, exists := p.lastPos[player.Name]
        if !exists {
            p.lastPos[player.Name] = [3]float32{x, y, z}
            p.afkTime[player.Name] = time.Now()
            continue
        }

        // 检查位置是否移动
        if lastPos[0] == x && lastPos[1] == y && lastPos[2] == z {
            // 位置未变，检查 AFK 时间
            if time.Since(p.afkTime[player.Name]) > 10*time.Minute {
                player.Kick("长时间未活动 (AFK)")
                delete(p.lastPos, player.Name)
                delete(p.afkTime, player.Name)
            }
        } else {
            // 位置已变，重置计时
            p.lastPos[player.Name] = [3]float32{x, y, z}
            p.afkTime[player.Name] = time.Now()
        }
    }
}
```

### 注意事项

1. **PlayerManager 初始化**：需要在主程序中初始化并通过 Context 提供，类似 GameUtils
2. **玩家列表同步**：主程序需要监听玩家加入/离开事件，自动更新 PlayerManager
3. **并发安全**：PlayerManager 内部使用读写锁，支持多 goroutine 并发访问
4. **GameUtils 依赖**：Player 的大部分方法依赖 GameUtils，确保正确初始化
5. **命令执行**：所有操作方法本质上是发送游戏命令，需要机器人有相应权限

### ToolDelta 插件主体速览

若需对照 ToolDelta 生态的经典插件实现，可继续参考其生命周期说明，但 FunInterwork 插件框架以 Go 插件形式发布，事件注册和命令接口均通过 `sdk.Context` 提供。
