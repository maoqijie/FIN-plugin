# FIN 插件开发指南

本文档说明如何为 FunInterWork 主程序编写并分发插件。当前框架仍在建设阶段，本文重点描述目录规范、插件清单（Manifest）和运行约定，后续 SDK 与示例会在对应目录中逐步补全。

## 目录约定

主程序运行时会在工作目录创建单一插件目录，所有插件按名称归档：

```
Plugin/
  example/
    example.so
    plugin.yaml
    assets/
```

在本仓库中，`templates/` 提供示例骨架，`docs/` 维护文档，`sdk/` 存放对接主程序的公共接口定义。

## 插件结构

每个插件需要放在 `Plugin/<插件名>/` 目录下，例如：

```
Plugin/example/
  main.so             # 默认入口文件
  plugin.yaml         # 可选，补充元数据
  assets/
  README.md
```

### `plugin.yaml` Manifest

Manifest 用于描述插件的基本信息及运行入口，建议使用 YAML：

```yaml
name: example
displayName: 示例插件
version: 0.1.0
entry: ./bin/example.so   # 可选，缺省为 main.so
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
- `entry`：入口脚本或可执行文件。缺省时主程序会使用 `main.so`。
- `sdkVersion`：声明依赖的 SDK 版本，便于主程序做兼容检查。
- `dependencies`：插件间依赖（可选）。
- `permissions`：声明本插件需要访问的能力，便于后续统一治理。
- `config`：插件默认配置，主程序首次加载时可据此生成用户可编辑的配置文件。

当插件目录中存在 `.go` 源码时，主程序会在加载或热重载阶段自动执行 `go build -buildmode=plugin -o main.so .` 生成共享库（目前仅支持 Linux 与 macOS）；因此只需提交源码即可，标题所述的 `main.so` 会由运行实例按需编译。

## 生命周期约定

主程序将在以下阶段与插件交互（具体接口以 `sdk/` 发布为准）：

1. **Discover**：读取 `Plugin/<插件名>/` 目录，解析 `plugin.yaml`（如缺省则使用默认元数据）。
2. **Validate**：校验 `entry`、`sdkVersion`、权限等信息，确保入口存在。
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

## 上下文能力

`sdk.Context` 会在 `Init` 阶段传入插件，内部封装了多种读取函数：

- `BotInfo()`：返回机器人昵称、XUID 以及实体 ID。
- `ServerInfo()`：返回租赁服号以及是否配置口令。
- `QQInfo()`：返回当前使用的 QQ 适配器、OneBot WS 地址及 AccessToken 配置状态。
- `InterworkInfo()`：返回互通群别名与群号。每次调用都会复制一份映射，避免插件误改主进程数据。

调用 `Context.Logf` 输出日志时会自动附带插件前缀；`Context.PluginName()` 可获取当前插件名称，便于打包或埋点。

### 注册控制台命令

通过 `Context.RegisterConsoleCommand` 可向主程序注册新的控制台命令。主程序会优先匹配这些命令，再将未命中的输入转发为租赁服指令，因此既支持直接输入 `info`，也支持 `/info` 的写法。`ConsoleCommand` 支持以下字段：

- `Triggers`：触发词列表，至少提供一个（等效于旧版的 `Name`）。会忽略大小写及重复项。
- `ArgumentHint`：参数提示字符串（可选），用于帮助信息展示。
- `Usage`：命令用途说明（可选），在控制台输入 `?`、`/?` 或 `？` 时统一展示。
- `Description`：补充描述（可选），在帮助信息中按行显示。
- `Handler`：命令回调，入参为按空白分割后的参数切片；返回错误会在控制台回显。

示例：

```go
ctx.RegisterConsoleCommand(sdk.ConsoleCommand{
    Triggers:     []string{"info", "botinfo"},
    ArgumentHint: "[详细]",
    Usage:        "查看机器人与服务器运行状态",
    Description:  "输出机器人、租赁服与互通配置的实时信息",
    Handler: func(args []string) error {
        bot := ctx.BotInfo()
        fmt.Printf("机器人昵称: %s\n", bot.Name)
        if len(args) > 0 && strings.EqualFold(args[0], "详细") {
            inter := ctx.InterworkInfo()
            fmt.Printf("已关联群组: %d 个\n", len(inter.LinkedGroups))
        }
        return nil
    },
})
```

控制台输入 `?`、`/?` 或 `？` 时会自动列出已注册的插件命令、参数提示与用途说明，方便快速查看；插件卸载或热重载时命令会自动清理，无需手工撤销。模板 `templates/bot/info` 会随后补充对应示例。

## 开发流程示例

1. **拉取框架**：在 FunInterWork 主仓库执行 `git submodule update --init PluginFramework`。
2. **选择模板**：从 `templates/` 复制骨架到 `Plugin/<插件名>/`，并根据需要调整入口文件名。
3. **编写逻辑**：实现入口文件（暂建议使用 Go）。入口需实现 SDK 定义的 `Plugin` 接口，并导出工厂方法：
   ```go
   type Plugin interface {
       Init(ctx *sdk.Context) error
       Start() error
       Stop() error
   }

   func NewPlugin() Plugin {
       return &Example{}
   }
   ```
   主程序会加载编译后的 `.so` 并调用 `NewPlugin()`，随后依次执行 `Init`、`Start`；卸载或热重载时会调用 `Stop`。
4. **声明 Manifest**：填写 `plugin.yaml` 并更新默认配置。
5. **调试**：运行主程序，确认自动加载插件并输出日志。无需先手编译 `main.so`，主程序会在 Linux/macOS 环境下自动完成插件构建。可通过 `FUN_PLUGIN_DEBUG=1` 环境变量启用更详细日志（计划中）。
6. **打包发布**：将插件目录打包成 zip 或直接提交到私有 Git 仓库，供主程序拉取。

## 路线图

- [ ] 发布 `sdk` 包含上下文、事件模型、日志工具。
- [ ] 提供 `templates/` 下的 Go 脚手架（游戏侧、QQ 侧、综合插件）。
- [ ] 支持插件热重载与权限校验。
- [ ] 开发内置示例插件，演示常见场景（欢迎贡献）。

如需补充或讨论框架设计，可在本仓库创建 Issue 或 PR。欢迎社区共同完善 FunInterWork 插件生态。
