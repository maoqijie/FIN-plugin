package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// ContextServer 包装真实的 Context，将其方法暴露为 gRPC 服务
// 运行在主进程中，接收插件的远程调用
type ContextServer struct {
	UnimplementedContextServiceServer

	ctx            *Context
	callbackClient CallbackServiceClient

	// 回调 ID 管理
	nextCallbackID uint32
	callbacks      map[uint32]*callbackInfo
	callbacksMu    sync.RWMutex

	// 延迟注册：当 callbackClient 为 nil 时缓存注册请求
	pendingRegistrations []func() error
	pendingMu            sync.Mutex
}

type callbackInfo struct {
	callbackID  uint32
	handlerType string // "chat", "player_join", "player_leave", etc.
}

func NewContextServer(ctx *Context) *ContextServer {
	return &ContextServer{
		ctx:                  ctx,
		callbacks:            make(map[uint32]*callbackInfo),
		nextCallbackID:       1,
		pendingRegistrations: make([]func() error, 0),
	}
}

func (s *ContextServer) SetCallbackClient(client CallbackServiceClient) {
	s.callbackClient = client

	// 执行所有待处理的注册
	s.pendingMu.Lock()
	pending := s.pendingRegistrations
	s.pendingRegistrations = nil
	s.pendingMu.Unlock()

	for _, register := range pending {
		if err := register(); err != nil {
			s.ctx.LogError("延迟注册失败: %v", err)
		}
	}
}

// deferOrExecute 延迟执行或立即执行注册函数
func (s *ContextServer) deferOrExecute(register func() error) error {
	if s.callbackClient == nil {
		s.pendingMu.Lock()
		s.pendingRegistrations = append(s.pendingRegistrations, register)
		s.pendingMu.Unlock()
		return nil
	}
	return register()
}

// 日志方法
func (s *ContextServer) Log(ctx context.Context, req *LogRequest) (*LogResponse, error) {
	s.ctx.Logf("%s", req.Message)
	return &LogResponse{Success: true}, nil
}

func (s *ContextServer) LogInfo(ctx context.Context, req *LogRequest) (*LogResponse, error) {
	s.ctx.LogInfo("%s", req.Message)
	return &LogResponse{Success: true}, nil
}

func (s *ContextServer) LogSuccess(ctx context.Context, req *LogRequest) (*LogResponse, error) {
	s.ctx.LogSuccess("%s", req.Message)
	return &LogResponse{Success: true}, nil
}

func (s *ContextServer) LogWarning(ctx context.Context, req *LogRequest) (*LogResponse, error) {
	s.ctx.LogWarning("%s", req.Message)
	return &LogResponse{Success: true}, nil
}

func (s *ContextServer) LogError(ctx context.Context, req *LogRequest) (*LogResponse, error) {
	s.ctx.LogError("%s", req.Message)
	return &LogResponse{Success: true}, nil
}

// 信息查询方法
func (s *ContextServer) GetPluginName(ctx context.Context, req *Empty) (*StringResponse, error) {
	return &StringResponse{Value: s.ctx.PluginName()}, nil
}

func (s *ContextServer) GetBotInfo(ctx context.Context, req *Empty) (*BotInfoResponse, error) {
	info := s.ctx.BotInfo()
	return &BotInfoResponse{
		BotName: info.Name,
		BotUuid: info.XUID,
	}, nil
}

func (s *ContextServer) GetServerInfo(ctx context.Context, req *Empty) (*ServerInfoResponse, error) {
	info := s.ctx.ServerInfo()
	passcodeSet := "false"
	if info.PasscodeSet {
		passcodeSet = "true"
	}
	return &ServerInfoResponse{
		ServerCode:     info.Code,
		ServerPassword: passcodeSet,
		ServerAddress:  "",
	}, nil
}

func (s *ContextServer) GetQQInfo(ctx context.Context, req *Empty) (*QQInfoResponse, error) {
	// QQInfo 实际字段: Adapter, WSURL, HasAccessToken
	// proto 定义: BotQq, BotNick, AdminQq
	// 暂时返回空值，后续可扩展
	return &QQInfoResponse{
		BotQq:   0,
		BotNick: "",
		AdminQq: []uint64{},
	}, nil
}

func (s *ContextServer) GetInterworkInfo(ctx context.Context, req *Empty) (*InterworkInfoResponse, error) {
	info := s.ctx.InterworkInfo()
	return &InterworkInfoResponse{
		LinkedGroups: info.LinkedGroups,
	}, nil
}

// 路径方法
func (s *ContextServer) GetDataPath(ctx context.Context, req *Empty) (*StringResponse, error) {
	return &StringResponse{Value: s.ctx.DataPath()}, nil
}

func (s *ContextServer) FormatDataPath(ctx context.Context, req *FormatDataPathRequest) (*StringResponse, error) {
	path := s.ctx.FormatDataPath(req.PathParts...)
	return &StringResponse{Value: path}, nil
}

// 游戏工具方法
func (s *ContextServer) SayTo(ctx context.Context, req *SayToRequest) (*BoolResponse, error) {
	gu := s.ctx.GameUtils()
	if gu == nil {
		return &BoolResponse{Success: false, Error: "GameUtils not available"}, nil
	}
	gu.SayTo(req.Player, req.Message)
	return &BoolResponse{Success: true}, nil
}

// 控制台命令注册
func (s *ContextServer) RegisterConsoleCommand(ctx context.Context, req *RegisterConsoleCommandRequest) (*BoolResponse, error) {
	callbackID := req.CallbackId
	name := req.Name
	triggers := req.Triggers
	usage := req.Usage

	err := s.deferOrExecute(func() error {
		cmd := ConsoleCommand{
			Name:     name,
			Triggers: triggers,
			Usage:    usage,
			Handler: func(args []string) error {
				// 通过 gRPC 回调插件
				resp, err := s.callbackClient.OnConsoleCommand(context.Background(), &ConsoleCommandRequest{
					CallbackId: callbackID,
					Args:       args,
				})
				if err != nil {
					return err
				}
				if !resp.Success {
					return fmt.Errorf(resp.Error)
				}
				return nil
			},
		}

		if err := s.ctx.RegisterConsoleCommand(cmd); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return &BoolResponse{Success: false, Error: err.Error()}, nil
	}
	return &BoolResponse{Success: true}, nil
}

// 事件注册方法
func (s *ContextServer) RegisterChatHandler(ctx context.Context, req *RegisterHandlerRequest) (*RegisterHandlerResponse, error) {
	callbackID := req.CallbackId
	priority := int(req.Priority)

	err := s.deferOrExecute(func() error {
		handler := func(event *ChatEvent) {
			resp, err := s.callbackClient.OnChatEvent(context.Background(), &ChatEventRequest{
				CallbackId: callbackID,
				Sender:     event.Sender,
				Message:    event.Message,
			})
			if err != nil {
				s.ctx.LogError("Chat handler gRPC call failed: %v", err)
				return
			}
			if resp.Cancel {
				event.Cancelled = true
			}
		}

		if err := s.ctx.ListenChatWithPriority(handler, priority); err != nil {
			return err
		}

		s.callbacksMu.Lock()
		s.callbacks[callbackID] = &callbackInfo{
			callbackID:  callbackID,
			handlerType: "chat",
		}
		s.callbacksMu.Unlock()
		return nil
	})

	if err != nil {
		return &RegisterHandlerResponse{Success: false, Error: err.Error()}, nil
	}
	return &RegisterHandlerResponse{Success: true, HandlerId: callbackID}, nil
}

func (s *ContextServer) RegisterPlayerJoinHandler(ctx context.Context, req *RegisterHandlerRequest) (*RegisterHandlerResponse, error) {
	callbackID := req.CallbackId
	priority := int(req.Priority)

	err := s.deferOrExecute(func() error {
		handler := func(event PlayerEvent) {
			rawData, err := json.Marshal(event.Raw)
			if err != nil {
				s.ctx.LogError("Failed to serialize PlayerEvent: %v", err)
				return
			}
			_, err = s.callbackClient.OnPlayerJoinEvent(context.Background(), &PlayerEventRequest{
				CallbackId: callbackID,
				RawData:    rawData,
			})
			if err != nil {
				s.ctx.LogError("PlayerJoin handler gRPC call failed: %v", err)
			}
		}

		if err := s.ctx.ListenPlayerJoinWithPriority(handler, priority); err != nil {
			return err
		}

		s.callbacksMu.Lock()
		s.callbacks[callbackID] = &callbackInfo{
			callbackID:  callbackID,
			handlerType: "player_join",
		}
		s.callbacksMu.Unlock()
		return nil
	})

	if err != nil {
		return &RegisterHandlerResponse{Success: false, Error: err.Error()}, nil
	}
	return &RegisterHandlerResponse{Success: true, HandlerId: callbackID}, nil
}

func (s *ContextServer) RegisterPlayerLeaveHandler(ctx context.Context, req *RegisterHandlerRequest) (*RegisterHandlerResponse, error) {
	callbackID := req.CallbackId
	priority := int(req.Priority)

	err := s.deferOrExecute(func() error {
		handler := func(event PlayerEvent) {
			rawData, err := json.Marshal(event.Raw)
			if err != nil {
				s.ctx.LogError("Failed to serialize PlayerEvent: %v", err)
				return
			}
			_, err = s.callbackClient.OnPlayerLeaveEvent(context.Background(), &PlayerEventRequest{
				CallbackId: callbackID,
				RawData:    rawData,
			})
			if err != nil {
				s.ctx.LogError("PlayerLeave handler gRPC call failed: %v", err)
			}
		}

		if err := s.ctx.ListenPlayerLeaveWithPriority(handler, priority); err != nil {
			return err
		}

		s.callbacksMu.Lock()
		s.callbacks[callbackID] = &callbackInfo{
			callbackID:  callbackID,
			handlerType: "player_leave",
		}
		s.callbacksMu.Unlock()
		return nil
	})

	if err != nil {
		return &RegisterHandlerResponse{Success: false, Error: err.Error()}, nil
	}
	return &RegisterHandlerResponse{Success: true, HandlerId: callbackID}, nil
}

func (s *ContextServer) RegisterPacketHandler(ctx context.Context, req *RegisterPacketHandlerRequest) (*RegisterHandlerResponse, error) {
	callbackID := req.CallbackId
	priority := int(req.Priority)
	packetIDs := req.PacketIds

	err := s.deferOrExecute(func() error {
		handler := func(event PacketEvent) {
			packetData, err := json.Marshal(event.Raw)
			if err != nil {
				s.ctx.LogError("Failed to serialize packet: %v", err)
				return
			}

			_, err = s.callbackClient.OnPacketEvent(context.Background(), &PacketEventRequest{
				CallbackId: callbackID,
				PacketId:   event.ID,
				PacketData: packetData,
			})
			if err != nil {
				s.ctx.LogError("Packet handler gRPC call failed: %v", err)
			}
		}

		if err := s.ctx.ListenPacketWithPriority(handler, priority, packetIDs...); err != nil {
			return err
		}

		s.callbacksMu.Lock()
		s.callbacks[callbackID] = &callbackInfo{
			callbackID:  callbackID,
			handlerType: "packet",
		}
		s.callbacksMu.Unlock()
		return nil
	})

	if err != nil {
		return &RegisterHandlerResponse{Success: false, Error: err.Error()}, nil
	}
	return &RegisterHandlerResponse{Success: true, HandlerId: callbackID}, nil
}

func (s *ContextServer) RegisterPacketAllHandler(ctx context.Context, req *RegisterHandlerRequest) (*RegisterHandlerResponse, error) {
	callbackID := req.CallbackId
	priority := int(req.Priority)

	err := s.deferOrExecute(func() error {
		handler := func(event PacketEvent) {
			packetData, err := json.Marshal(event.Raw)
			if err != nil {
				s.ctx.LogError("Failed to serialize packet: %v", err)
				return
			}

			_, err = s.callbackClient.OnPacketEvent(context.Background(), &PacketEventRequest{
				CallbackId: callbackID,
				PacketId:   event.ID,
				PacketData: packetData,
			})
			if err != nil {
				s.ctx.LogError("PacketAll handler gRPC call failed: %v", err)
			}
		}

		if err := s.ctx.ListenPacketAllWithPriority(handler, priority); err != nil {
			return err
		}

		s.callbacksMu.Lock()
		s.callbacks[callbackID] = &callbackInfo{
			callbackID:  callbackID,
			handlerType: "packet_all",
		}
		s.callbacksMu.Unlock()
		return nil
	})

	if err != nil {
		return &RegisterHandlerResponse{Success: false, Error: err.Error()}, nil
	}
	return &RegisterHandlerResponse{Success: true, HandlerId: callbackID}, nil
}

func (s *ContextServer) RegisterPreloadHandler(ctx context.Context, req *RegisterHandlerRequest) (*RegisterHandlerResponse, error) {
	callbackID := req.CallbackId
	priority := int(req.Priority)

	err := s.deferOrExecute(func() error {
		handler := func() {
			_, err := s.callbackClient.OnPreloadEvent(context.Background(), &PreloadEventRequest{
				CallbackId: callbackID,
			})
			if err != nil {
				s.ctx.LogError("Preload handler gRPC call failed: %v", err)
			}
		}

		if err := s.ctx.ListenPreloadWithPriority(handler, priority); err != nil {
			return err
		}

		s.callbacksMu.Lock()
		s.callbacks[callbackID] = &callbackInfo{
			callbackID:  callbackID,
			handlerType: "preload",
		}
		s.callbacksMu.Unlock()
		return nil
	})

	if err != nil {
		return &RegisterHandlerResponse{Success: false, Error: err.Error()}, nil
	}
	return &RegisterHandlerResponse{Success: true, HandlerId: callbackID}, nil
}

func (s *ContextServer) RegisterActiveHandler(ctx context.Context, req *RegisterHandlerRequest) (*RegisterHandlerResponse, error) {
	callbackID := req.CallbackId
	priority := int(req.Priority)

	err := s.deferOrExecute(func() error {
		handler := func() {
			_, err := s.callbackClient.OnActiveEvent(context.Background(), &ActiveEventRequest{
				CallbackId: callbackID,
			})
			if err != nil {
				s.ctx.LogError("Active handler gRPC call failed: %v", err)
			}
		}

		if err := s.ctx.ListenActiveWithPriority(handler, priority); err != nil {
			return err
		}

		s.callbacksMu.Lock()
		s.callbacks[callbackID] = &callbackInfo{
			callbackID:  callbackID,
			handlerType: "active",
		}
		s.callbacksMu.Unlock()
		return nil
	})

	if err != nil {
		return &RegisterHandlerResponse{Success: false, Error: err.Error()}, nil
	}
	return &RegisterHandlerResponse{Success: true, HandlerId: callbackID}, nil
}

func (s *ContextServer) RegisterFrameExitHandler(ctx context.Context, req *RegisterHandlerRequest) (*RegisterHandlerResponse, error) {
	callbackID := req.CallbackId
	priority := int(req.Priority)

	err := s.deferOrExecute(func() error {
		handler := func(event FrameExitEvent) {
			_, err := s.callbackClient.OnFrameExitEvent(context.Background(), &FrameExitEventRequest{
				CallbackId: callbackID,
			})
			if err != nil {
				s.ctx.LogError("FrameExit handler gRPC call failed: %v", err)
			}
		}

		if err := s.ctx.ListenFrameExitWithPriority(handler, priority); err != nil {
			return err
		}

		s.callbacksMu.Lock()
		s.callbacks[callbackID] = &callbackInfo{
			callbackID:  callbackID,
			handlerType: "frame_exit",
		}
		s.callbacksMu.Unlock()
		return nil
	})

	if err != nil {
		return &RegisterHandlerResponse{Success: false, Error: err.Error()}, nil
	}
	return &RegisterHandlerResponse{Success: true, HandlerId: callbackID}, nil
}

func (s *ContextServer) RegisterBroadcastHandler(ctx context.Context, req *RegisterBroadcastHandlerRequest) (*RegisterHandlerResponse, error) {
	callbackID := req.CallbackId
	priority := int(req.Priority)
	eventName := req.EventName

	err := s.deferOrExecute(func() error {
		handler := func(broadcast Broadcast) interface{} {
			dataBytes, err := json.Marshal(broadcast.Data)
			if err != nil {
				s.ctx.LogError("Failed to serialize broadcast data: %v", err)
				return nil
			}

			resp, err := s.callbackClient.OnBroadcastEvent(context.Background(), &BroadcastEventRequest{
				CallbackId: callbackID,
				Name:       broadcast.Name,
				Data:       dataBytes,
			})
			if err != nil {
				s.ctx.LogError("Broadcast handler gRPC call failed: %v", err)
				return nil
			}

			var result interface{}
			if err := json.Unmarshal(resp.Result, &result); err != nil {
				s.ctx.LogError("Failed to deserialize broadcast result: %v", err)
				return nil
			}
			return result
		}

		if err := s.ctx.ListenBroadcastWithPriority(eventName, handler, priority); err != nil {
			return err
		}

		s.callbacksMu.Lock()
		s.callbacks[callbackID] = &callbackInfo{
			callbackID:  callbackID,
			handlerType: "broadcast",
		}
		s.callbacksMu.Unlock()
		return nil
	})

	if err != nil {
		return &RegisterHandlerResponse{Success: false, Error: err.Error()}, nil
	}
	return &RegisterHandlerResponse{Success: true, HandlerId: callbackID}, nil
}

// 消息控制
func (s *ContextServer) CancelMessage(ctx context.Context, req *CancelMessageRequest) (*BoolResponse, error) {
	s.ctx.CancelMessage(req.Sender, req.Message)
	return &BoolResponse{Success: true}, nil
}

func (s *ContextServer) WaitMessage(ctx context.Context, req *WaitMessageRequest) (*WaitMessageResponse, error) {
	timeout := time.Duration(req.TimeoutMs) * time.Millisecond
	message, err := s.ctx.WaitMessage(req.PlayerName, timeout)
	if err != nil {
		return &WaitMessageResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}
	return &WaitMessageResponse{
		Success: true,
		Message: message,
	}, nil
}

// 广播
func (s *ContextServer) TriggerBroadcast(ctx context.Context, req *TriggerBroadcastRequest) (*TriggerBroadcastResponse, error) {
	var data map[string]interface{}
	if err := json.Unmarshal(req.Data, &data); err != nil {
		return nil, fmt.Errorf("failed to deserialize broadcast data: %w", err)
	}

	broadcast := Broadcast{
		Name: req.Name,
		Data: data,
	}

	results := s.ctx.Broadcast(broadcast)

	resultsBytes, err := json.Marshal(results)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize broadcast results: %w", err)
	}

	return &TriggerBroadcastResponse{
		Results: resultsBytes,
	}, nil
}
