# FIN 插件开发指南

欢迎使用 FunInterWork 插件框架！

> **注意**：文档已重组为模块化结构，方便查阅。如需查看完整文档，请参考 [development-full.md](development-full.md)。

## 📚 快速导航

### 🚀 入门指南

从这里开始你的插件开发之旅：

- **[快速开始](getting-started/quickstart.md)** - 5 分钟创建第一个插件
- **[插件结构](getting-started/plugin-structure.md)** - 目录约定、Manifest 和生命周期
- **[开发流程](getting-started/development-workflow.md)** - 开发、调试和热重载

### 📖 API 参考

#### 核心 API

- **[Context](api/context.md)** - 上下文能力、控制台命令注册
- **[事件监听](api/events.md)** - 监听游戏、QQ、数据包事件
- **[GameUtils](api/game-utils.md)** - 高级游戏交互（命令、查询、消息）

#### 工具类

- **[Utils](api/utils.md)** - 字符串、类型转换、异步、定时器
- **[Translator](api/translator.md)** - Minecraft 文本翻译
- **[Console](api/console.md)** - 终端彩色输出、进度条、表格

#### 数据管理

- **[Config](api/config.md)** - 配置文件管理、版本控制、验证
- **[TempJSON](api/tempjson.md)** - 高性能 JSON 缓存管理
- **[PlayerManager](api/player-manager.md)** - 玩家信息查询和操作

### 🔧 高级功能

- **[前置插件（Plugin API）](advanced/plugin-api.md)** - 跨插件方法调用和依赖管理
- **[数据包监听](advanced/packet-listener.md)** - 底层协议包监听和等待
- **[最佳实践](advanced/best-practices.md)** - 性能优化、安全建议、常见问题

## 🎯 常用场景

### 我想...

- **监听玩家聊天** → [事件监听 - ChatEvent](api/events.md#chatvent)
- **向玩家发送消息** → [GameUtils - 消息发送](api/game-utils.md#消息发送方法)
- **获取玩家坐标** → [PlayerManager - GetPos](api/player-manager.md#getpos)
- **注册控制台命令** → [Context - 控制台命令](api/context.md#控制台命令)
- **保存插件配置** → [Config - 配置管理](api/config.md)
- **让其他插件调用我的方法** → [前置插件](advanced/plugin-api.md)
- **监听特定数据包** → [数据包监听](advanced/packet-listener.md)

## 📦 示例代码

完整示例位于 `templates/` 目录：

- `bot/info/` - 控制台命令示例
- `api_plugin/` - 前置插件示例
- `api_consumer/` - API 消费者示例

## 🛠️ 开发环境

### 系统要求

- **Go 版本**：1.25.1 或更高
- **操作系统**：Linux 或 macOS（Windows 不支持 Go plugin）
- **FunInterWork**：最新版本

### 目录结构

```
FunInterWork/
├── PluginFramework/     # SDK 子模块
│   ├── sdk/             # SDK 源码
│   ├── templates/       # 插件模板
│   └── docs/            # 文档（当前目录）
└── Plugin/              # 插件目录（运行时创建）
    └── <插件名>/
        ├── main.go      # 插件源码
        ├── main.so      # 编译产物（自动生成）
        ├── go.mod
        └── plugin.yaml  # 可选
```

## 📝 核心概念

### 插件接口

所有插件必须实现 `sdk.Plugin` 接口：

```go
type Plugin interface {
    Init(ctx *sdk.Context) error  // 初始化
    Start() error                 // 启动
    Stop() error                  // 停止
}
```

### 生命周期

1. **Discover** - 扫描插件目录
2. **Validate** - 校验插件合法性
3. **Init** - 调用插件 Init 方法
4. **Start** - 调用插件 Start 方法
5. **Stop** - 卸载或热重载时调用

### Context 上下文

`sdk.Context` 提供核心能力：

- `BotInfo()` - 机器人信息
- `ServerInfo()` - 租赁服信息
- `QQInfo()` - QQ 适配器信息
- `InterworkInfo()` - 互通群信息
- `GameUtils()` - 游戏交互接口
- `PlayerManager()` - 玩家管理器
- `Logf()` - 日志输出

## 🤝 参与贡献

- **报告问题**：[GitHub Issues](https://github.com/Yeah114/FunInterwork/issues)
- **提交代码**：欢迎 Pull Request
- **文档改进**：文档存放在 `PluginFramework/docs/`

## 📄 许可证

本项目遵循主仓库的许可证条款。

---

**参考 ToolDelta**：本插件框架参考了 ToolDelta 的 API 设计，提供类似的开发体验。

**完整文档**：[development-full.md](development-full.md) 包含所有 API 的详细说明和示例。
