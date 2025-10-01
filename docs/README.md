# FIN 插件开发文档

欢迎使用 FunInterWork 插件框架！本文档将帮助你快速上手插件开发。

## 文档导航

### 入门指南

- [快速开始](getting-started/quickstart.md) - 5 分钟创建第一个插件
- [插件结构](getting-started/plugin-structure.md) - 插件目录、Manifest 和生命周期
- [开发流程](getting-started/development-workflow.md) - 从创建到调试的完整流程

### API 参考

#### 核心 API
- [事件监听](api/events.md) - 监听游戏、QQ 和数据包事件
- [GameUtils](api/game-utils.md) - 高级游戏交互接口
- [Context](api/context.md) - 上下文能力和控制台命令

#### 工具类 API
- [Utils](api/utils.md) - 字符串、类型转换、异步等实用工具
- [Translator](api/translator.md) - Minecraft 文本翻译
- [Console](api/console.md) - 终端彩色输出和格式化

#### 数据管理 API
- [Config](api/config.md) - 配置文件管理和版本控制
- [TempJSON](api/tempjson.md) - 高性能 JSON 缓存管理
- [PlayerManager](api/player-manager.md) - 玩家信息查询和操作

### 高级功能

- [前置插件](advanced/plugin-api.md) - 创建可被其他插件调用的 API 插件
- [数据包监听](advanced/packet-listener.md) - 底层数据包监听和等待
- [最佳实践](advanced/best-practices.md) - 性能优化和安全建议

## 快速链接

- [示例插件](../templates/) - 查看完整示例代码
- [SDK 源码](../sdk/) - 查看 SDK 实现
- [问题反馈](https://github.com/Yeah114/FunInterwork/issues) - 报告 Bug 或提建议

## 版本信息

- SDK 版本：0.1.0
- Go 版本：1.25.1+
- 支持平台：Linux、macOS（Windows 不支持 Go plugin）

## 参与贡献

欢迎提交 Issue 和 Pull Request 来完善 FunInterWork 插件生态！
