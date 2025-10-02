package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// ContextGRPCProxy 是 Context 的 gRPC 代理实现
// 运行在插件进程中，将所有方法调用转发到主进程的 ContextServer
type ContextGRPCProxy struct {
	pluginName    string
	client        ContextServiceClient
	callbackServer *CallbackServerImpl

	nextCallbackID uint32
	callbacksMu    sync.Mutex
}

func NewContextGRPCProxy(pluginName string, client ContextServiceClient, callbackServer *CallbackServerImpl) *ContextGRPCProxy {
	return &ContextGRPCProxy{
		pluginName:     pluginName,
		client:         client,
		callbackServer: callbackServer,
		nextCallbackID: 1,
	}
}

// ToContext 将 ContextGRPCProxy 转换为 Context
// 通过创建 ContextOptions 并委托所有调用给 proxy
func (c *ContextGRPCProxy) ToContext() *Context {
	opts := ContextOptions{
		PluginName: c.pluginName,
		BotInfoFunc: func() BotInfo {
			return c.BotInfo()
		},
		ServerInfoFunc: func() ServerInfo {
			return c.ServerInfo()
		},
		QQInfoFunc: func() QQInfo {
			return c.QQInfo()
		},
		InterworkInfoFunc: func() InterworkInfo {
			return c.InterworkInfo()
		},
		ConsoleRegistrar: func(cmd ConsoleCommand) error {
			return c.RegisterConsoleCommand(cmd)
		},
		Logger: func(format string, args ...interface{}) {
			c.Logf(format, args...)
		},
		RegisterPreload: func(handler PreloadHandler, priority int) error {
			return c.ListenPreloadWithPriority(handler, priority)
		},
		RegisterActive: func(handler ActiveHandler, priority int) error {
			return c.ListenActiveWithPriority(handler, priority)
		},
		RegisterPlayerJoin: func(handler PlayerEventHandler, priority int) error {
			return c.ListenPlayerJoinWithPriority(handler, priority)
		},
		RegisterPlayerLeave: func(handler PlayerEventHandler, priority int) error {
			return c.ListenPlayerLeaveWithPriority(handler, priority)
		},
		RegisterChat: func(handler ChatHandler, priority int) error {
			return c.ListenChatWithPriority(handler, priority)
		},
		RegisterFrameExit: func(handler FrameExitHandler, priority int) error {
			return c.ListenFrameExitWithPriority(handler, priority)
		},
		RegisterPacket: func(handler PacketHandler, packetIDs []uint32, priority int) error {
			return c.ListenPacketWithPriority(handler, priority, packetIDs...)
		},
		RegisterPacketAll: func(handler PacketHandler, priority int) error {
			return c.ListenPacketAllWithPriority(handler, priority)
		},
		CancelChatMessage: func(sender, message string) {
			c.CancelMessage(sender, message)
		},
		WaitPlayerMessage: func(playerName string, timeout time.Duration) (string, error) {
			return c.WaitMessage(playerName, timeout)
		},
		RegisterBroadcast: func(name string, handler BroadcastHandler, priority int) error {
			return c.ListenBroadcastWithPriority(name, handler, priority)
		},
		TriggerBroadcast: func(broadcast Broadcast) []interface{} {
			return c.Broadcast(broadcast)
		},
	}
	return NewContext(opts)
}

func (c *ContextGRPCProxy) allocCallbackID() uint32 {
	c.callbacksMu.Lock()
	defer c.callbacksMu.Unlock()
	id := c.nextCallbackID
	c.nextCallbackID++
	return id
}

// 实现 Context 接口

func (c *ContextGRPCProxy) PluginName() string {
	resp, err := c.client.GetPluginName(context.Background(), &Empty{})
	if err != nil {
		return c.pluginName
	}
	return resp.Value
}

func (c *ContextGRPCProxy) BotInfo() BotInfo {
	resp, err := c.client.GetBotInfo(context.Background(), &Empty{})
	if err != nil {
		return BotInfo{}
	}
	return BotInfo{
		Name: resp.BotName,
		XUID: resp.BotUuid,
	}
}

func (c *ContextGRPCProxy) ServerInfo() ServerInfo {
	resp, err := c.client.GetServerInfo(context.Background(), &Empty{})
	if err != nil {
		return ServerInfo{}
	}
	passcodeSet := resp.ServerPassword == "true"
	return ServerInfo{
		Code:        resp.ServerCode,
		PasscodeSet: passcodeSet,
	}
}

func (c *ContextGRPCProxy) QQInfo() QQInfo {
	_, err := c.client.GetQQInfo(context.Background(), &Empty{})
	if err != nil {
		return QQInfo{}
	}
	// QQInfo 实际字段: Adapter, WSURL, HasAccessToken
	// proto 定义: BotQq, BotNick, AdminQq
	// 暂时返回空值
	return QQInfo{
		Adapter:        "",
		WSURL:          "",
		HasAccessToken: false,
	}
}

func (c *ContextGRPCProxy) InterworkInfo() InterworkInfo {
	resp, err := c.client.GetInterworkInfo(context.Background(), &Empty{})
	if err != nil {
		return InterworkInfo{}
	}
	return InterworkInfo{
		LinkedGroups: resp.LinkedGroups,
	}
}

func (c *ContextGRPCProxy) RegisterConsoleCommand(cmd ConsoleCommand) error {
	callbackID := c.allocCallbackID()
	c.callbackServer.RegisterConsoleCommandHandler(callbackID, cmd.Handler)

	resp, err := c.client.RegisterConsoleCommand(context.Background(), &RegisterConsoleCommandRequest{
		Name:       cmd.Name,
		Triggers:   cmd.Triggers,
		Usage:      cmd.Usage,
		CallbackId: callbackID,
	})
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}
	return nil
}

func (c *ContextGRPCProxy) Logf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	c.client.Log(context.Background(), &LogRequest{Message: msg})
}

func (c *ContextGRPCProxy) LogInfo(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	c.client.LogInfo(context.Background(), &LogRequest{Message: msg})
}

func (c *ContextGRPCProxy) LogSuccess(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	c.client.LogSuccess(context.Background(), &LogRequest{Message: msg})
}

func (c *ContextGRPCProxy) LogWarning(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	c.client.LogWarning(context.Background(), &LogRequest{Message: msg})
}

func (c *ContextGRPCProxy) LogError(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	c.client.LogError(context.Background(), &LogRequest{Message: msg})
}

func (c *ContextGRPCProxy) ListenPreload(handler PreloadHandler) error {
	return c.ListenPreloadWithPriority(handler, 0)
}

func (c *ContextGRPCProxy) ListenPreloadWithPriority(handler PreloadHandler, priority int) error {
	callbackID := c.allocCallbackID()
	c.callbackServer.RegisterPreloadHandler(callbackID, handler)

	resp, err := c.client.RegisterPreloadHandler(context.Background(), &RegisterHandlerRequest{
		CallbackId: callbackID,
		Priority:   int32(priority),
	})
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}
	return nil
}

func (c *ContextGRPCProxy) ListenActive(handler ActiveHandler) error {
	return c.ListenActiveWithPriority(handler, 0)
}

func (c *ContextGRPCProxy) ListenActiveWithPriority(handler ActiveHandler, priority int) error {
	callbackID := c.allocCallbackID()
	c.callbackServer.RegisterActiveHandler(callbackID, handler)

	resp, err := c.client.RegisterActiveHandler(context.Background(), &RegisterHandlerRequest{
		CallbackId: callbackID,
		Priority:   int32(priority),
	})
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}
	return nil
}

func (c *ContextGRPCProxy) ListenPlayerJoin(handler PlayerEventHandler) error {
	return c.ListenPlayerJoinWithPriority(handler, 0)
}

func (c *ContextGRPCProxy) ListenPlayerJoinWithPriority(handler PlayerEventHandler, priority int) error {
	callbackID := c.allocCallbackID()
	c.callbackServer.RegisterPlayerJoinHandler(callbackID, handler)

	resp, err := c.client.RegisterPlayerJoinHandler(context.Background(), &RegisterHandlerRequest{
		CallbackId: callbackID,
		Priority:   int32(priority),
	})
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}
	return nil
}

func (c *ContextGRPCProxy) ListenPlayerLeave(handler PlayerEventHandler) error {
	return c.ListenPlayerLeaveWithPriority(handler, 0)
}

func (c *ContextGRPCProxy) ListenPlayerLeaveWithPriority(handler PlayerEventHandler, priority int) error {
	callbackID := c.allocCallbackID()
	c.callbackServer.RegisterPlayerLeaveHandler(callbackID, handler)

	resp, err := c.client.RegisterPlayerLeaveHandler(context.Background(), &RegisterHandlerRequest{
		CallbackId: callbackID,
		Priority:   int32(priority),
	})
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}
	return nil
}

func (c *ContextGRPCProxy) ListenChat(handler ChatHandler) error {
	return c.ListenChatWithPriority(handler, 0)
}

func (c *ContextGRPCProxy) ListenChatWithPriority(handler ChatHandler, priority int) error {
	callbackID := c.allocCallbackID()
	c.callbackServer.RegisterChatHandler(callbackID, handler)

	resp, err := c.client.RegisterChatHandler(context.Background(), &RegisterHandlerRequest{
		CallbackId: callbackID,
		Priority:   int32(priority),
	})
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}
	return nil
}

func (c *ContextGRPCProxy) ListenFrameExit(handler FrameExitHandler) error {
	return c.ListenFrameExitWithPriority(handler, 0)
}

func (c *ContextGRPCProxy) ListenFrameExitWithPriority(handler FrameExitHandler, priority int) error {
	callbackID := c.allocCallbackID()
	c.callbackServer.RegisterFrameExitHandler(callbackID, handler)

	resp, err := c.client.RegisterFrameExitHandler(context.Background(), &RegisterHandlerRequest{
		CallbackId: callbackID,
		Priority:   int32(priority),
	})
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}
	return nil
}

func (c *ContextGRPCProxy) ListenPacket(handler PacketHandler, packetIDs ...uint32) error {
	return c.ListenPacketWithPriority(handler, 0, packetIDs...)
}

func (c *ContextGRPCProxy) ListenPacketWithPriority(handler PacketHandler, priority int, packetIDs ...uint32) error {
	callbackID := c.allocCallbackID()
	c.callbackServer.RegisterPacketHandler(callbackID, handler)

	resp, err := c.client.RegisterPacketHandler(context.Background(), &RegisterPacketHandlerRequest{
		CallbackId: callbackID,
		PacketIds:  packetIDs,
		Priority:   int32(priority),
	})
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}
	return nil
}

func (c *ContextGRPCProxy) ListenPacketAll(handler PacketHandler) error {
	return c.ListenPacketAllWithPriority(handler, 0)
}

func (c *ContextGRPCProxy) ListenPacketAllWithPriority(handler PacketHandler, priority int) error {
	callbackID := c.allocCallbackID()
	c.callbackServer.RegisterPacketHandler(callbackID, handler)

	resp, err := c.client.RegisterPacketAllHandler(context.Background(), &RegisterHandlerRequest{
		CallbackId: callbackID,
		Priority:   int32(priority),
	})
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}
	return nil
}

func (c *ContextGRPCProxy) DataPath() string {
	resp, err := c.client.GetDataPath(context.Background(), &Empty{})
	if err != nil {
		return ""
	}
	return resp.Value
}

func (c *ContextGRPCProxy) FormatDataPath(path ...string) string {
	resp, err := c.client.FormatDataPath(context.Background(), &FormatDataPathRequest{
		PathParts: path,
	})
	if err != nil {
		return ""
	}
	return resp.Value
}

func (c *ContextGRPCProxy) CancelMessage(sender, message string) {
	c.client.CancelMessage(context.Background(), &CancelMessageRequest{
		Sender:  sender,
		Message: message,
	})
}

func (c *ContextGRPCProxy) WaitMessage(playerName string, timeout time.Duration) (string, error) {
	resp, err := c.client.WaitMessage(context.Background(), &WaitMessageRequest{
		PlayerName: playerName,
		TimeoutMs:  int64(timeout / time.Millisecond),
	})
	if err != nil {
		return "", err
	}
	if !resp.Success {
		return "", fmt.Errorf(resp.Error)
	}
	return resp.Message, nil
}

func (c *ContextGRPCProxy) ListenBroadcast(name string, handler BroadcastHandler) error {
	return c.ListenBroadcastWithPriority(name, handler, 0)
}

func (c *ContextGRPCProxy) ListenBroadcastWithPriority(name string, handler BroadcastHandler, priority int) error {
	callbackID := c.allocCallbackID()
	c.callbackServer.RegisterBroadcastHandler(callbackID, handler)

	resp, err := c.client.RegisterBroadcastHandler(context.Background(), &RegisterBroadcastHandlerRequest{
		EventName:  name,
		CallbackId: callbackID,
		Priority:   int32(priority),
	})
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}
	return nil
}

func (c *ContextGRPCProxy) Broadcast(broadcast Broadcast) []interface{} {
	dataBytes, err := json.Marshal(broadcast.Data)
	if err != nil {
		return []interface{}{}
	}

	resp, err := c.client.TriggerBroadcast(context.Background(), &TriggerBroadcastRequest{
		Name: broadcast.Name,
		Data: dataBytes,
	})
	if err != nil {
		return []interface{}{}
	}

	var results []interface{}
	if err := json.Unmarshal(resp.Results, &results); err != nil {
		return []interface{}{}
	}
	return results
}

// 这些方法返回 nil 或代理实现
func (c *ContextGRPCProxy) GameUtils() *GameUtils {
	return &GameUtils{
		sayToFunc: func(player, message string) {
			resp, err := c.client.SayTo(context.Background(), &SayToRequest{
				Player:  player,
				Message: message,
			})
			if err != nil {
				fmt.Printf("[GameUtils gRPC] SayTo 调用失败: %v\n", err)
				return
			}
			if !resp.Success {
				fmt.Printf("[GameUtils gRPC] SayTo 返回失败: %s\n", resp.Error)
			}
		},
	}
}
func (c *ContextGRPCProxy) Utils() *Utils                   { return NewUtils() }
func (c *ContextGRPCProxy) Translator() *Translator         { return NewTranslator() }
func (c *ContextGRPCProxy) Console() *Console               { return NewConsole(c.pluginName) }
func (c *ContextGRPCProxy) Config(configDir ...string) *Config { return NewConfig(c.pluginName, configDir...) }
func (c *ContextGRPCProxy) TempJSON(defaultDir ...string) *TempJSON { return NewTempJSON(defaultDir...) }
func (c *ContextGRPCProxy) PlayerManager() *PlayerManager   { return nil }
func (c *ContextGRPCProxy) PacketWaiter() *PacketWaiter     { return nil }
func (c *ContextGRPCProxy) GetPluginAPI(name string) (Plugin, PluginAPIVersion, error) {
	return nil, PluginAPIVersion{}, fmt.Errorf("GetPluginAPI not supported in gRPC plugins")
}
func (c *ContextGRPCProxy) GetPluginAPIWithVersion(name string, version PluginAPIVersion) (Plugin, error) {
	return nil, fmt.Errorf("GetPluginAPIWithVersion not supported in gRPC plugins")
}
func (c *ContextGRPCProxy) RegisterPluginAPI(name string, version PluginAPIVersion, plugin Plugin) error {
	return fmt.Errorf("RegisterPluginAPI not supported in gRPC plugins")
}
func (c *ContextGRPCProxy) ListPluginAPIs() []PluginAPIInfo {
	return []PluginAPIInfo{}
}
