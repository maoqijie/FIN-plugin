# FIN 插件开发指南

本文档说明如何为 FunInterWork 主程序编写并分发插件。当前框架仍在建设阶段，本文重点描述目录规范、插件清单（Manifest）和运行约定，后续 SDK 与示例会在对应目录中逐步补全。

## 目录约定

主程序运行时会在工作目录创建以下三类插件目录：

```
Plugin/
  Nbot/   # Minecraft 游戏侧插件（仅处理游戏事件）
  Qbot/   # QQ 机器人插件（仅处理 QQ 事件）
  bot/    # 综合插件（可同时操作游戏与 QQ）
```

在本仓库中，`templates/` 将提供各目录的示例骨架，`docs/` 负责文档，`sdk/` 将存放对接主程序的公共接口定义。

## 插件结构

每个插件应放置在目标分类目录下的独立文件夹，例如：

```
Plugin/Nbot/example/
  plugin.yaml
  main.go            # 或其它实现语言（待 SDK 发布）
  assets/
  README.md
```

### `plugin.yaml` Manifest

Manifest 用于描述插件的基本信息及运行入口，建议使用 YAML：

```yaml
name: example
displayName: 示例插件
version: 0.1.0
type: Nbot            # Nbot | Qbot | bot
entry: main.go        # 或 ./bin/example.so 等
sdkVersion: 0.1.0
authors:
  - 猫七街
description: |
  这是一个演示插件，展示如何监听游戏事件并广播到 QQ。
dependencies:
  - name: core
    version: ">=0.1.0"
permissions:
  - minecraft.chat.read
  - minecraft.chat.write
  - qq.group.send
config:
  enable: true
  targetGroup: 123456789
```

字段说明：

- `name`：插件唯一标识。
- `type`：必须与所在目录一致（`Nbot`、`Qbot` 或 `bot`）。
- `entry`：入口脚本或可执行文件。主程序将读取 Manifest 并根据 `type` 调用相应的装载器。
- `sdkVersion`：声明依赖的 SDK 版本，便于主程序做兼容检查。
- `dependencies`：插件间依赖（可选）。
- `permissions`：声明本插件需要访问的能力，便于后续统一治理。
- `config`：插件默认配置，主程序首次加载时可据此生成用户可编辑的配置文件。

## 生命周期约定

主程序将在以下阶段与插件交互（具体接口以 `sdk/` 发布为准）：

1. **Discover**：读取分类目录，解析 `plugin.yaml`。
2. **Validate**：校验 `type`、`sdkVersion`、权限等信息。
3. **Init**：调用插件入口的 `Init(ctx)`，传入运行环境（日志输出、事件总线、配置等）。
4. **Start**：对可运行插件执行 `Start()`，开始监听事件或任务。
5. **Stop**：进程退出或热重载时调用 `Stop()`，要求插件自行释放资源。

当插件在运行期崩溃或返回错误时，主程序会按照既定策略（重试 / 熔断）处理，并记录到统一日志。

## 事件模型概览

| 事件源     | 说明                             | 是否可写 | 示例事件 ID                 |
| ---------- | -------------------------------- | -------- | --------------------------- |
| Minecraft  | 来自游戏内的聊天、玩家状态、命令 | 读/写    | `minecraft.chat.message`    |
| QQ 机器人  | 群消息、私聊、通知               | 读/写    | `qq.group.message`          |
| 桥接层     | 主程序内部桥接状态               | 只读     | `bridge.status.reconnected` |

SDK 将提供统一事件总线，插件可订阅或发送事件；综合插件可同时订阅两侧事件，实现跨平台逻辑。

## 开发流程示例

1. **拉取框架**：在 FunInterWork 主仓库执行 `git submodule update --init PluginFramework`。
2. **选择模板**：从 `templates/` 复制对应类型的骨架至目标目录，例如 `Plugin/bot/example/`。
3. **编写逻辑**：实现入口文件（暂建议使用 Go）。入口需实现 SDK 定义的 `Plugin` 接口：
   ```go
   type Plugin interface {
       Init(ctx *sdk.Context) error
       Start() error
       Stop() error
   }
   ```
4. **声明 Manifest**：填写 `plugin.yaml` 并更新默认配置。
5. **调试**：运行主程序，确认自动加载插件并输出日志。可通过 `FUN_PLUGIN_DEBUG=1` 环境变量启用更详细日志（计划中）。
6. **打包发布**：将插件目录打包成 zip 或直接提交到私有 Git 仓库，供主程序拉取。

## 路线图

- [ ] 发布 `sdk` 包含上下文、事件模型、日志工具。
- [ ] 提供 `templates/` 下的 Go 脚手架（游戏侧、QQ 侧、综合插件）。
- [ ] 支持插件热重载与权限校验。
- [ ] 开发内置示例插件，演示常见场景（欢迎贡献）。

如需补充或讨论框架设计，可在本仓库创建 Issue 或 PR。欢迎社区共同完善 FunInterWork 插件生态。
