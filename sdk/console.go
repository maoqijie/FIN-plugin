package sdk

import (
	"fmt"
	"strings"
)

// Console 提供控制台输出管理功能，类似 ToolDelta 的 fmts
type Console struct {
	pluginName string
}

// NewConsole 创建控制台输出管理器
func NewConsole(pluginName string) *Console {
	return &Console{
		pluginName: pluginName,
	}
}

// ANSI 颜色代码
const (
	ColorReset   = "\033[0m"
	ColorRed     = "\033[31m"
	ColorGreen   = "\033[32m"
	ColorYellow  = "\033[33m"
	ColorBlue    = "\033[34m"
	ColorMagenta = "\033[35m"
	ColorCyan    = "\033[36m"
	ColorWhite   = "\033[37m"
	ColorGray    = "\033[90m"

	ColorBrightRed     = "\033[91m"
	ColorBrightGreen   = "\033[92m"
	ColorBrightYellow  = "\033[93m"
	ColorBrightBlue    = "\033[94m"
	ColorBrightMagenta = "\033[95m"
	ColorBrightCyan    = "\033[96m"
	ColorBrightWhite   = "\033[97m"

	// 背景色
	BgRed     = "\033[41m"
	BgGreen   = "\033[42m"
	BgYellow  = "\033[43m"
	BgBlue    = "\033[44m"
	BgMagenta = "\033[45m"
	BgCyan    = "\033[46m"
	BgWhite   = "\033[47m"

	// 样式
	StyleBold      = "\033[1m"
	StyleDim       = "\033[2m"
	StyleItalic    = "\033[3m"
	StyleUnderline = "\033[4m"
	StyleBlink     = "\033[5m"
	StyleReverse   = "\033[7m"
	StyleHidden    = "\033[8m"
)

// PrintInf 输出普通信息（蓝色前缀）
// text: 要输出的文本
// needLog: 是否记录日志（默认 true）
//
// 示例:
//   console.PrintInf("这是普通信息", true)
//   // 输出: [插件名] 这是普通信息
func (c *Console) PrintInf(text string, needLog bool) {
	prefix := fmt.Sprintf("%s[%s]%s", ColorBrightBlue, c.pluginName, ColorReset)
	fmt.Printf("%s %s\n", prefix, text)
}

// PrintSuc 输出成功消息（绿色前缀）
// text: 要输出的文本
// needLog: 是否记录日志（默认 true）
//
// 示例:
//   console.PrintSuc("操作成功", true)
//   // 输出: [√] [插件名] 操作成功
func (c *Console) PrintSuc(text string, needLog bool) {
	prefix := fmt.Sprintf("%s[√] [%s]%s", ColorBrightGreen, c.pluginName, ColorReset)
	fmt.Printf("%s %s\n", prefix, text)
}

// PrintWar 输出警告消息（橙色/黄色前缀）
// text: 要输出的文本
// needLog: 是否记录日志（默认 true）
//
// 示例:
//   console.PrintWar("这是警告信息", true)
//   // 输出: [!] [插件名] 这是警告信息
func (c *Console) PrintWar(text string, needLog bool) {
	prefix := fmt.Sprintf("%s[!] [%s]%s", ColorBrightYellow, c.pluginName, ColorReset)
	fmt.Printf("%s %s\n", prefix, text)
}

// PrintErr 输出错误消息（红色前缀）
// text: 要输出的文本
// needLog: 是否记录日志（默认 true）
//
// 示例:
//   console.PrintErr("发生错误", true)
//   // 输出: [×] [插件名] 发生错误
func (c *Console) PrintErr(text string, needLog bool) {
	prefix := fmt.Sprintf("%s[×] [%s]%s", ColorBrightRed, c.pluginName, ColorReset)
	fmt.Printf("%s %s\n", prefix, text)
}

// PrintLoad 输出加载消息（紫色前缀）
// text: 要输出的文本
// needLog: 是否记录日志（默认 true）
//
// 示例:
//   console.PrintLoad("正在加载...", true)
//   // 输出: [...] [插件名] 正在加载...
func (c *Console) PrintLoad(text string, needLog bool) {
	prefix := fmt.Sprintf("%s[...] [%s]%s", ColorBrightMagenta, c.pluginName, ColorReset)
	fmt.Printf("%s %s\n", prefix, text)
}

// FmtInfo 格式化输出信息（用于 input 提示等场景）
// text: 提示文本
// info: 附加信息
// 返回: 格式化后的字符串
//
// 示例:
//   prompt := console.FmtInfo("请输入", "玩家名称")
//   // 返回: [插件名] 请输入 > 玩家名称
func (c *Console) FmtInfo(text, info string) string {
	prefix := fmt.Sprintf("%s[%s]%s", ColorBrightCyan, c.pluginName, ColorReset)
	return fmt.Sprintf("%s %s > %s", prefix, text, info)
}

// CleanPrint 无前缀打印（转换 Minecraft 颜色代码）
// text: 要打印的文本（可包含 § 颜色代码）
//
// 示例:
//   console.CleanPrint("§c红色文本§r 普通文本")
//   // 输出: 红色文本 普通文本（带颜色）
func (c *Console) CleanPrint(text string) {
	formatted := c.CleanFmt(text)
	fmt.Println(formatted)
}

// CleanFmt 转换 Minecraft 颜色代码为控制台颜色代码
// text: 包含 § 颜色代码的文本
// 返回: 转换为 ANSI 颜色代码的文本
//
// 支持的 Minecraft 颜色代码:
//   §0-§9, §a-§f: 16 种颜色
//   §l: 粗体, §o: 斜体, §n: 下划线, §r: 重置
//
// 示例:
//   text := console.CleanFmt("§c红色§l粗体§r普通")
//   // 返回带 ANSI 颜色代码的字符串
func (c *Console) CleanFmt(text string) string {
	// Minecraft 颜色代码映射到 ANSI 颜色
	colorMap := map[rune]string{
		'0': "\033[30m",    // 黑色
		'1': "\033[34m",    // 深蓝
		'2': "\033[32m",    // 深绿
		'3': "\033[36m",    // 深青
		'4': "\033[31m",    // 深红
		'5': "\033[35m",    // 深紫
		'6': "\033[33m",    // 金色/橙色
		'7': "\033[37m",    // 灰色
		'8': "\033[90m",    // 深灰
		'9': "\033[94m",    // 蓝色
		'a': "\033[92m",    // 绿色
		'b': "\033[96m",    // 青色
		'c': "\033[91m",    // 红色
		'd': "\033[95m",    // 粉红
		'e': "\033[93m",    // 黄色
		'f': "\033[97m",    // 白色
		'l': "\033[1m",     // 粗体
		'o': "\033[3m",     // 斜体
		'n': "\033[4m",     // 下划线
		'm': "\033[9m",     // 删除线
		'k': "\033[8m",     // 混淆（隐藏）
		'r': "\033[0m",     // 重置
	}

	var result strings.Builder
	runes := []rune(text)

	for i := 0; i < len(runes); i++ {
		if runes[i] == '§' && i+1 < len(runes) {
			// 找到颜色代码
			code := runes[i+1]
			if ansiCode, exists := colorMap[code]; exists {
				result.WriteString(ansiCode)
				i++ // 跳过颜色代码字符
				continue
			}
		}
		result.WriteRune(runes[i])
	}

	// 确保在末尾重置颜色
	if !strings.HasSuffix(result.String(), ColorReset) {
		result.WriteString(ColorReset)
	}

	return result.String()
}

// PrintWithColor 使用指定颜色打印文本
// text: 要打印的文本
// color: ANSI 颜色代码
//
// 示例:
//   console.PrintWithColor("红色文本", ColorRed)
func (c *Console) PrintWithColor(text, color string) {
	fmt.Printf("%s%s%s\n", color, text, ColorReset)
}

// PrintRainbow 彩虹色打印（每个字符不同颜色）
// text: 要打印的文本
//
// 示例:
//   console.PrintRainbow("彩虹文本")
func (c *Console) PrintRainbow(text string) {
	colors := []string{
		ColorRed,
		ColorBrightYellow,
		ColorBrightGreen,
		ColorBrightCyan,
		ColorBrightBlue,
		ColorBrightMagenta,
	}

	runes := []rune(text)
	for i, r := range runes {
		color := colors[i%len(colors)]
		fmt.Printf("%s%c", color, r)
	}
	fmt.Print(ColorReset + "\n")
}

// PrintBox 打印带边框的文本
// text: 要打印的文本
// boxChar: 边框字符（如 "═", "─", "*"）
//
// 示例:
//   console.PrintBox("重要消息", "═")
//   // 输出:
//   // ═══════════════
//   // ║ 重要消息    ║
//   // ═══════════════
func (c *Console) PrintBox(text string, boxChar string) {
	if boxChar == "" {
		boxChar = "═"
	}

	// 计算文本宽度（考虑中文字符）
	width := 0
	for _, r := range text {
		if r > 127 {
			width += 2 // 中文字符占两个宽度
		} else {
			width += 1
		}
	}

	// 边框宽度（文本宽度 + 4 个空格）
	borderWidth := width + 4

	// 上边框
	fmt.Println(strings.Repeat(boxChar, borderWidth))

	// 文本行
	fmt.Printf("║ %s ║\n", text)

	// 下边框
	fmt.Println(strings.Repeat(boxChar, borderWidth))
}

// PrintProgress 打印进度条
// current: 当前进度
// total: 总进度
// barWidth: 进度条宽度（字符数）
//
// 示例:
//   console.PrintProgress(50, 100, 30)
//   // 输出: [███████████████               ] 50/100 (50%)
func (c *Console) PrintProgress(current, total int, barWidth int) {
	if total == 0 {
		return
	}

	percent := float64(current) / float64(total)
	filled := int(percent * float64(barWidth))
	empty := barWidth - filled

	bar := strings.Repeat("█", filled) + strings.Repeat(" ", empty)
	fmt.Printf("\r[%s] %d/%d (%.0f%%)", bar, current, total, percent*100)

	if current >= total {
		fmt.Println() // 完成时换行
	}
}

// PrintTable 打印简单表格
// headers: 表头
// rows: 数据行
//
// 示例:
//   console.PrintTable(
//       []string{"名称", "等级", "分数"},
//       [][]string{
//           {"玩家A", "10", "100"},
//           {"玩家B", "8", "80"},
//       },
//   )
func (c *Console) PrintTable(headers []string, rows [][]string) {
	if len(headers) == 0 {
		return
	}

	// 计算每列最大宽度
	colWidths := make([]int, len(headers))
	for i, header := range headers {
		colWidths[i] = len(header)
	}

	for _, row := range rows {
		for i, cell := range row {
			if i < len(colWidths) && len(cell) > colWidths[i] {
				colWidths[i] = len(cell)
			}
		}
	}

	// 打印表头
	fmt.Print("┌")
	for i, width := range colWidths {
		fmt.Print(strings.Repeat("─", width+2))
		if i < len(colWidths)-1 {
			fmt.Print("┬")
		}
	}
	fmt.Println("┐")

	fmt.Print("│")
	for i, header := range headers {
		fmt.Printf(" %-*s │", colWidths[i], header)
	}
	fmt.Println()

	// 分隔线
	fmt.Print("├")
	for i, width := range colWidths {
		fmt.Print(strings.Repeat("─", width+2))
		if i < len(colWidths)-1 {
			fmt.Print("┼")
		}
	}
	fmt.Println("┤")

	// 打印数据行
	for _, row := range rows {
		fmt.Print("│")
		for i := 0; i < len(colWidths); i++ {
			cell := ""
			if i < len(row) {
				cell = row[i]
			}
			fmt.Printf(" %-*s │", colWidths[i], cell)
		}
		fmt.Println()
	}

	// 底部边框
	fmt.Print("└")
	for i, width := range colWidths {
		fmt.Print(strings.Repeat("─", width+2))
		if i < len(colWidths)-1 {
			fmt.Print("┴")
		}
	}
	fmt.Println("┘")
}

// ClearLine 清除当前行（用于动态更新）
func (c *Console) ClearLine() {
	fmt.Print("\r\033[K")
}

// MoveCursorUp 将光标上移 n 行
func (c *Console) MoveCursorUp(n int) {
	fmt.Printf("\033[%dA", n)
}

// MoveCursorDown 将光标下移 n 行
func (c *Console) MoveCursorDown(n int) {
	fmt.Printf("\033[%dB", n)
}

// HideCursor 隐藏光标
func (c *Console) HideCursor() {
	fmt.Print("\033[?25l")
}

// ShowCursor 显示光标
func (c *Console) ShowCursor() {
	fmt.Print("\033[?25h")
}

// ClearScreen 清屏
func (c *Console) ClearScreen() {
	fmt.Print("\033[2J\033[H")
}
