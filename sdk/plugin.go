package sdk

import "fmt"

type Plugin interface {
	Init(ctx *Context) error
	Start() error
	Stop() error
}

type ConsoleCommandHandler func(args []string) error

type ConsoleCommand struct {
	Name        string
	Description string
	Handler     ConsoleCommandHandler
}

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
	PluginName        string
	BotInfoFunc       func() BotInfo
	ServerInfoFunc    func() ServerInfo
	QQInfoFunc        func() QQInfo
	InterworkInfoFunc func() InterworkInfo
	ConsoleRegistrar  func(ConsoleCommand) error
	Logger            func(format string, args ...interface{})
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
	return c.opts.ConsoleRegistrar(cmd)
}

func (c *Context) Logf(format string, args ...interface{}) {
	if c == nil || c.opts.Logger == nil {
		return
	}
	c.opts.Logger(format, args...)
}
