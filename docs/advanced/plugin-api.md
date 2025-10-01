## 前置插件（Plugin API）

前置插件是一种特殊类型的插件，允许其他插件调用和使用其 API 功能。当你希望其他插件能够访问和使用你的插件功能时，可以将插件注册为前置插件。

### 核心概念

- **API 注册**：插件通过 `RegisterPluginAPI` 注册自己为 API 插件
- **API 获取**：其他插件通过 `GetPluginAPI` 获取 API 插件实例
- **版本管理**：使用语义化版本（Major.Minor.Patch）管理 API 兼容性
- **类型安全**：通过类型断言或接口定义实现类型安全的方法调用

### 创建 API 插件（前置插件）

#### 1. 定义插件结构

```go
package main

import (
    "fmt"
    sdk "github.com/maoqijie/FIN-plugin/sdk"
)

type ExampleAPIPlugin struct {
    ctx *sdk.Context
}

// 公共 API 方法
func (p *ExampleAPIPlugin) Greet(name string) string {
    return fmt.Sprintf("Hello, %s!", name)
}

func (p *ExampleAPIPlugin) GetServerStatus() map[string]interface{} {
    serverInfo := p.ctx.ServerInfo()
    return map[string]interface{}{
        "code":        serverInfo.Code,
        "has_passcode": serverInfo.PasscodeSet,
    }
}
```

#### 2. 注册 API

在 `Init` 方法中注册插件为 API：

```go
func (p *ExampleAPIPlugin) Init(ctx *sdk.Context) error {
    p.ctx = ctx

    // 注册为 API 插件
    err := ctx.RegisterPluginAPI("example-api", sdk.PluginAPIVersion{
        Major: 0,  // 主版本号：不兼容的 API 修改
        Minor: 0,  // 次版本号：向后兼容的功能新增
        Patch: 1,  // 修订号：向后兼容的问题修正
    }, p)
    if err != nil {
        return fmt.Errorf("注册 API 失败: %w", err)
    }

    ctx.Logf("example-api 已注册，版本 0.0.1")
    return nil
}

func (p *ExampleAPIPlugin) Start() error {
    return nil
}

func (p *ExampleAPIPlugin) Stop() error {
    return nil
}

func NewPlugin() sdk.Plugin {
    return &ExampleAPIPlugin{}
}
```

### 使用 API 插件

#### 1. 获取 API

在其他插件中获取并使用 API：

```go
package main

import (
    sdk "github.com/maoqijie/FIN-plugin/sdk"
)

type ConsumerPlugin struct {
    ctx *sdk.Context
}

func (p *ConsumerPlugin) Init(ctx *sdk.Context) error {
    p.ctx = ctx

    // 在 Preload 阶段获取 API
    ctx.ListenPreload(func() {
        // 方式 1: 获取 API（不检查版本）
        api, version, err := ctx.GetPluginAPI("example-api")
        if err != nil {
            ctx.Logf("获取 API 失败: %v", err)
            return
        }

        ctx.Logf("获取到 example-api，版本: %s", version.String())

        // 方式 2: 获取 API（检查版本兼容性）
        api, err = ctx.GetPluginAPIWithVersion("example-api",
            sdk.PluginAPIVersion{0, 0, 1})
        if err != nil {
            ctx.Logf("API 版本不兼容: %v", err)
            return
        }

        // 使用类型断言调用方法（需要知道具体类型）
        // 注意：这需要导入 API 插件的包，或使用接口定义
    })

    return nil
}
```

#### 2. 类型断言调用

如果可以导入 API 插件的类型：

```go
import apiPlugin "path/to/example-api-plugin"

if exampleAPI, ok := api.(*apiPlugin.ExampleAPIPlugin); ok {
    message := exampleAPI.Greet("World")
    ctx.Logf("返回: %s", message)

    status := exampleAPI.GetServerStatus()
    ctx.Logf("服务器状态: %+v", status)
}
```

#### 3. 使用接口契约

更推荐的方式是定义接口契约：

```go
// 定义 API 接口
type ExampleAPI interface {
    Greet(name string) string
    GetServerStatus() map[string]interface{}
}

// 使用接口调用
if exampleAPI, ok := api.(ExampleAPI); ok {
    message := exampleAPI.Greet("World")
    ctx.Logf("返回: %s", message)
}
```

### 版本管理

#### 语义化版本规则

- **主版本号（Major）**：不兼容的 API 修改
- **次版本号（Minor）**：向后兼容的功能新增
- **修订号（Patch）**：向后兼容的问题修正

#### 版本兼容性检查

`GetPluginAPIWithVersion` 检查规则：
- 主版本号必须完全相同
- 次版本号必须大于等于所需版本
- 修订号不做检查

示例：
```go
// API 版本 1.2.3 兼容请求版本 1.0.0 ✓
// API 版本 1.0.0 不兼容请求版本 1.2.0 ✗
// API 版本 2.0.0 不兼容请求版本 1.0.0 ✗
```

### 完整示例

详见 `templates/` 目录：
- `templates/api_plugin/` - API 插件示例（前置插件）
- `templates/api_consumer/` - API 消费者插件示例

### Context 方法

- **RegisterPluginAPI(name, version, plugin)** - 注册当前插件为 API 插件
- **GetPluginAPI(name)** - 获取 API 插件实例和版本
- **GetPluginAPIWithVersion(name, version)** - 获取指定版本的 API 插件
- **ListPluginAPIs()** - 列出所有已注册的 API 插件

### 最佳实践

1. **在 Init 中注册**：确保在其他插件访问前完成 API 注册
2. **在 Preload 中获取**：在 `ListenPreload` 回调中获取其他插件的 API
3. **接口定义**：API 提供者应该定义清晰的接口契约
4. **版本管理**：遵循语义化版本规则，谨慎修改 API
5. **错误处理**：妥善处理 API 不存在或版本不兼容的情况
6. **依赖文档**：在 README 中明确说明依赖的 API 插件

### 注意事项

1. **加载顺序**：API 插件需要在依赖它的插件之前加载
2. **类型导入**：类型断言需要导入 API 插件的包，或使用接口
3. **卸载影响**：卸载 API 插件时，依赖它的插件可能会出错
4. **并发安全**：跨插件调用需要注意并发安全问题
5. **API 稳定性**：频繁修改 API 会破坏依赖插件的兼容性
