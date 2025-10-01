package sdk

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Plugin interface {
	Init(ctx *Context) error
	Start() error
	Stop() error
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
	RegisterPreload       func(PreloadHandler) error
	RegisterActive        func(ActiveHandler) error
	RegisterPlayerJoin    func(PlayerEventHandler) error
	RegisterPlayerLeave   func(PlayerEventHandler) error
	RegisterChat          func(ChatHandler) error
	RegisterFrameExit     func(FrameExitHandler) error
	RegisterPacket        func(PacketHandler, []uint32) error
	RegisterPacketAll     func(PacketHandler) error
	CancelChatMessage     func(sender, message string) // 取消聊天消息转发到 QQ
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

func (c *Context) ListenPreload(handler PreloadHandler) error {
	if c == nil || c.opts.RegisterPreload == nil {
		return fmt.Errorf("预加载事件注册未启用")
	}
	if handler == nil {
		return fmt.Errorf("预加载事件处理器不能为空")
	}
	return c.opts.RegisterPreload(handler)
}

func (c *Context) ListenActive(handler ActiveHandler) error {
	if c == nil || c.opts.RegisterActive == nil {
		return fmt.Errorf("激活事件注册未启用")
	}
	if handler == nil {
		return fmt.Errorf("激活事件处理器不能为空")
	}
	return c.opts.RegisterActive(handler)
}

func (c *Context) ListenPlayerJoin(handler PlayerEventHandler) error {
	if c == nil || c.opts.RegisterPlayerJoin == nil {
		return fmt.Errorf("玩家加入事件注册未启用")
	}
	if handler == nil {
		return fmt.Errorf("玩家加入事件处理器不能为空")
	}
	return c.opts.RegisterPlayerJoin(handler)
}

func (c *Context) ListenPlayerLeave(handler PlayerEventHandler) error {
	if c == nil || c.opts.RegisterPlayerLeave == nil {
		return fmt.Errorf("玩家离开事件注册未启用")
	}
	if handler == nil {
		return fmt.Errorf("玩家离开事件处理器不能为空")
	}
	return c.opts.RegisterPlayerLeave(handler)
}

func (c *Context) ListenChat(handler ChatHandler) error {
	if c == nil || c.opts.RegisterChat == nil {
		return fmt.Errorf("聊天事件注册未启用")
	}
	if handler == nil {
		return fmt.Errorf("聊天事件处理器不能为空")
	}
	return c.opts.RegisterChat(handler)
}

func (c *Context) ListenFrameExit(handler FrameExitHandler) error {
	if c == nil || c.opts.RegisterFrameExit == nil {
		return fmt.Errorf("退出事件注册未启用")
	}
	if handler == nil {
		return fmt.Errorf("退出事件处理器不能为空")
	}
	return c.opts.RegisterFrameExit(handler)
}

func (c *Context) ListenPacket(handler PacketHandler, packetIDs ...uint32) error {
	if c == nil || c.opts.RegisterPacket == nil {
		return fmt.Errorf("数据包事件注册未启用")
	}
	if handler == nil {
		return fmt.Errorf("数据包事件处理器不能为空")
	}
	return c.opts.RegisterPacket(handler, packetIDs)
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

func (c *Context) ListenPacketAll(handler PacketHandler) error {
	if c == nil || c.opts.RegisterPacketAll == nil {
		return fmt.Errorf("全部数据包监听未启用")
	}
	if handler == nil {
		return fmt.Errorf("数据包事件处理器不能为空")
	}
	return c.opts.RegisterPacketAll(handler)
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
