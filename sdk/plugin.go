package sdk

import (
	"fmt"
	"strings"

	"github.com/Yeah114/FunInterwork/bot/core/minecraft/protocol/packet"
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
	Raw             *packet.PlayerList
	EntryIndex      int
}

type PlayerEventHandler func(PlayerEvent)

type ChatEvent struct {
	Sender     string
	Message    string
	TextType   byte
	Parameters []string
	Raw        *packet.Text
}

type ChatHandler func(ChatEvent)

type FrameExitEvent struct {
	Signal string
	Reason string
}

type FrameExitHandler func(FrameExitEvent)

type PacketEvent struct {
	ID  uint32
	Raw packet.Packet
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
	PluginName          string
	BotInfoFunc         func() BotInfo
	ServerInfoFunc      func() ServerInfo
	QQInfoFunc          func() QQInfo
	InterworkInfoFunc   func() InterworkInfo
	ConsoleRegistrar    func(ConsoleCommand) error
	Logger              func(format string, args ...interface{})
	RegisterPreload     func(PreloadHandler) error
	RegisterActive      func(ActiveHandler) error
	RegisterPlayerJoin  func(PlayerEventHandler) error
	RegisterPlayerLeave func(PlayerEventHandler) error
	RegisterChat        func(ChatHandler) error
	RegisterFrameExit   func(FrameExitHandler) error
	RegisterPacket      func(PacketHandler, []uint32) error
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
