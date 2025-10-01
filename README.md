# FIN Plugin Framework

FunInterWork 插件框架，支持**跨平台插件系统**（Windows/Linux/macOS/Android）和传统 .so 插件。

## 🚀 快速开始

### 创建跨平台插件（推荐）

```bash
# 在项目根目录执行
./scripts/plugin-tool.sh create my-plugin

# 编辑插件逻辑
cd plugins/my-plugin
vim main.go

# 构建所有平台
./scripts/plugin-tool.sh build my-plugin
```

### 传统 .so 插件（仅 Linux/macOS）

参考 `templates/` 目录中的示例。

## 📚 文档

- **[跨平台插件指南](../CROSS_PLATFORM_PLUGIN_GUIDE.md)** - 完整的架构与实现说明
- **[插件迁移指南](../PLUGIN_MIGRATION_GUIDE.md)** - 从 .so 迁移到跨平台
- **[插件市场文档](../PLUGIN_MARKET_README.md)** - 插件市场使用说明

## 🗂️ 目录结构

- `sdk/` - 插件 SDK 公共接口
  - `plugin.go` - 插件接口定义
  - `plugin_interface_grpc.go` - gRPC 跨平台实现
  - `game_utils.go` - 游戏控制工具
  - `player.go` - 玩家管理
  - `console.go` - 控制台输出
  - `config.go` - 配置管理
- `templates/` - 插件模板与示例
  - `cross-platform-plugin/` - **跨平台插件模板（推荐）**
  - `api_plugin/` - API 插件示例
  - `api_consumer/` - API 消费者示例
  - `shop/` - 商店插件示例
  - `data_management/` - 数据管理示例

## 🔌 插件类型对比

| 特性 | 传统插件 (.so) | 跨平台插件 (gRPC) |
|------|---------------|------------------|
| 支持平台 | Linux/macOS | Windows/Linux/macOS/Android |
| 入口方式 | `NewPlugin()` | `main()` + `plugin.Serve()` |
| 编译命令 | `go build -buildmode=plugin` | `go build` |
| 性能开销 | 无 | 启动+50-100ms，调用+0.1-0.5ms |
| 进程隔离 | ❌ | ✅ |
| 推荐度 | ⭐⭐ | ⭐⭐⭐⭐⭐ |

## 🛠️ 插件工具

使用 `scripts/plugin-tool.sh` 管理插件：

```bash
# 创建新插件
./scripts/plugin-tool.sh create <名称>

# 构建插件
./scripts/plugin-tool.sh build <名称> [平台]

# 构建所有插件
./scripts/plugin-tool.sh build-all

# 列出所有插件
./scripts/plugin-tool.sh list

# 清理构建产物
./scripts/plugin-tool.sh clean <名称>
```

## 📦 SDK 功能

### 控制台命令

```go
ctx.RegisterConsoleCommand(sdk.ConsoleCommand{
    Name: "mycmd",
    Handler: func(args []string) error {
        ctx.LogInfo("命令执行")
        return nil
    },
})
```

### 游戏事件监听

```go
// 玩家加入
ctx.ListenPlayerJoin(func(event sdk.PlayerEvent) {
    ctx.LogSuccess("玩家 %s 加入", event.Name)
})

// 聊天消息
ctx.ListenChat(func(event *sdk.ChatEvent) {
    ctx.LogInfo("%s: %s", event.Sender, event.Message)
})
```

### 游戏控制

```go
gu := ctx.GameUtils()
gu.SayTo("玩家名", "§a你好！")
gu.SendCommand("/give @a diamond 64")
```

### 数据存储

```go
config := ctx.Config()
config.SetDefault("key", "value")
value := config.GetString("key")
```

## 🌟 示例插件

- **`cross-platform-plugin/`** - 跨平台插件完整示例（推荐起点）
- **`api_plugin/`** - 提供 API 供其他插件调用
- **`api_consumer/`** - 调用其他插件的 API
- **`shop/`** - 商店系统示例
- **`data_management/`** - 数据持久化示例

## 🤝 贡献

欢迎提交 PR 补充插件示例或完善文档。
