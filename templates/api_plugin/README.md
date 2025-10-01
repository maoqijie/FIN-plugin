# API 插件示例（前置插件）

这是一个前置插件示例，展示如何创建可供其他插件调用的 API 插件。

## 功能说明

此插件注册为 `example-api`，提供以下公共方法：

- `Greet(name string) string` - 问候方法
- `SendMessage(message string) error` - 向所有玩家发送消息
- `GetServerStatus() map[string]interface{}` - 获取服务器状态
- `CalculateDistance(x1, y1, z1, x2, y2, z2 float32) float32` - 计算坐标距离

## 使用方法

### 1. 注册 API 插件

在 `Init` 方法中调用 `RegisterPluginAPI`：

```go
func (p *ExampleAPIPlugin) Init(ctx *sdk.Context) error {
    p.ctx = ctx
    return ctx.RegisterPluginAPI("example-api", sdk.PluginAPIVersion{0, 0, 1}, p)
}
```

### 2. 在其他插件中调用

创建另一个插件来使用此 API：

```go
package main

import (
    sdk "github.com/maoqijie/FIN-plugin/sdk"
)

type MyPlugin struct {
    ctx *sdk.Context
}

func (p *MyPlugin) Init(ctx *sdk.Context) error {
    p.ctx = ctx

    // 在 Preload 时获取 API（确保 API 插件已加载）
    ctx.ListenPreload(func() {
        // 获取 API 插件实例
        api, version, err := ctx.GetPluginAPI("example-api")
        if err != nil {
            p.ctx.Logf("获取 API 失败: %v", err)
            return
        }

        p.ctx.Logf("获取到 example-api，版本: %s", version.String())

        // 类型断言获取具体插件类型
        if exampleAPI, ok := api.(*ExampleAPIPlugin); ok {
            // 调用 API 方法
            message := exampleAPI.Greet("World")
            p.ctx.Logf("API 返回: %s", message)

            // 发送消息
            exampleAPI.SendMessage("§a来自其他插件的问候！")

            // 获取服务器状态
            status := exampleAPI.GetServerStatus()
            p.ctx.Logf("服务器状态: %+v", status)
        }
    })

    return nil
}

func (p *MyPlugin) Start() error {
    return nil
}

func (p *MyPlugin) Stop() error {
    return nil
}

func NewPlugin() sdk.Plugin {
    return &MyPlugin{}
}
```

### 3. 使用版本检查

如果需要特定版本的 API，使用 `GetPluginAPIWithVersion`：

```go
// 要求 API 版本至少为 0.0.1
api, err := ctx.GetPluginAPIWithVersion("example-api", sdk.PluginAPIVersion{0, 0, 1})
if err != nil {
    return fmt.Errorf("API 版本不兼容: %w", err)
}
```

## 版本兼容性规则

遵循语义化版本（Semantic Versioning）：

- **主版本号（Major）**：不兼容的 API 修改
- **次版本号（Minor）**：向后兼容的功能新增
- **修订号（Patch）**：向后兼容的问题修正

版本检查规则：
- 主版本号必须完全相同
- 次版本号必须大于等于所需版本
- 修订号不做检查

示例：
- API 版本 `1.2.3` 兼容请求版本 `1.0.0`（✓）
- API 版本 `1.0.0` 不兼容请求版本 `1.2.0`（✗）
- API 版本 `2.0.0` 不兼容请求版本 `1.0.0`（✗）

## 最佳实践

1. **在 Init 中注册**：确保在其他插件访问前完成 API 注册
2. **在 Preload 中获取**：在 `ListenPreload` 回调中获取其他插件的 API
3. **类型断言检查**：获取 API 后使用类型断言确保类型正确
4. **版本管理**：更新 API 时遵循语义化版本规则
5. **错误处理**：妥善处理 API 不存在或版本不兼容的情况

## 注意事项

- API 插件需要在其他插件之前加载，确保依赖顺序正确
- 类型断言时需要导入 API 插件的包，或使用 interface 定义契约
- 卸载 API 插件时，依赖它的插件可能会出错，需要妥善处理
