package sdk

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// PluginInfo 插件信息
type PluginInfo struct {
	Name        string
	DisplayName string
	Version     string
	Description string
	Author      string
}

type Plugin interface {
	Init(ctx *Context) error
	Start() error
	Stop() error
	GetInfo() PluginInfo
}

type ConsoleCommandHandler func(args []string) error

type ConsoleCommand struct {
	Name         string
	Triggers     []string
	ArgumentHint string
	Usage        string
	Description  string
	Handler      ConsoleCommandHandler
}

type PreloadHandler func()

type ActiveHandler func()

type PlayerEvent struct {
	Name            string
	XUID            string
	UUID            string
	EntityUniqueID  int64
	EntityRuntimeID uint64
	BuildPlatform   int32
	Raw             any
	EntryIndex      int
}

type PlayerEventHandler func(PlayerEvent)

type ChatEvent struct {
	Sender     string
	Message    string
	TextType   byte
	Parameters []string
	Raw        any
	Cancelled  bool // 插件可设置为 true 来取消事件传播（如取消转发到 QQ）
}

type ChatHandler func(*ChatEvent)

type FrameExitEvent struct {
	Signal string
	Reason string
}

type FrameExitHandler func(FrameExitEvent)

type PacketEvent struct {
	ID  uint32
	Raw any
}

type PacketHandler func(PacketEvent)

// Broadcast 表示插件间广播的事件
type Broadcast struct {
	Name string                 // 事件名称（如 "player.teleport", "economy.trade"）
	Data map[string]interface{} // 事件附加数据
}

// BroadcastHandler 广播事件处理器
// 返回值可以被广播发送者收集
type BroadcastHandler func(Broadcast) interface{}

type BotInfo struct {
	Name            string
	XUID            string
	EntityUniqueID  int64
	EntityRuntimeID uint64
}

type ServerInfo struct {
	Code        string
	PasscodeSet bool
}

type QQInfo struct {
	Adapter        string
	WSURL          string
	HasAccessToken bool
}

type InterworkInfo struct {
	LinkedGroups map[string]int64
}

type ContextOptions struct {
	PluginName            string
	BotInfoFunc           func() BotInfo
	ServerInfoFunc        func() ServerInfo
	QQInfoFunc            func() QQInfo
	InterworkInfoFunc     func() InterworkInfo
	GameUtilsProvider     func() *GameUtils
	PlayerManagerProvider func() *PlayerManager
	PacketWaiterProvider  func() *PacketWaiter
	APIRegistryProvider   func() *PluginAPIRegistry
	ConsoleRegistrar      func(ConsoleCommand) error
	Logger                func(format string, args ...interface{})
	RegisterPreload       func(PreloadHandler, int) error  // 添加优先级参数
	RegisterActive        func(ActiveHandler, int) error
	RegisterPlayerJoin    func(PlayerEventHandler, int) error
	RegisterPlayerLeave   func(PlayerEventHandler, int) error
	RegisterChat          func(ChatHandler, int) error
	RegisterFrameExit     func(FrameExitHandler, int) error
	RegisterPacket        func(PacketHandler, []uint32, int) error
	RegisterPacketAll     func(PacketHandler, int) error
	CancelChatMessage     func(sender, message string)          // 取消聊天消息转发到 QQ
	WaitPlayerMessage     func(playerName string, timeout time.Duration) (string, error) // 等待玩家发送消息
	RegisterBroadcast     func(name string, handler BroadcastHandler, priority int) error // 注册广播监听器
	TriggerBroadcast      func(broadcast Broadcast) []interface{} // 触发广播事件
}

type Context struct {
	opts ContextOptions
}

func NewContext(opts ContextOptions) *Context {
	return &Context{opts: opts}
}

func (c *Context) PluginName() string {
	return c.opts.PluginName
}

func (c *Context) BotInfo() BotInfo {
	if c == nil || c.opts.BotInfoFunc == nil {
		return BotInfo{}
	}
	return c.opts.BotInfoFunc()
}

func (c *Context) ServerInfo() ServerInfo {
	if c == nil || c.opts.ServerInfoFunc == nil {
		return ServerInfo{}
	}
	return c.opts.ServerInfoFunc()
}

func (c *Context) QQInfo() QQInfo {
	if c == nil || c.opts.QQInfoFunc == nil {
		return QQInfo{}
	}
	return c.opts.QQInfoFunc()
}

func (c *Context) InterworkInfo() InterworkInfo {
	if c == nil || c.opts.InterworkInfoFunc == nil {
		return InterworkInfo{}
	}
	info := c.opts.InterworkInfoFunc()
	if len(info.LinkedGroups) == 0 {
		return info
	}
	copied := make(map[string]int64, len(info.LinkedGroups))
	for k, v := range info.LinkedGroups {
		copied[k] = v
	}
	info.LinkedGroups = copied
	return info
}

func (c *Context) RegisterConsoleCommand(cmd ConsoleCommand) error {
	if c == nil || c.opts.ConsoleRegistrar == nil {
		return fmt.Errorf("控制台命令注册未启用")
	}
	if cmd.Handler == nil {
		return fmt.Errorf("命令处理器不能为空")
	}
	if len(cmd.Triggers) == 0 {
		name := strings.TrimSpace(cmd.Name)
		if name != "" {
			cmd.Triggers = []string{name}
		}
	}
	if len(cmd.Triggers) == 0 {
		return fmt.Errorf("命令触发词不能为空")
	}
	seen := make(map[string]struct{}, len(cmd.Triggers))
	normalized := make([]string, 0, len(cmd.Triggers))
	for _, trig := range cmd.Triggers {
		trimmed := strings.TrimSpace(trig)
		if trimmed == "" {
			continue
		}
		key := strings.ToLower(trimmed)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		normalized = append(normalized, trimmed)
	}
	if len(normalized) == 0 {
		return fmt.Errorf("命令触发词不能为空")
	}
	cmd.Triggers = normalized
	if strings.TrimSpace(cmd.Name) == "" {
		cmd.Name = cmd.Triggers[0]
	}
	return c.opts.ConsoleRegistrar(cmd)
}

func (c *Context) Logf(format string, args ...interface{}) {
	if c == nil || c.opts.Logger == nil {
		return
	}
	c.opts.Logger(format, args...)
}

// LogInfo 输出信息级别日志（蓝色背景）
// 用于一般性信息输出
//
// 示例:
//   ctx.LogInfo("插件配置已加载")
func (c *Context) LogInfo(format string, args ...interface{}) {
	if c == nil || c.opts.Logger == nil {
		return
	}
	// ANSI 蓝色背景
	msg := fmt.Sprintf(format, args...)
	c.opts.Logger("\x1b[44m[INFO]\x1b[0m %s", msg)
}

// LogSuccess 输出成功级别日志（绿色背景）
// 用于操作成功的提示
//
// 示例:
//   ctx.LogSuccess("玩家数据保存成功")
func (c *Context) LogSuccess(format string, args ...interface{}) {
	if c == nil || c.opts.Logger == nil {
		return
	}
	// ANSI 绿色背景
	msg := fmt.Sprintf(format, args...)
	c.opts.Logger("\x1b[42m[SUCCESS]\x1b[0m %s", msg)
}

// LogWarning 输出警告级别日志（黄色背景）
// 用于需要注意但不影响运行的问题
//
// 示例:
//   ctx.LogWarning("配置文件未找到，使用默认配置")
func (c *Context) LogWarning(format string, args ...interface{}) {
	if c == nil || c.opts.Logger == nil {
		return
	}
	// ANSI 黄色背景
	msg := fmt.Sprintf(format, args...)
	c.opts.Logger("\x1b[43m[WARNING]\x1b[0m %s", msg)
}

// LogError 输出错误级别日志（红色背景）
// 用于严重错误提示
//
// 示例:
//   ctx.LogError("数据库连接失败: %v", err)
func (c *Context) LogError(format string, args ...interface{}) {
	if c == nil || c.opts.Logger == nil {
		return
	}
	// ANSI 红色背景
	msg := fmt.Sprintf(format, args...)
	c.opts.Logger("\x1b[41m[ERROR]\x1b[0m %s", msg)
}

// ListenPreload 监听预加载事件（默认优先级 0）
func (c *Context) ListenPreload(handler PreloadHandler) error {
	return c.ListenPreloadWithPriority(handler, 0)
}

// ListenPreloadWithPriority 监听预加载事件（指定优先级）
// priority: 优先级，数值越大越先执行
func (c *Context) ListenPreloadWithPriority(handler PreloadHandler, priority int) error {
	if c == nil || c.opts.RegisterPreload == nil {
		return fmt.Errorf("预加载事件注册未启用")
	}
	if handler == nil {
		return fmt.Errorf("预加载事件处理器不能为空")
	}
	return c.opts.RegisterPreload(handler, priority)
}

// ListenActive 监听激活事件（默认优先级 0）
func (c *Context) ListenActive(handler ActiveHandler) error {
	return c.ListenActiveWithPriority(handler, 0)
}

// ListenActiveWithPriority 监听激活事件（指定优先级）
func (c *Context) ListenActiveWithPriority(handler ActiveHandler, priority int) error {
	if c == nil || c.opts.RegisterActive == nil {
		return fmt.Errorf("激活事件注册未启用")
	}
	if handler == nil {
		return fmt.Errorf("激活事件处理器不能为空")
	}
	return c.opts.RegisterActive(handler, priority)
}

// ListenPlayerJoin 监听玩家加入事件（默认优先级 0）
func (c *Context) ListenPlayerJoin(handler PlayerEventHandler) error {
	return c.ListenPlayerJoinWithPriority(handler, 0)
}

// ListenPlayerJoinWithPriority 监听玩家加入事件（指定优先级）
func (c *Context) ListenPlayerJoinWithPriority(handler PlayerEventHandler, priority int) error {
	if c == nil || c.opts.RegisterPlayerJoin == nil {
		return fmt.Errorf("玩家加入事件注册未启用")
	}
	if handler == nil {
		return fmt.Errorf("玩家加入事件处理器不能为空")
	}
	return c.opts.RegisterPlayerJoin(handler, priority)
}

// ListenPlayerLeave 监听玩家离开事件（默认优先级 0）
func (c *Context) ListenPlayerLeave(handler PlayerEventHandler) error {
	return c.ListenPlayerLeaveWithPriority(handler, 0)
}

// ListenPlayerLeaveWithPriority 监听玩家离开事件（指定优先级）
func (c *Context) ListenPlayerLeaveWithPriority(handler PlayerEventHandler, priority int) error {
	if c == nil || c.opts.RegisterPlayerLeave == nil {
		return fmt.Errorf("玩家离开事件注册未启用")
	}
	if handler == nil {
		return fmt.Errorf("玩家离开事件处理器不能为空")
	}
	return c.opts.RegisterPlayerLeave(handler, priority)
}

// ListenChat 监听聊天事件（默认优先级 0）
func (c *Context) ListenChat(handler ChatHandler) error {
	return c.ListenChatWithPriority(handler, 0)
}

// ListenChatWithPriority 监听聊天事件（指定优先级）
// priority: 优先级，数值越大越先执行，可用于拦截消息
func (c *Context) ListenChatWithPriority(handler ChatHandler, priority int) error {
	if c == nil || c.opts.RegisterChat == nil {
		return fmt.Errorf("聊天事件注册未启用")
	}
	if handler == nil {
		return fmt.Errorf("聊天事件处理器不能为空")
	}
	return c.opts.RegisterChat(handler, priority)
}

// ListenFrameExit 监听框架退出事件（默认优先级 0）
func (c *Context) ListenFrameExit(handler FrameExitHandler) error {
	return c.ListenFrameExitWithPriority(handler, 0)
}

// ListenFrameExitWithPriority 监听框架退出事件（指定优先级）
func (c *Context) ListenFrameExitWithPriority(handler FrameExitHandler, priority int) error {
	if c == nil || c.opts.RegisterFrameExit == nil {
		return fmt.Errorf("退出事件注册未启用")
	}
	if handler == nil {
		return fmt.Errorf("退出事件处理器不能为空")
	}
	return c.opts.RegisterFrameExit(handler, priority)
}

// ListenPacket 监听数据包事件（默认优先级 0）
func (c *Context) ListenPacket(handler PacketHandler, packetIDs ...uint32) error {
	return c.ListenPacketWithPriority(handler, 0, packetIDs...)
}

// ListenPacketWithPriority 监听数据包事件（指定优先级）
func (c *Context) ListenPacketWithPriority(handler PacketHandler, priority int, packetIDs ...uint32) error {
	if c == nil || c.opts.RegisterPacket == nil {
		return fmt.Errorf("数据包事件注册未启用")
	}
	if handler == nil {
		return fmt.Errorf("数据包事件处理器不能为空")
	}
	return c.opts.RegisterPacket(handler, packetIDs, priority)
}

func (c *Context) GameUtils() *GameUtils {
	if c == nil || c.opts.GameUtilsProvider == nil {
		return nil
	}
	return c.opts.GameUtilsProvider()
}

func (c *Context) Utils() *Utils {
	return NewUtils()
}

func (c *Context) Translator() *Translator {
	return NewTranslator()
}

func (c *Context) Console() *Console {
	pluginName := c.PluginName()
	if pluginName == "" {
		pluginName = "Plugin"
	}
	return NewConsole(pluginName)
}

func (c *Context) Config(configDir ...string) *Config {
	pluginName := c.PluginName()
	if pluginName == "" {
		pluginName = "Plugin"
	}
	return NewConfig(pluginName, configDir...)
}

func (c *Context) TempJSON(defaultDir ...string) *TempJSON {
	return NewTempJSON(defaultDir...)
}

func (c *Context) PlayerManager() *PlayerManager {
	if c == nil || c.opts.PlayerManagerProvider == nil {
		return nil
	}
	return c.opts.PlayerManagerProvider()
}

func (c *Context) PacketWaiter() *PacketWaiter {
	if c == nil || c.opts.PacketWaiterProvider == nil {
		return nil
	}
	return c.opts.PacketWaiterProvider()
}

// ListenPacketAll 监听所有数据包事件（默认优先级 0）
func (c *Context) ListenPacketAll(handler PacketHandler) error {
	return c.ListenPacketAllWithPriority(handler, 0)
}

// ListenPacketAllWithPriority 监听所有数据包事件（指定优先级）
func (c *Context) ListenPacketAllWithPriority(handler PacketHandler, priority int) error {
	if c == nil || c.opts.RegisterPacketAll == nil {
		return fmt.Errorf("全部数据包监听未启用")
	}
	if handler == nil {
		return fmt.Errorf("数据包事件处理器不能为空")
	}
	return c.opts.RegisterPacketAll(handler, priority)
}

// GetPluginAPI 获取其他插件的 API
// name: API 名称
// 返回: 插件实例、API 版本、错误
//
// 示例:
//   api, version, err := ctx.GetPluginAPI("example-api")
//   if err != nil {
//       return err
//   }
//   // 类型断言获取具体插件类型
//   if examplePlugin, ok := api.(*ExamplePlugin); ok {
//       examplePlugin.SomeMethod()
//   }
func (c *Context) GetPluginAPI(name string) (Plugin, PluginAPIVersion, error) {
	if c == nil || c.opts.APIRegistryProvider == nil {
		return nil, PluginAPIVersion{}, fmt.Errorf("插件 API 注册表未启用")
	}
	registry := c.opts.APIRegistryProvider()
	if registry == nil {
		return nil, PluginAPIVersion{}, fmt.Errorf("插件 API 注册表未初始化")
	}
	return registry.Get(name)
}

// GetPluginAPIWithVersion 获取指定版本的插件 API
// name: API 名称
// version: 所需版本（主版本号必须相同，次版本号必须大于等于）
// 返回: 插件实例、错误
//
// 示例:
//   api, err := ctx.GetPluginAPIWithVersion("example-api", sdk.PluginAPIVersion{0, 0, 1})
//   if err != nil {
//       return err
//   }
func (c *Context) GetPluginAPIWithVersion(name string, version PluginAPIVersion) (Plugin, error) {
	if c == nil || c.opts.APIRegistryProvider == nil {
		return nil, fmt.Errorf("插件 API 注册表未启用")
	}
	registry := c.opts.APIRegistryProvider()
	if registry == nil {
		return nil, fmt.Errorf("插件 API 注册表未初始化")
	}
	return registry.GetWithVersion(name, version)
}

// RegisterPluginAPI 注册当前插件为 API 插件（前置插件）
// name: API 名称
// version: API 版本
// plugin: 插件实例（通常是 self）
// 返回: 错误
//
// 注意: 应在 Init 方法中调用，确保在其他插件访问前完成注册
//
// 示例:
//   func (p *ExamplePlugin) Init(ctx *sdk.Context) error {
//       p.ctx = ctx
//       return ctx.RegisterPluginAPI("example-api", sdk.PluginAPIVersion{0, 0, 1}, p)
//   }
func (c *Context) RegisterPluginAPI(name string, version PluginAPIVersion, plugin Plugin) error {
	if c == nil || c.opts.APIRegistryProvider == nil {
		return fmt.Errorf("插件 API 注册表未启用")
	}
	registry := c.opts.APIRegistryProvider()
	if registry == nil {
		return fmt.Errorf("插件 API 注册表未初始化")
	}
	return registry.Register(name, version, plugin)
}

// ListPluginAPIs 列出所有已注册的插件 API
// 返回: API 信息列表
//
// 示例:
//   apis := ctx.ListPluginAPIs()
//   for _, api := range apis {
//       ctx.Logf("API: %s, 版本: %s", api.Name, api.Version.String())
//   }
func (c *Context) ListPluginAPIs() []PluginAPIInfo {
	if c == nil || c.opts.APIRegistryProvider == nil {
		return []PluginAPIInfo{}
	}
	registry := c.opts.APIRegistryProvider()
	if registry == nil {
		return []PluginAPIInfo{}
	}
	return registry.List()
}

// DataPath 获取插件数据目录路径
// 自动创建并返回插件专属数据文件夹路径
// 默认为 "plugins/{插件名}/"
//
// 示例:
//   dataPath := ctx.DataPath()
//   // 返回: "plugins/my-plugin/"
func (c *Context) DataPath() string {
	if c == nil {
		return ""
	}
	pluginName := c.PluginName()
	if pluginName == "" {
		pluginName = "unknown"
	}

	dataPath := filepath.Join("plugins", pluginName)

	// 确保目录存在
	os.MkdirAll(dataPath, 0755)

	return dataPath
}

// FormatDataPath 格式化数据文件路径
// 方便地生成插件内部文件路径
//
// path: 相对于插件数据目录的路径片段
// 返回: 完整的文件路径
//
// 示例:
//   // 获取配置文件路径
//   configPath := ctx.FormatDataPath("config.json")
//   // 返回: "plugins/my-plugin/config.json"
//
//   // 获取子目录文件路径
//   playerDataPath := ctx.FormatDataPath("players", "steve.json")
//   // 返回: "plugins/my-plugin/players/steve.json"
func (c *Context) FormatDataPath(path ...string) string {
	if c == nil {
		return ""
	}

	// 获取插件数据目录
	dataPath := c.DataPath()

	// 拼接路径
	if len(path) == 0 {
		return dataPath
	}

	fullPath := filepath.Join(append([]string{dataPath}, path...)...)

	// 确保父目录存在
	parentDir := filepath.Dir(fullPath)
	os.MkdirAll(parentDir, 0755)

	return fullPath
}

// CancelMessage 取消聊天消息转发到 QQ 群
// 用于插件拦截特定消息，防止转发（例如商店交互消息）
//
// sender: 消息发送者
// message: 消息内容
//
// 示例:
//   func (p *ShopPlugin) onChat(event sdk.ChatEvent) {
//       if event.Message == "购买" {
//           p.ctx.CancelMessage(event.Sender, event.Message)
//       }
//   }
func (c *Context) CancelMessage(sender, message string) {
	if c == nil || c.opts.CancelChatMessage == nil {
		return
	}
	c.opts.CancelChatMessage(sender, message)
}

// WaitMessage 等待玩家发送消息
// 用于交互式插件等待玩家输入
//
// playerName: 玩家名称
// timeout: 超时时间
//
// 返回: 玩家发送的消息，如果超时返回错误
//
// 示例:
//   ctx.GameUtils().SayTo(player, "请输入你的选择：")
//   choice, err := ctx.WaitMessage(player, 30*time.Second)
//   if err != nil {
//       ctx.GameUtils().SayTo(player, "§c超时未回复")
//       return
//   }
func (c *Context) WaitMessage(playerName string, timeout time.Duration) (string, error) {
	if c == nil || c.opts.WaitPlayerMessage == nil {
		return "", fmt.Errorf("等待消息功能未启用")
	}
	return c.opts.WaitPlayerMessage(playerName, timeout)
}

// ListenBroadcast 监听指定名称的广播事件（默认优先级）
// name: 广播事件名称（如 "player.teleport"）
// handler: 事件处理器，可以返回数据给广播发送者
//
// 示例:
//   ctx.ListenBroadcast("player.teleport", func(evt sdk.Broadcast) interface{} {
//       player := evt.Data["player"].(string)
//       ctx.Logf("玩家 %s 传送了", player)
//       return nil
//   })
func (c *Context) ListenBroadcast(name string, handler BroadcastHandler) error {
	return c.ListenBroadcastWithPriority(name, handler, 0)
}

// ListenBroadcastWithPriority 监听指定名称的广播事件（指定优先级）
// name: 广播事件名称
// handler: 事件处理器
// priority: 优先级，数值越大越先执行
//
// 示例:
//   ctx.ListenBroadcastWithPriority("player.teleport", func(evt sdk.Broadcast) interface{} {
//       // 高优先级处理逻辑
//       return "processed"
//   }, 100)
func (c *Context) ListenBroadcastWithPriority(name string, handler BroadcastHandler, priority int) error {
	if c == nil || c.opts.RegisterBroadcast == nil {
		return fmt.Errorf("广播事件注册未启用")
	}
	if handler == nil {
		return fmt.Errorf("广播事件处理器不能为空")
	}
	if name == "" {
		return fmt.Errorf("广播事件名称不能为空")
	}
	return c.opts.RegisterBroadcast(name, handler, priority)
}

// Broadcast 触发广播事件，将事件发送给所有监听者
// broadcast: 广播事件对象
// 返回: 所有处理器的返回值切片
//
// 示例:
//   results := ctx.Broadcast(sdk.Broadcast{
//       Name: "player.teleport",
//       Data: map[string]interface{}{
//           "player": "Steve",
//           "from": [3]float32{100, 64, 100},
//           "to": [3]float32{200, 64, 200},
//       },
//   })
func (c *Context) Broadcast(broadcast Broadcast) []interface{} {
	if c == nil || c.opts.TriggerBroadcast == nil {
		return []interface{}{}
	}
	return c.opts.TriggerBroadcast(broadcast)
}
