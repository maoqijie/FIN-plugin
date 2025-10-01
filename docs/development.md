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
- `GameUtils()`：返回高级游戏交互接口，提供类似 ToolDelta 的游戏操作方法（详见下文）。
- `Utils()`：返回实用工具方法，提供字符串格式化、类型转换、异步执行等常用功能（详见下文）。
- `Translator()`：返回游戏文本翻译器，将 Minecraft 文本键翻译为中文（详见下文）。
- `Console()`：返回控制台输出管理器，提供彩色输出、进度条、表格等功能（详见下文）。

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

控制台输入 `?`、`/?` 或 `？` 时会自动列出已注册的插件命令、参数提示与用途说明，方便快速查看；插件卸载或热重载时命令会自动清理，无需手工撤销。仓库中的 `templates/bot/info/` 示例演示了如何注册多个触发词并展示参数提示与用途说明。

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
5. **事件监听**：在 `Init` 阶段通过 `ctx.ListenPreload`、`ctx.ListenActive` 等方法注册回调（见下文）。
6. **调试**：运行主程序，确认自动加载插件并输出日志。无需先手编译 `main.so`，主程序会在 Linux/macOS 环境下自动完成插件构建。可通过 `FUN_PLUGIN_DEBUG=1` 环境变量启用更详细日志（计划中）。
7. **打包发布**：将插件目录打包成 zip 或直接提交到私有 Git 仓库，供主程序拉取。

### 事件监听

`sdk.Context` 已内置与 ToolDelta 类似的事件注册接口，所有监听方法会在插件重载时自动清理。默认情况下依赖公开仓库 `github.com/maoqijie/FIN-plugin`，若 `go mod tidy` 无法访问远端，主程序会自动写入本地 `replace` 规则作为回退：

- `ListenPreload(func())`：插件加载完成且尚未连接服务器时触发一次。
- `ListenActive(func())`：插件启动成功、互通链路就绪后触发。
- `ListenPlayerJoin(func(sdk.PlayerEvent))`：玩家加入时触发，包含昵称、XUID、UUID、平台信息。
- `ListenPlayerLeave(func(sdk.PlayerEvent))`：玩家离开时触发，若可获取到历史信息会一并返回。
- `ListenChat(func(sdk.ChatEvent))`：收到游戏聊天消息时触发，附带消息类型与原始数据包。
- `ListenFrameExit(func(sdk.FrameExitEvent))`：插件即将卸载或框架退出时触发，可用于清理资源。
- `ListenPacket(func(sdk.PacketEvent), packetIDs ...uint32)`：监听指定或全部 MC 数据包（不拦截传递，只读）。

示例：

```go
func (p *ExamplePlugin) Init(ctx *sdk.Context) error {
    p.ctx = ctx
    ctx.ListenPreload(func() {
        p.ctx.Logf("插件已加载，等待与服务器建立连接")
    })
    ctx.ListenActive(func() {
        p.ctx.Logf("已与服务器建立连接")
    })
    ctx.ListenChat(func(evt sdk.ChatEvent) {
        p.ctx.Logf("%s: %s", evt.Sender, evt.Message)
    })
    ctx.ListenPlayerJoin(func(evt sdk.PlayerEvent) {
        p.ctx.Logf("玩家 %s 加入，平台=%d", evt.Name, evt.BuildPlatform)
    })
    ctx.ListenPacket(func(evt sdk.PacketEvent) {
        p.ctx.Logf("收到数据包: %d", evt.ID)
    }, packet.IDText)
    return nil
}
```

> 示例需额外引入 `github.com/Yeah114/FunInterwork/bot/core/minecraft/protocol/packet`。

所有监听接口在插件 `Stop()`、热重载或程序退出时均会自动注销，无需手动清理。

### GameUtils 高级游戏交互接口

`sdk.Context.GameUtils()` 提供类似 ToolDelta 的高级游戏交互功能，使用反射机制封装底层 `game_interface.GameInterface` 方法。所有方法返回错误时应检查并处理。

#### 核心方法

- **GetTarget(target string, timeout float64)** - 获取匹配目标选择器的玩家名称列表
  - `target`：目标选择器（如 `"@a"`, `"@p"`, `"PlayerName"`）
  - `timeout`：超时时间（秒），默认 5 秒
  - 返回：`([]string, error)` 玩家名称列表

- **GetPos(target string)** - 获取玩家的详细坐标信息
  - 返回：`(*Position, error)` 包含 X/Y/Z/Dimension/YRot 的结构体

- **GetPosXYZ(target string)** - 获取玩家的简单坐标值
  - 返回：`(x, y, z float32, err error)` 坐标元组

- **GetItem(target, itemName string, itemSpecialID int)** - 统计玩家背包中特定物品的数量
  - `itemName`：物品的 Minecraft ID（如 `"minecraft:diamond"`）
  - `itemSpecialID`：物品特殊 ID（默认 -1 表示忽略）
  - 返回：`(int, error)` 物品数量（基于 `clear` 命令测试模式）

- **GetScore(scbName, target string, timeout float64)** - 获取计分板中目标的分数
  - `scbName`：计分板名称
  - `target`：目标名称
  - `timeout`：超时时间（秒），默认 30 秒
  - 返回：`(int, error)` 分数值

- **IsCmdSuccess(cmd string, timeout float64)** - 检查命令是否执行成功
  - `cmd`：要执行的 Minecraft 命令
  - `timeout`：超时时间（秒），默认 30 秒
  - 返回：`(bool, error)` 命令是否成功

- **IsOp(playerName string)** - 检查玩家是否拥有管理员权限
  - 返回：`(bool, error)` 是否是 OP（通过 `tag` 命令测试）

- **TakeItemOutItemFrame(x, y, z int)** - 从展示框中取出物品
  - 返回：`error` 使用 `kill` 命令移除展示框

#### 辅助方法

- **SendCommand(cmd string)** - 发送游戏命令（WebSocket 身份）
- **SendChat(message string)** - 让机器人在聊天栏发言
- **Title(message string)** - 以 actionbar 形式向所有玩家显示消息
- **Tellraw(selector, message string)** - 使用 tellraw 命令发送 JSON 格式消息（自动包装为 rawtext 格式）

#### 使用示例

```go
func (p *plugin) Start() error {
    utils := p.ctx.GameUtils()
    if utils == nil {
        return fmt.Errorf("GameUtils 未初始化")
    }

    // 获取玩家坐标
    pos, err := utils.GetPos("@p")
    if err != nil {
        p.ctx.Logf("获取坐标失败: %v", err)
        return err
    }
    p.ctx.Logf("玩家坐标: %.2f, %.2f, %.2f (维度 %d)",
        pos.X, pos.Y, pos.Z, pos.Dimension)

    // 统计钻石数量
    count, err := utils.GetItem("@p", "minecraft:diamond", -1)
    if err == nil {
        p.ctx.Logf("玩家拥有 %d 颗钻石", count)
    }

    // 检查命令执行结果
    success, _ := utils.IsCmdSuccess("testfor @a", 5.0)
    if success {
        p.ctx.Logf("服务器中有玩家在线")
    }

    // 发送聊天消息
    utils.SendChat("插件已启动！")

    // 显示 actionbar 标题
    utils.Title("欢迎来到服务器")

    // 使用 tellraw 发送格式化消息
    utils.Tellraw("@a", "这是一条来自插件的消息")

    return nil
}
```

#### 注意事项

1. 所有方法都使用反射调用底层接口，性能开销略高于直接调用
2. 超时参数传入 0 或负数时会使用默认超时时间
3. `GetTarget` 方法目前返回空切片，具体实现需要解析 querytarget 的 JSON 响应
4. `IsOp` 通过尝试执行需要权限的命令来判断，可能不够准确
5. 所有方法在 GameInterface 未初始化时会返回错误

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
    ```go
    results := utils.Gather(
        func() interface{} { return "结果1" },
        func() interface{} { return "结果2" },
        func() interface{} { return "结果3" },
    )
    ```

#### 定时器

- **NewTimer(interval float64, fn func())** - 创建定时器
  - `interval`：执行间隔（秒）
  - `fn`：要定时执行的函数
  - 返回：Timer 实例，支持 `Start()`、`Stop()`、`IsRunning()` 方法
  - 示例：
    ```go
    timer := utils.NewTimer(5.0, func() {
        fmt.Println("每 5 秒执行一次")
    })
    timer.Start()
    defer timer.Stop()
    ```

#### 其他工具方法

- **Sleep(seconds float64)** - 睡眠指定秒数
- **Contains(slice []string, item string)** - 检查字符串切片是否包含元素
- **ContainsInt(slice []int, item int)** - 检查整数切片是否包含元素
- **Max(a, b int)** - 返回较大值
- **Min(a, b int)** - 返回较小值
- **Clamp(value, min, max int)** - 将值限制在指定范围内

#### 完整示例

```go
func (p *plugin) Start() error {
    utils := p.ctx.Utils()

    // 字符串格式化
    msg := utils.SimpleFormat(map[string]string{
        "server": "我的服务器",
        "count": "10",
    }, "欢迎来到 {server}，当前在线 {count} 人")

    // 类型转换
    if count, ok := utils.TryInt("123"); ok {
        p.ctx.Logf("玩家数量: %d", count)
    }

    // 异步回调
    getter, setter := utils.CreateResultCallback(10.0)
    go func() {
        // 模拟异步操作
        utils.Sleep(3.0)
        setter("任务完成")
    }()

    if result, ok := getter(); ok {
        p.ctx.Logf("异步结果: %v", result)
    }

    // 定时任务
    timer := utils.NewTimer(60.0, func() {
        p.ctx.Logf("定时任务执行")
    })
    timer.Start()

    return nil
}
```

### Translator 游戏文本翻译

`sdk.Context.Translator()` 提供 Minecraft 游戏文本翻译功能，类似 ToolDelta 的 mc_translator。使用内置的中文翻译表将游戏文本键翻译为中文。

#### 核心方法

- **Translate(key string, args []interface{}, translateArgs bool)** - 翻译游戏文本
  - `key`：要翻译的消息文本键（如 `"item.diamond.name"`）
  - `args`：可选的翻译参数列表
  - `translateArgs`：是否翻译参数项
  - 示例：
    ```go
    // 简单翻译
    msg := translator.Translate("item.diamond.name", nil, false)
    // 返回: "钻石"

    // 带参数翻译
    msg := translator.Translate("death.attack.anvil", []interface{}{"SkyblueSuper"}, false)
    // 返回: "SkyblueSuper 被坠落的铁砧压扁了"

    // 翻译参数（参数以 % 开头会被翻译）
    msg := translator.Translate(
        "commands.enchant.invalidLevel",
        []interface{}{"enchantment.mending", 6},
        true,
    )
    // 返回: "经验修补 不支持等级 6"
    ```

#### 便捷方法

- **TranslateSimple(key string)** - 简单翻译，不带参数
- **TranslateWithArgs(key string, args ...interface{})** - 带参数翻译
- **TranslateItemName(itemID string)** - 翻译物品名称
- **TranslateBlockName(blockID string)** - 翻译方块名称
- **TranslateEnchantment(enchantID string)** - 翻译附魔名称

#### 管理方法

- **Has(key string)** - 检查是否存在某个翻译键
- **AddTranslation(key, value string)** - 添加自定义翻译
- **AddTranslations(translations map[string]string)** - 批量添加翻译
- **LoadFromLangFile(content string)** - 从 .lang 格式文件加载翻译

#### 颜色代码处理

- **ParseColorCodes(text string, stripCodes bool)** - 解析颜色代码
- **StripColorCodes(text string)** - 移除所有颜色代码（§ 格式）

#### 内置翻译

内置翻译表包含常用的中文翻译：
- 物品名称（钻石、绿宝石、铁锭等）
- 方块名称（泥土、石头、原木等）
- 死亡消息（被压扁、溺水、爆炸等）
- 命令消息（语法错误、玩家不存在等）
- 附魔名称（锋利、保护、效率等）
- 游戏模式（生存、创造、冒险、旁观）

#### 完整示例

```go
func (p *plugin) Start() error {
    translator := p.ctx.Translator()

    // 翻译物品名称
    itemName := translator.TranslateItemName("diamond")
    p.ctx.Logf("物品: %s", itemName) // 输出: 物品: 钻石

    // 翻译死亡消息
    deathMsg := translator.TranslateWithArgs(
        "death.attack.anvil",
        "玩家A",
    )
    p.ctx.Logf(deathMsg) // 输出: 玩家A 被坠落的铁砧压扁了

    // 添加自定义翻译
    translator.AddTranslation("custom.message", "这是自定义消息")
    msg := translator.TranslateSimple("custom.message")

    // 从文件加载翻译
    langContent := `
item.custom_item.name=自定义物品
tile.custom_block.name=自定义方块
`
    translator.LoadFromLangFile(langContent)

    // 移除颜色代码
    colorText := "§c红色文本§r 普通文本"
    plainText := translator.StripColorCodes(colorText)
    p.ctx.Logf(plainText) // 输出: 红色文本 普通文本

    return nil
}
```

#### 支持的占位符

Minecraft 使用以下格式的占位符：
- `%s` - 简单占位符（按顺序替换）
- `%1$s`, `%2$s` - 位置参数（指定顺序）

### Console 控制台输出管理

`sdk.Context.Console()` 提供控制台输出管理功能，类似 ToolDelta 的 fmts。支持彩色输出、进度条、表格等高级功能，自动附带插件名称前缀。

#### 基础输出方法

- **PrintInf(text string, needLog bool)** - 输出普通信息（蓝色前缀）
- **PrintSuc(text string, needLog bool)** - 输出成功消息（绿色 √ 前缀）
- **PrintWar(text string, needLog bool)** - 输出警告消息（黄色 ! 前缀）
- **PrintErr(text string, needLog bool)** - 输出错误消息（红色 × 前缀）
- **PrintLoad(text string, needLog bool)** - 输出加载消息（紫色 ... 前缀）

示例：
```go
console := p.ctx.Console()

console.PrintInf("这是普通信息", true)
// 输出: [插件名] 这是普通信息

console.PrintSuc("操作成功", true)
// 输出: [√] [插件名] 操作成功

console.PrintWar("这是警告", true)
// 输出: [!] [插件名] 这是警告

console.PrintErr("发生错误", true)
// 输出: [×] [插件名] 发生错误
```

#### 格式化与转换

- **FmtInfo(text, info string)** - 格式化输出信息（用于提示）
- **CleanPrint(text string)** - 无前缀打印，转换 Minecraft 颜色代码
- **CleanFmt(text string)** - 转换 Minecraft (§) 颜色代码为 ANSI 颜色

示例：
```go
// 格式化提示
prompt := console.FmtInfo("请输入", "玩家名称")
// 返回: [插件名] 请输入 > 玩家名称

// 转换颜色代码
console.CleanPrint("§c红色文本§r 普通文本")
// 输出: 红色文本 普通文本（带颜色）
```

#### 高级输出功能

- **PrintWithColor(text, color string)** - 使用指定颜色打印
- **PrintRainbow(text string)** - 彩虹色打印（每个字符不同颜色）
- **PrintBox(text, boxChar string)** - 打印带边框的文本
- **PrintProgress(current, total, barWidth int)** - 打印进度条
- **PrintTable(headers []string, rows [][]string)** - 打印表格

示例：
```go
// 彩虹文本
console.PrintRainbow("彩虹效果")

// 边框文本
console.PrintBox("重要消息", "═")
// 输出:
// ═══════════════
// ║ 重要消息    ║
// ═══════════════

// 进度条
for i := 0; i <= 100; i += 10 {
    console.PrintProgress(i, 100, 30)
    time.Sleep(100 * time.Millisecond)
}
// 输出: [███████████████               ] 50/100 (50%)

// 表格
console.PrintTable(
    []string{"名称", "等级", "分数"},
    [][]string{
        {"玩家A", "10", "100"},
        {"玩家B", "8", "80"},
    },
)
```

#### 光标控制

- **ClearLine()** - 清除当前行
- **MoveCursorUp(n int)** - 将光标上移 n 行
- **MoveCursorDown(n int)** - 将光标下移 n 行
- **HideCursor()** - 隐藏光标
- **ShowCursor()** - 显示光标
- **ClearScreen()** - 清屏

#### ANSI 颜色常量

SDK 提供了丰富的 ANSI 颜色常量：
- 基础颜色：`ColorRed`, `ColorGreen`, `ColorYellow`, `ColorBlue` 等
- 高亮颜色：`ColorBrightRed`, `ColorBrightGreen` 等
- 背景色：`BgRed`, `BgGreen` 等
- 样式：`StyleBold`, `StyleItalic`, `StyleUnderline` 等

#### Minecraft 颜色代码支持

支持的颜色代码（§ 格式）：
- `§0-§9`, `§a-§f` - 16 种颜色
- `§l` - 粗体
- `§o` - 斜体
- `§n` - 下划线
- `§m` - 删除线
- `§k` - 混淆（隐藏）
- `§r` - 重置

#### 完整示例

```go
func (p *plugin) Start() error {
    console := p.ctx.Console()

    // 基础输出
    console.PrintInf("插件启动中...", true)
    console.PrintSuc("插件启动成功！", true)

    // 进度条示例
    console.PrintInf("正在加载数据...", true)
    for i := 0; i <= 100; i += 20 {
        console.PrintProgress(i, 100, 40)
        time.Sleep(200 * time.Millisecond)
    }

    // 表格输出
    console.PrintTable(
        []string{"玩家", "在线时长", "积分"},
        [][]string{
            {"张三", "2小时", "150"},
            {"李四", "1小时", "80"},
        },
    )

    // 彩色文本
    console.CleanPrint("§a成功：§r配置已加载")
    console.CleanPrint("§c错误：§r无法连接服务器")

    // 边框消息
    console.PrintBox("欢迎使用本插件！", "═")

    return nil
}
```

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

## TempJSON - 缓存式 JSON 文件管理

`TempJSON` 提供高性能的 JSON 文件缓存管理，通过内存缓存减少磁盘 I/O，类似 ToolDelta 的 tempjson 模块。

### 获取 TempJSON 实例

```go
func (p *plugin) Init(ctx *sdk.Context) error {
    p.ctx = ctx

    // 获取 TempJSON 管理器（默认当前目录）
    tj := ctx.TempJSON()

    // 或指定默认目录
    tj := ctx.TempJSON("data")

    return nil
}
```

### 核心概念

TempJSON 通过内存缓存加速 JSON 文件操作：
- **加载（Load）**：将文件读入内存缓存
- **读写（Read/Write）**：在内存中操作数据
- **卸载（Unload）**：保存修改并从缓存中移除
- **自动卸载**：设置超时自动卸载，释放内存

### 快捷方法（推荐）

#### LoadAndRead - 快速读取

最常用的方法，适合一次性读取操作。

```go
func (p *plugin) Start() error {
    tj := p.ctx.TempJSON()

    // 快速读取（读取后立即卸载）
    data, err := tj.LoadAndRead("user_data.json", false, map[string]interface{}{
        "score": 0,
        "level": 1,
    }, 0) // timeout=0 表示立即卸载
    if err != nil {
        return err
    }

    // 类型断言
    dataMap := data.(map[string]interface{})
    score := int(dataMap["score"].(float64)) // JSON 数字默认为 float64
    level := int(dataMap["level"].(float64))

    p.ctx.Logf("用户分数: %d, 等级: %d", score, level)

    return nil
}
```

#### LoadAndWrite - 快速写入

适合一次性写入操作，写入后立即保存到磁盘。

```go
// 更新用户数据
userData := map[string]interface{}{
    "score": 1000,
    "level": 5,
    "last_login": time.Now().Format(time.RFC3339),
}

err := tj.LoadAndWrite("user_data.json", userData, false, 0)
if err != nil {
    return err
}
// 数据已自动保存到磁盘
```

### 持久缓存方法

适合需要频繁读写的场景。

#### Load - 加载到缓存

```go
tj := p.ctx.TempJSON()

// 加载文件到缓存（30 秒后自动卸载）
err := tj.Load("player_stats.json", false, map[string]interface{}{
    "players": []interface{}{},
}, 30.0)
if err != nil {
    return err
}

// 文件现在在内存中，可以多次读写
```

#### Read - 从缓存读取

```go
// 深拷贝读取（默认，安全）
data, err := tj.Read("player_stats.json", true)
if err != nil {
    return err
}

// 浅拷贝读取（性能更好，但要小心修改）
data, err := tj.Read("player_stats.json", false)
```

#### Write - 写入缓存

```go
// 修改数据
stats := map[string]interface{}{
    "players": []interface{}{
        map[string]interface{}{
            "name":  "玩家A",
            "score": 100,
        },
    },
}

// 写入缓存（不会立即保存到磁盘）
err := tj.Write("player_stats.json", stats)
```

#### Unload - 卸载缓存

```go
// 保存修改并从缓存中移除
err := tj.Unload("player_stats.json")
```

### 实际应用示例

#### 示例 1：玩家数据管理

```go
type plugin struct {
    ctx *sdk.Context
    tj  *sdk.TempJSON
}

func (p *plugin) Init(ctx *sdk.Context) error {
    p.ctx = ctx
    p.tj = ctx.TempJSON("plugin_data")
    return nil
}

// 获取玩家分数
func (p *plugin) GetPlayerScore(playerName string) (int, error) {
    // 快速读取
    data, err := p.tj.LoadAndRead("scores.json", false, map[string]interface{}{}, 0)
    if err != nil {
        return 0, err
    }

    scores := data.(map[string]interface{})
    if score, exists := scores[playerName]; exists {
        return int(score.(float64)), nil
    }

    return 0, nil
}

// 设置玩家分数
func (p *plugin) SetPlayerScore(playerName string, score int) error {
    // 读取现有数据
    data, err := p.tj.LoadAndRead("scores.json", false, map[string]interface{}{}, 30.0)
    if err != nil {
        return err
    }

    scores := data.(map[string]interface{})
    scores[playerName] = score

    // 写入并保存
    return p.tj.LoadAndWrite("scores.json", scores, false, 0)
}
```

#### 示例 2：高频读写场景

```go
func (p *plugin) Start() error {
    tj := p.ctx.TempJSON("cache")

    // 加载到缓存（60 秒后自动卸载）
    err := tj.Load("game_state.json", false, map[string]interface{}{
        "running": false,
        "players": []interface{}{},
    }, 60.0)
    if err != nil {
        return err
    }

    // 多次读写操作（在内存中进行，速度快）
    for i := 0; i < 100; i++ {
        // 读取
        data, _ := tj.Read("game_state.json", true)
        state := data.(map[string]interface{})

        // 修改
        state["tick"] = i
        players := state["players"].([]interface{})
        players = append(players, fmt.Sprintf("player_%d", i))
        state["players"] = players

        // 写入缓存
        tj.Write("game_state.json", state)
    }

    // 手动保存到磁盘
    return tj.Unload("game_state.json")
}
```

#### 示例 3：批量操作

```go
func (p *plugin) ProcessAllPlayerData() error {
    tj := p.ctx.TempJSON()

    // 加载多个文件到缓存
    files := []string{"players.json", "scores.json", "achievements.json"}
    for _, file := range files {
        err := tj.Load(file, false, map[string]interface{}{}, 30.0)
        if err != nil {
            p.ctx.Logf("加载 %s 失败: %v", file, err)
        }
    }

    // 批量处理
    for _, file := range files {
        data, err := tj.Read(file, true)
        if err != nil {
            continue
        }

        // 处理数据...
        processedData := p.processData(data)

        // 写回缓存
        tj.Write(file, processedData)
    }

    // 保存所有修改
    return tj.SaveAll()
}
```

### 高级功能

#### SaveAll - 保存所有缓存

保存所有已修改的文件到磁盘，但保持在缓存中。

```go
// 保存所有修改（不卸载）
err := tj.SaveAll()
```

#### UnloadAll - 卸载所有缓存

保存并卸载所有缓存的文件。

```go
// 插件停止时清理
func (p *plugin) Stop() error {
    return p.tj.UnloadAll()
}
```

#### GetCachedPaths - 获取缓存列表

```go
// 查看哪些文件在缓存中
paths := tj.GetCachedPaths()
for _, path := range paths {
    p.ctx.Logf("已缓存: %s", path)
}
```

#### IsCached - 检查缓存状态

```go
if tj.IsCached("data.json") {
    p.ctx.Logf("文件在缓存中")
}
```

### 性能优化建议

1. **使用快捷方法**：对于一次性操作，使用 `LoadAndRead` 和 `LoadAndWrite`
2. **合理设置超时**：频繁访问的文件设置较长超时（30-60 秒）
3. **及时卸载**：不再使用的文件及时卸载，释放内存
4. **批量操作**：多个文件需要处理时，一起加载到缓存后再批量处理
5. **深拷贝 vs 浅拷贝**：
   - 读取后要修改：使用深拷贝（`deepCopy=true`）
   - 只读取不修改：可使用浅拷贝（`deepCopy=false`）提升性能

### 注意事项

1. **JSON 数字类型**：JSON 解析后数字默认为 `float64`，需要类型转换
   ```go
   count := int(data["count"].(float64))
   ```

2. **并发安全**：TempJSON 内部已实现并发安全，可在多个 goroutine 中使用

3. **文件路径**：支持相对路径和绝对路径，相对路径基于 `defaultDir`

4. **自动卸载**：设置 `timeout > 0` 会启动定时器，超时后自动保存并卸载

5. **内存管理**：长期运行的插件应定期调用 `UnloadAll` 或设置合理的超时

### ToolDelta 插件主体速览

若需对照 ToolDelta 生态的经典插件实现，可继续参考其生命周期说明，但 FunInterwork 插件框架以 Go 插件形式发布，事件注册和命令接口均通过 `sdk.Context` 提供。

## 路线图

- [ ] 发布 `sdk` 包含上下文、事件模型、日志工具。
- [ ] 提供 `templates/` 下的 Go 脚手架（游戏侧、QQ 侧、综合插件）。
- [ ] 支持插件热重载与权限校验。
- [ ] 开发内置示例插件，演示常见场景（欢迎贡献）。

如需补充或讨论框架设计，可在本仓库创建 Issue 或 PR。欢迎社区共同完善 FunInterWork 插件生态。
