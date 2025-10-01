
    // 发送命令并等待响应
    output, timedOut, err := utils.SendCommandWithResponse("list", 5.0)
    if !timedOut && err == nil {
        p.ctx.Logf("命令执行成功，输出: %v", output)
    }

    return nil
}
```

#### 命令发送完整示例

```go
func (p *plugin) BroadcastMessage() error {
    utils := p.ctx.GameUtils()

    // 方式 1: 使用 SendChat（机器人发言）
    utils.SendChat("大家好！")

    // 方式 2: 使用 Tellraw（向玩家发送消息）
    utils.Tellraw("@a", "§a欢迎来到服务器！")

    // 方式 3: 使用 SayTo（更友好的 API）
    utils.SayTo("@a", "§e系统消息：服务器将在 5 分钟后重启")

    // 向特定玩家发送消息
    utils.SayTo("Steve", "§c你有一条私信")

    // 显示不同类型的标题
    utils.PlayerTitle("@a", "§6重要公告")
    utils.PlayerSubtitle("@a", "请注意查看聊天栏")
    utils.PlayerActionbar("@a", "§a在线玩家: 10")

    return nil
}

// 执行命令并处理结果
func (p *plugin) ExecuteCommand() error {
    utils := p.ctx.GameUtils()

    // 简单命令执行（不等待结果）
    err := utils.SendCommand("time set day")
    if err != nil {
        return fmt.Errorf("设置时间失败: %w", err)
    }

    // 执行命令并获取结果
    output, timedOut, err := utils.SendCommandWithResponse("testfor @a", 10.0)
    if timedOut {
        p.ctx.Logf("命令执行超时")
        return fmt.Errorf("命令超时")
    }
    if err != nil {
        return fmt.Errorf("命令执行失败: %w", err)
    }

    p.ctx.Logf("命令输出: %v", output)

    // 检查命令是否成功
    success, _ := utils.IsCmdSuccess("testfor @a", 5.0)
    if success {
        p.ctx.Logf("至少有一个玩家在线")
    }

    return nil
}
```

#### 注意事项

1. 所有方法都使用反射调用底层接口，性能开销略高于直接调用
2. 超时参数传入 0 或负数时会使用默认超时时间
3. `GetTarget` 方法目前返回空切片，具体实现需要解析 querytarget 的 JSON 响应
4. `IsOp` 通过尝试执行需要权限的命令来判断，可能不够准确
5. 所有方法在 GameInterface 未初始化时会返回错误
6. **命令发送方法**：
   - `SendCommand` - 普通 WebSocket 命令，不等待响应
   - `SendCommandWithResponse` - 等待命令响应和结果
   - `SendWOCommand` - 高权限控制台命令（Settings 通道）
   - `SendPacket` - 低级网络数据包 API
7. **消息发送最佳实践**：
   - 机器人发言：使用 `SendChat`
   - 向玩家发送消息：使用 `SayTo` 或 `Tellraw`
   - 显示标题：使用 `PlayerTitle`、`PlayerSubtitle`、`PlayerActionbar`
   - 支持目标选择器：`@a`（所有玩家）、`@p`（最近玩家）、玩家名

### Utils 实用工具方法

`sdk.Context.Utils()` 提供类似 ToolDelta 的实用工具方法，用于字符串格式化、类型转换、异步执行等常用操作。

#### 字符串与格式化

- **SimpleFormat(kw map[string]string, sub string)** - 简单的字符串格式化替换
  - `kw`：替换字典，键为占位符（不含 `{}`），值为替换内容
  - `sub`：要格式化的字符串，使用 `{key}` 作为占位符
  - 示例：
    ```go
    kw := map[string]string{"name": "玩家1", "score": "100"}
    result := utils.SimpleFormat(kw, "玩家 {name} 的分数是 {score}")
    // 返回: "玩家 玩家1 的分数是 100"
    ```

- **ToPlayerSelector(playerName string)** - 将玩家名转换为目标选择器
  - 示例：
    ```go
    selector := utils.ToPlayerSelector("玩家1")
    // 返回: "@a[name=\"玩家1\"]"
    ```

#### 类型转换

- **TryInt(input interface{})** - 尝试将输入转换为整数
  - 支持 int、uint、float、string、bool 等类型
  - 返回：`(int, bool)` 转换结果和是否成功
  - 示例：
    ```go
    if value, ok := utils.TryInt("123"); ok {
        fmt.Println("转换成功:", value)
    }
    ```

#### 列表操作

- **FillListIndex(list, reference []interface{}, defaultValue interface{})** - 用默认值填充列表
- **FillStringList(list []string, referenceLen int, defaultValue string)** - 填充字符串列表（类型安全）
- **FillIntList(list []int, referenceLen int, defaultValue int)** - 填充整数列表（类型安全）

#### 异步与并发

- **CreateResultCallback(timeout float64)** - 创建一对回调锁（getter 和 setter）
  - 用于异步操作中等待结果
  - `timeout`：超时时间（秒），0 表示永不超时
  - 返回：getter 函数和 setter 函数
  - 示例：
    ```go
    getter, setter := utils.CreateResultCallback(5.0)

    // 在协程中等待结果
    go func() {
        result, ok := getter()
        if ok {
            fmt.Println("收到结果:", result)
        } else {
            fmt.Println("超时或未设置")
        }
    }()

    // 在另一个地方设置结果
    time.Sleep(2 * time.Second)
    setter("操作完成")
    ```

- **RunAsync(fn func())** - 在新的 goroutine 中运行函数
- **RunAsyncWithResult(fn func() interface{})** - 异步运行并通过 channel 返回结果
- **Gather(fns ...func() interface{})** - 并行运行多个函数并收集结果
  - 示例：