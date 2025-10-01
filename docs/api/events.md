### 事件监听

`sdk.Context` 已内置与 ToolDelta 类似的事件注册接口，所有监听方法会在插件重载时自动清理。默认情况下依赖公开仓库 `github.com/maoqijie/FIN-plugin`，若 `go mod tidy` 无法访问远端，主程序会自动写入本地 `replace` 规则作为回退：

- `ListenPreload(func())`：插件加载完成且尚未连接服务器时触发一次。
- `ListenActive(func())`：插件启动成功、互通链路就绪后触发。
- `ListenPlayerJoin(func(sdk.PlayerEvent))`：玩家加入时触发，包含昵称、XUID、UUID、平台信息。
- `ListenPlayerLeave(func(sdk.PlayerEvent))`：玩家离开时触发，若可获取到历史信息会一并返回。
- `ListenChat(func(sdk.ChatEvent))`：收到游戏聊天消息时触发，附带消息类型与原始数据包。
- `ListenFrameExit(func(sdk.FrameExitEvent))`：插件即将卸载或框架退出时触发，可用于清理资源。
- `ListenPacket(func(sdk.PacketEvent), packetIDs ...uint32)`：监听指定的 MC 数据包（不拦截传递，只读）。
- `ListenPacketAll(func(sdk.PacketEvent))`：监听所有 MC 数据包（**警告：性能开销大，不建议使用**）。

示例：

```go
func (p *ExamplePlugin) Init(ctx *sdk.Context) error {
    p.ctx = ctx
    ctx.ListenPreload(func() {
        p.ctx.Logf("插件已加载，等待与服务器建立连接")
    })
    ctx.ListenActive(func() {
        p.ctx.Logf("已与服务器建立连接")
    })
    ctx.ListenChat(func(evt sdk.ChatEvent) {
        p.ctx.Logf("%s: %s", evt.Sender, evt.Message)
    })
    ctx.ListenPlayerJoin(func(evt sdk.PlayerEvent) {
        p.ctx.Logf("玩家 %s 加入，平台=%d", evt.Name, evt.BuildPlatform)
    })

    // 监听特定数据包
    ctx.ListenPacket(func(evt sdk.PacketEvent) {
        p.ctx.Logf("收到文本数据包: %d", evt.ID)
    }, packet.IDText)

    // 监听所有数据包（不推荐，仅用于调试）
    // ctx.ListenPacketAll(func(evt sdk.PacketEvent) {
    //     p.ctx.Logf("收到数据包 ID: %d", evt.ID)
    // })

    return nil
}
```

> 示例需额外引入 `github.com/Yeah114/FunInterwork/bot/core/minecraft/protocol/packet`。

所有监听接口在插件 `Stop()`、热重载或程序退出时均会自动注销，无需手动清理。

### 数据包监听与等待

除了事件监听，SDK 还提供了 `PacketWaiter` 用于主动等待特定数据包的到达，类似 ToolDelta 的 `wait_next_packet` 方法。

#### PacketWaiter 方法

通过 `ctx.PacketWaiter()` 获取实例，提供以下方法：

- **WaitNextPacket(packetID uint32, timeout float64)** - 等待下一个指定类型的数据包
  - `packetID`：数据包 ID（如 `packet.IDText`）
  - `timeout`：超时时间（秒）
  - 返回：`(interface{}, error)` 数据包内容，超时返回错误
  - 示例：
    ```go
    waiter := p.ctx.PacketWaiter()
    packet, err := waiter.WaitNextPacket(packet.IDText, 30.0)
    if err != nil {
        // 超时或错误处理
        p.ctx.Logf("等待数据包超时: %v", err)
    } else {
        // 处理数据包
        p.ctx.Logf("收到数据包: %v", packet)
    }
    ```

- **WaitNextPacketAny(timeout float64)** - 等待任意数据包
  - `timeout`：超时时间（秒）
  - 返回：`(sdk.PacketEvent, error)` 数据包事件，超时返回错误
  - **警告：此方法性能开销较大，不建议频繁使用**
  - 示例：
    ```go
    waiter := p.ctx.PacketWaiter()
    event, err := waiter.WaitNextPacketAny(10.0)
    if err == nil {
        p.ctx.Logf("收到数据包 ID: %d", event.ID)
    }
    ```

#### 完整示例

```go
func (p *plugin) Start() error {
    waiter := p.ctx.PacketWaiter()

    // 等待玩家发送聊天消息
    go func() {
        for {
            packet, err := waiter.WaitNextPacket(packet.IDText, 60.0)
            if err != nil {
                p.ctx.Logf("等待超时: %v", err)
                continue
            }
            p.ctx.Logf("收到文本数据包: %v", packet)
        }
    }()

    return nil
}
```

#### 注意事项

1. **性能考虑**：
   - 监听特定数据包（`ListenPacket` 指定 ID）性能开销小
   - 监听所有数据包（`ListenPacketAll` 或 `WaitNextPacketAny`）性能开销大
   - 不建议在生产环境中监听所有数据包

2. **超时处理**：
   - 所有等待方法都支持超时设置
   - 超时后会返回错误，不会永久阻塞
   - 超时的等待器会自动清理，不会造成内存泄漏

3. **并发安全**：
   - `PacketWaiter` 内部使用锁保护，支持多 goroutine 并发调用
   - 每个等待器只会接收一个数据包后自动清理

4. **自动清理**：
   - 插件重载或卸载时，所有等待器会自动清理
   - 无需手动管理资源释放
