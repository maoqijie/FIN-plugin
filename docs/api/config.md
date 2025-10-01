## Config - 配置文件管理

`Config` 提供插件配置文件的读取、保存、验证和版本管理功能，类似 ToolDelta 的配置管理系统。

### 获取 Config 实例

```go
func (p *plugin) Init(ctx *sdk.Context) error {
    p.ctx = ctx

    // 获取配置管理器（默认目录为 plugins/{插件名}/）
    cfg := ctx.Config()

    // 或指定自定义配置目录
    cfg := ctx.Config("custom/config/path")

    return nil
}
```

### 配置版本管理

#### ConfigVersion - 配置版本结构

```go
type ConfigVersion struct {
    Major int // 主版本号
    Minor int // 次版本号
    Patch int // 修订版本号
}

// 创建版本
version := sdk.ConfigVersion{Major: 1, Minor: 0, Patch: 0}
fmt.Println(version.String()) // "1.0.0"

// 解析版本字符串
version, err := sdk.ParseVersion("1.2.3")

// 比较版本
result := version1.Compare(version2)
// 返回 1: version1 > version2
// 返回 0: 相等
// 返回 -1: version1 < version2
```

#### GetPluginConfigAndVersion - 获取配置和版本

自动处理配置文件的读取、默认值创建和验证。

```go
func (p *plugin) Init(ctx *sdk.Context) error {
    cfg := ctx.Config()

    // 定义默认配置
    defaultConfig := map[string]interface{}{
        "enable":     true,
        "max_count":  10,
        "mode":       "normal",
        "admin_list": []string{"admin1", "admin2"},
    }

    // 定义默认版本
    defaultVersion := sdk.ConfigVersion{Major: 1, Minor: 0, Patch: 0}

    // 可选的验证函数
    validateFunc := func(config map[string]interface{}) error {
        // 自定义验证逻辑
        if maxCount, ok := config["max_count"].(float64); ok {
            if maxCount < 1 || maxCount > 100 {
                return fmt.Errorf("max_count 必须在 1-100 之间")
            }
        }
        return nil
    }

    // 获取配置和版本
    config, version, err := cfg.GetPluginConfigAndVersion(
        "config.json",
        defaultConfig,
        defaultVersion,
        validateFunc, // 可传 nil 跳过自定义验证
    )
    if err != nil {
        return fmt.Errorf("加载配置失败: %v", err)
    }

    ctx.Logf("配置版本: %s", version.String())
    ctx.Logf("enable: %v", config["enable"])

    return nil
}
```

配置文件格式（`plugins/{插件名}/config.json`）：
```json
{
  "config": {
    "enable": true,
    "max_count": 10,
    "mode": "normal",
    "admin_list": ["admin1", "admin2"]
  },
  "version": "1.0.0"
}
```

#### UpgradePluginConfig - 升级配置

```go
// 读取旧配置
config, version, _ := cfg.GetPluginConfigAndVersion("config.json", defaultConfig, defaultVersion, nil)

// 检查是否需要升级
newVersion := sdk.ConfigVersion{Major: 1, Minor: 1, Patch: 0}
if version.Compare(newVersion) < 0 {
    // 添加新配置项
    config["new_feature"] = "enabled"
    config["new_option"] = 20

    // 保存升级后的配置
    err := cfg.UpgradePluginConfig("config.json", config, newVersion)
    if err != nil {
        return fmt.Errorf("升级配置失败: %v", err)
    }

    ctx.Logf("配置已升级到 %s", newVersion.String())
}
```

### 简单配置管理（无版本）

如果不需要版本管理，可以使用简化的 API。

#### GetConfig - 获取简单配置

```go
defaultConfig := map[string]interface{}{
    "feature_enabled": true,
    "timeout": 30,
}

config, err := cfg.GetConfig("settings.json", defaultConfig)
if err != nil {
    return err
}

// 配置文件格式为纯 JSON（无版本字段）
// {
//   "feature_enabled": true,
//   "timeout": 30
// }
```

#### SaveConfig - 保存配置

```go
config := map[string]interface{}{
    "feature_enabled": false,
    "timeout": 60,
}

err := cfg.SaveConfig("settings.json", config)
```

### 配置验证

#### CheckAuto - 自动类型验证

`CheckAuto` 提供强大的配置验证功能，支持类型检查和枚举值验证。

**支持的类型标识符：**

| 类型 | 说明 | 示例值 |
|------|------|--------|
| `"int"` | 整数 | `123`, `-456` |
| `"str"` | 字符串 | `"hello"` |
| `"bool"` | 布尔值 | `true`, `false` |
| `"float"` | 浮点数 | `3.14`, `2.0` |
| `"list"` | 列表/数组 | `[1, 2, 3]` |
| `"dict"` | 字典/对象 | `{"key": "value"}` |
| `"pint"` | 正整数（> 0） | `1`, `100` |
| `"nnint"` | 非负整数（>= 0） | `0`, `10` |

```go
// 基础类型验证
err := sdk.CheckAuto("int", 123, "max_count")
err := sdk.CheckAuto("str", "hello", "username")
err := sdk.CheckAuto("bool", true, "enable")
err := sdk.CheckAuto("pint", 8080, "port") // 正整数
err := sdk.CheckAuto("nnint", 0, "retry_count") // 非负整数

// 枚举值验证
validModes := []string{"normal", "advanced", "expert"}
err := sdk.CheckAuto(validModes, "normal", "mode") // OK
err := sdk.CheckAuto(validModes, "invalid", "mode") // Error

// 数值范围枚举
validLevels := []int{1, 2, 3, 4, 5}
err := sdk.CheckAuto(validLevels, 3, "level") // OK
```

#### ValidateConfig - 批量验证配置

使用标准模板验证整个配置对象。

```go
// 定义验证标准
standard := map[string]interface{}{
    "enable":    "bool",
    "port":      "pint",              // 正整数
    "max_count": "nnint",             // 非负整数
    "mode":      []string{"normal", "advanced", "expert"},
    "timeout":   "int",
}

// 验证配置
config := map[string]interface{}{
    "enable":    true,
    "port":      8080,
    "max_count": 100,
    "mode":      "normal",
    "timeout":   30,
}

err := sdk.ValidateConfig(config, standard)
if err != nil {
    ctx.Logf("配置验证失败: %v", err)
}
```

#### 嵌套配置验证

```go
standard := map[string]interface{}{
    "server": map[string]interface{}{
        "host": "str",
        "port": "pint",
    },
    "features": map[string]interface{}{
        "auto_restart": "bool",
        "max_retries":  "nnint",
    },
}

config := map[string]interface{}{
    "server": map[string]interface{}{
        "host": "localhost",
        "port": 8080,
    },
    "features": map[string]interface{}{
        "auto_restart": true,
        "max_retries":  3,
    },
}

err := sdk.ValidateConfig(config, standard)
```

### 完整示例

```go
type plugin struct {
    ctx    *sdk.Context
    config map[string]interface{}
}

func (p *plugin) Init(ctx *sdk.Context) error {
    p.ctx = ctx
    cfg := ctx.Config()

    // 默认配置
    defaultConfig := map[string]interface{}{
        "enable":     true,
        "mode":       "normal",
        "port":       8080,
        "max_users":  100,
        "admin_list": []string{},
    }

    // 验证标准
    standard := map[string]interface{}{
        "enable":    "bool",
        "mode":      []string{"normal", "advanced", "expert"},
        "port":      "pint",
        "max_users": "pint",
        "admin_list": "list",
    }

    // 验证函数
    validateFunc := func(config map[string]interface{}) error {
        return sdk.ValidateConfig(config, standard)
    }

    // 加载配置
    config, version, err := cfg.GetPluginConfigAndVersion(
        "config.json",
        defaultConfig,
        sdk.ConfigVersion{Major: 1, Minor: 0, Patch: 0},
        validateFunc,
    )
    if err != nil {
        return err
    }

    p.config = config
    ctx.Logf("插件配置加载完成，版本: %s", version.String())

    return nil
}

func (p *plugin) Start() error {
    // 使用配置
    if enable, ok := p.config["enable"].(bool); ok && !enable {
        p.ctx.Logf("插件已禁用")
        return nil
    }

    mode := p.config["mode"].(string)
    port := int(p.config["port"].(float64)) // JSON 数字默认为 float64

    p.ctx.Logf("启动模式: %s, 端口: %d", mode, port)

    return nil
}
```

### 其他实用方法

```go
// 获取配置文件完整路径
path := cfg.GetConfigPath("config.json")

// 检查配置文件是否存在
if cfg.ConfigExists("config.json") {
    // ...
}

// 删除配置文件
err := cfg.DeleteConfig("old_config.json")
```
