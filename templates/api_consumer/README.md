# API 消费者插件示例

这是一个使用其他插件 API 的示例插件，展示如何在插件中调用前置插件提供的功能。

## 功能说明

此插件演示如何：
- 在 `Preload` 阶段获取其他插件的 API
- 使用版本检查确保 API 兼容性
- 通过类型断言或接口调用 API 方法
- 列出所有已注册的 API 插件

## 使用方法

### 1. 确保依赖的 API 插件已加载

在使用此插件前，确保 `example-api` 插件已经加载并注册。

### 2. 在 Preload 中获取 API

```go
ctx.ListenPreload(func() {
    // 获取 API（不检查版本）
    api, version, err := ctx.GetPluginAPI("example-api")
    if err != nil {
        ctx.Logf("获取 API 失败: %v", err)
        return
    }

    ctx.Logf("获取到 API，版本: %s", version.String())
})
```

### 3. 使用版本检查

```go
// 要求 API 版本至少为 0.0.1
api, err := ctx.GetPluginAPIWithVersion("example-api", sdk.PluginAPIVersion{0, 0, 1})
if err != nil {
    return fmt.Errorf("API 版本不兼容: %w", err)
}
```

### 4. 调用 API 方法

有两种方式调用 API 方法：

#### 方式 1: 类型断言（推荐，类型安全）

```go
// 如果可以导入 API 插件的包
import apiPlugin "path/to/example-api-plugin"

if exampleAPI, ok := api.(*apiPlugin.ExampleAPIPlugin); ok {
    message := exampleAPI.Greet("World")
    ctx.Logf("返回: %s", message)
}
```

#### 方式 2: 定义接口契约

```go
// 在消费者插件中定义接口
type ExampleAPI interface {
    Greet(name string) string
    SendMessage(message string) error
    GetServerStatus() map[string]interface{}
}

// API 提供者只要实现了这些方法就可以被调用
if exampleAPI, ok := api.(ExampleAPI); ok {
    message := exampleAPI.Greet("World")
    ctx.Logf("返回: %s", message)
}
```

## 控制台命令

插件提供以下控制台命令：

- `testapi` 或 `tapi` - 测试 API 插件功能

## 完整示例

以下是一个完整的调用示例：

```go
func (p *APIConsumerPlugin) useAPI() error {
    // 获取 API
    api, err := p.ctx.GetPluginAPIWithVersion("example-api",
        sdk.PluginAPIVersion{0, 0, 1})
    if err != nil {
        return err
    }

    // 类型断言
    if exampleAPI, ok := api.(ExampleAPI); ok {
        // 调用问候方法
        greeting := exampleAPI.Greet("用户")
        p.ctx.Logf(greeting)

        // 发送消息给所有玩家
        exampleAPI.SendMessage("§a这是来自 API 的消息")

        // 获取服务器状态
        status := exampleAPI.GetServerStatus()
        for k, v := range status {
            p.ctx.Logf("%s: %v", k, v)
        }

        // 计算距离
        distance := exampleAPI.CalculateDistance(0, 0, 0, 100, 100, 100)
        p.ctx.Logf("距离: %.2f", distance)
    }

    return nil
}
```

## 插件加载顺序

重要：API 消费者插件依赖于 API 提供者插件，需要确保加载顺序：

1. 先加载 `example-api` 插件（API 提供者）
2. 再加载此插件（API 消费者）

## 最佳实践

1. **接口定义**：API 提供者应该定义清晰的接口契约
2. **错误处理**：妥善处理 API 不存在或版本不兼容的情况
3. **延迟获取**：在 `ListenPreload` 中获取 API，不要在 `Init` 中获取
4. **版本检查**：使用 `GetPluginAPIWithVersion` 确保兼容性
5. **依赖文档**：在 README 中明确说明依赖的 API 插件

## 注意事项

- 如果 API 插件未加载，`GetPluginAPI` 会返回错误
- 类型断言失败说明 API 插件类型不匹配
- API 插件卸载后，已获取的实例可能失效
- 跨插件调用需要注意并发安全问题
