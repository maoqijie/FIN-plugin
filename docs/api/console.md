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
