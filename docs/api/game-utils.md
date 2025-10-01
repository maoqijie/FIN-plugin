### GameUtils 高级游戏交互接口

`sdk.Context.GameUtils()` 提供类似 ToolDelta 的高级游戏交互功能，使用反射机制封装底层 `game_interface.GameInterface` 方法。所有方法返回错误时应检查并处理。

#### 核心方法

- **GetTarget(target string, timeout float64)** - 获取匹配目标选择器的玩家名称列表
  - `target`：目标选择器（如 `"@a"`, `"@p"`, `"PlayerName"`）
  - `timeout`：超时时间（秒），默认 5 秒
  - 返回：`([]string, error)` 玩家名称列表

- **GetPos(target string)** - 获取玩家的详细坐标信息
  - 返回：`(*Position, error)` 包含 X/Y/Z/Dimension/YRot 的结构体

- **GetPosXYZ(target string)** - 获取玩家的简单坐标值
  - 返回：`(x, y, z float32, err error)` 坐标元组

- **GetItem(target, itemName string, itemSpecialID int)** - 统计玩家背包中特定物品的数量
  - `itemName`：物品的 Minecraft ID（如 `"minecraft:diamond"`）
  - `itemSpecialID`：物品特殊 ID（默认 -1 表示忽略）
  - 返回：`(int, error)` 物品数量（基于 `clear` 命令测试模式）

- **GetScore(scbName, target string, timeout float64)** - 获取计分板中目标的分数
  - `scbName`：计分板名称
  - `target`：目标名称
  - `timeout`：超时时间（秒），默认 30 秒
  - 返回：`(int, error)` 分数值

- **IsCmdSuccess(cmd string, timeout float64)** - 检查命令是否执行成功
  - `cmd`：要执行的 Minecraft 命令
  - `timeout`：超时时间（秒），默认 30 秒
  - 返回：`(bool, error)` 命令是否成功

- **IsOp(playerName string)** - 检查玩家是否拥有管理员权限
  - 返回：`(bool, error)` 是否是 OP（通过 `tag` 命令测试）

- **TakeItemOutItemFrame(x, y, z int)** - 从展示框中取出物品
  - 返回：`error` 使用 `kill` 命令移除展示框

#### 命令发送方法

- **SendCommand(cmd string)** - 发送游戏命令（WebSocket 身份）
  - 示例：`utils.SendCommand("gamemode 1 Steve")`

- **SendCommandWithResponse(cmd string, timeout ...float64)** - 发送命令并等待响应
  - 返回：`(output interface{}, timedOut bool, err error)` 命令输出、是否超时、错误
  - 示例：
    ```go
    output, timedOut, err := utils.SendCommandWithResponse("testfor @a", 10.0)
    if !timedOut && err == nil {
        // 处理命令输出
    }
    ```

- **SendWOCommand(cmd string)** - 发送高权限控制台命令（Settings 通道）
  - 用于需要更高权限的命令
  - 示例：`utils.SendWOCommand("list")`

- **SendPacket(packetID uint32, packet interface{})** - 发送游戏网络数据包
  - 低级 API，需要了解 Minecraft 协议
  - 示例：
    ```go
    utils.SendPacket(0x09, map[string]interface{}{
        "text": "Hello",
    })
    ```

#### 消息发送方法

- **SendChat(message string)** - 让机器人在聊天栏发言
  - 示例：`utils.SendChat("大家好！")`

- **Title(message string)** - 以 actionbar 形式向所有玩家显示消息
  - 示例：`utils.Title("欢迎来到服务器")`

- **Tellraw(selector, message string)** - 使用 tellraw 命令发送 JSON 格式消息
  - 自动包装为 rawtext 格式
  - 示例：`utils.Tellraw("@a", "这是一条消息")`

- **SayTo(target, text string)** - 向指定目标发送聊天消息
  - 等同于 Tellraw，提供更友好的方法名
  - 示例：
    ```go
    utils.SayTo("@a", "欢迎所有玩家！")
    utils.SayTo("Steve", "你好，Steve！")
    ```

- **PlayerTitle(target, text string)** - 向指定玩家显示标题
  - 示例：`utils.PlayerTitle("@a", "游戏开始")`

- **PlayerSubtitle(target, text string)** - 向指定玩家显示副标题
  - 示例：`utils.PlayerSubtitle("@a", "Good Luck")`

- **PlayerActionbar(target, text string)** - 向指定玩家显示 ActionBar 消息
  - 示例：`utils.PlayerActionbar("@a", "当前血量: 20/20")`

#### 使用示例

```go
func (p *plugin) Start() error {
    utils := p.ctx.GameUtils()
    if utils == nil {
        return fmt.Errorf("GameUtils 未初始化")
    }

    // 获取玩家坐标
    pos, err := utils.GetPos("@p")
    if err != nil {
        p.ctx.Logf("获取坐标失败: %v", err)
        return err
    }
    p.ctx.Logf("玩家坐标: %.2f, %.2f, %.2f (维度 %d)",
        pos.X, pos.Y, pos.Z, pos.Dimension)

    // 统计钻石数量
    count, err := utils.GetItem("@p", "minecraft:diamond", -1)
    if err == nil {
        p.ctx.Logf("玩家拥有 %d 颗钻石", count)
    }

    // 检查命令执行结果
    success, _ := utils.IsCmdSuccess("testfor @a", 5.0)
    if success {
        p.ctx.Logf("服务器中有玩家在线")
    }

    // 发送聊天消息
    utils.SendChat("插件已启动！")

    // 显示 actionbar 标题
    utils.Title("欢迎来到服务器")

    // 使用 tellraw 发送格式化消息
    utils.Tellraw("@a", "这是一条来自插件的消息")

    // 使用便捷的消息发送方法
    utils.SayTo("@a", "欢迎来到服务器！")
    utils.PlayerTitle("@a", "游戏开始")
    utils.PlayerSubtitle("@a", "祝你好运")
    utils.PlayerActionbar("Steve", "血量: 20/20")