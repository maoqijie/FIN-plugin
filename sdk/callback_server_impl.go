package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
)

// CallbackServerImpl 实现 CallbackService，运行在插件进程中
// 接收主进程的事件回调
type CallbackServerImpl struct {
	UnimplementedCallbackServiceServer

	chatHandlers       map[uint32]ChatHandler
	playerJoinHandlers map[uint32]PlayerEventHandler
	playerLeaveHandlers map[uint32]PlayerEventHandler
	packetHandlers     map[uint32]PacketHandler
	preloadHandlers    map[uint32]PreloadHandler
	activeHandlers     map[uint32]ActiveHandler
	frameExitHandlers  map[uint32]FrameExitHandler
	broadcastHandlers  map[uint32]BroadcastHandler
	consoleHandlers    map[uint32]func([]string) error

	mu sync.RWMutex
}

func NewCallbackServerImpl() *CallbackServerImpl {
	return &CallbackServerImpl{
		chatHandlers:       make(map[uint32]ChatHandler),
		playerJoinHandlers: make(map[uint32]PlayerEventHandler),
		playerLeaveHandlers: make(map[uint32]PlayerEventHandler),
		packetHandlers:     make(map[uint32]PacketHandler),
		preloadHandlers:    make(map[uint32]PreloadHandler),
		activeHandlers:     make(map[uint32]ActiveHandler),
		frameExitHandlers:  make(map[uint32]FrameExitHandler),
		broadcastHandlers:  make(map[uint32]BroadcastHandler),
		consoleHandlers:    make(map[uint32]func([]string) error),
	}
}

// 注册 handler 方法
func (s *CallbackServerImpl) RegisterChatHandler(callbackID uint32, handler ChatHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.chatHandlers[callbackID] = handler
}

func (s *CallbackServerImpl) RegisterPlayerJoinHandler(callbackID uint32, handler PlayerEventHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.playerJoinHandlers[callbackID] = handler
}

func (s *CallbackServerImpl) RegisterPlayerLeaveHandler(callbackID uint32, handler PlayerEventHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.playerLeaveHandlers[callbackID] = handler
}

func (s *CallbackServerImpl) RegisterPacketHandler(callbackID uint32, handler PacketHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.packetHandlers[callbackID] = handler
}

func (s *CallbackServerImpl) RegisterPreloadHandler(callbackID uint32, handler PreloadHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.preloadHandlers[callbackID] = handler
}

func (s *CallbackServerImpl) RegisterActiveHandler(callbackID uint32, handler ActiveHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.activeHandlers[callbackID] = handler
}

func (s *CallbackServerImpl) RegisterFrameExitHandler(callbackID uint32, handler FrameExitHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.frameExitHandlers[callbackID] = handler
}

func (s *CallbackServerImpl) RegisterBroadcastHandler(callbackID uint32, handler BroadcastHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.broadcastHandlers[callbackID] = handler
}

func (s *CallbackServerImpl) RegisterConsoleCommandHandler(callbackID uint32, handler func([]string) error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.consoleHandlers[callbackID] = handler
}

// gRPC 回调方法实现
func (s *CallbackServerImpl) OnChatEvent(ctx context.Context, req *ChatEventRequest) (*ChatEventResponse, error) {
	s.mu.RLock()
	handler, ok := s.chatHandlers[req.CallbackId]
	s.mu.RUnlock()

	if !ok {
		return &ChatEventResponse{Cancel: false}, fmt.Errorf("chat handler %d not found", req.CallbackId)
	}

	event := &ChatEvent{
		Sender:  req.Sender,
		Message: req.Message,
	}

	handler(event)
	return &ChatEventResponse{Cancel: event.Cancelled}, nil
}

func (s *CallbackServerImpl) OnPlayerJoinEvent(ctx context.Context, req *PlayerEventRequest) (*PlayerEventResponse, error) {
	s.mu.RLock()
	handler, ok := s.playerJoinHandlers[req.CallbackId]
	s.mu.RUnlock()

	if !ok {
		return &PlayerEventResponse{Success: false}, fmt.Errorf("player join handler %d not found", req.CallbackId)
	}

	// 反序列化 PlayerEvent.Raw
	var raw interface{}
	if err := json.Unmarshal(req.RawData, &raw); err != nil {
		return &PlayerEventResponse{Success: false}, err
	}

	event := PlayerEvent{Raw: raw}
	handler(event)

	return &PlayerEventResponse{Success: true}, nil
}

func (s *CallbackServerImpl) OnPlayerLeaveEvent(ctx context.Context, req *PlayerEventRequest) (*PlayerEventResponse, error) {
	s.mu.RLock()
	handler, ok := s.playerLeaveHandlers[req.CallbackId]
	s.mu.RUnlock()

	if !ok {
		return &PlayerEventResponse{Success: false}, fmt.Errorf("player leave handler %d not found", req.CallbackId)
	}

	var raw interface{}
	if err := json.Unmarshal(req.RawData, &raw); err != nil {
		return &PlayerEventResponse{Success: false}, err
	}

	event := PlayerEvent{Raw: raw}
	handler(event)

	return &PlayerEventResponse{Success: true}, nil
}

func (s *CallbackServerImpl) OnPacketEvent(ctx context.Context, req *PacketEventRequest) (*PacketEventResponse, error) {
	s.mu.RLock()
	handler, ok := s.packetHandlers[req.CallbackId]
	s.mu.RUnlock()

	if !ok {
		return &PacketEventResponse{Success: false}, fmt.Errorf("packet handler %d not found", req.CallbackId)
	}

	var packet interface{}
	if err := json.Unmarshal(req.PacketData, &packet); err != nil {
		return &PacketEventResponse{Success: false}, err
	}

	event := PacketEvent{
		ID:  req.PacketId,
		Raw: packet,
	}
	handler(event)

	return &PacketEventResponse{Success: true}, nil
}

func (s *CallbackServerImpl) OnPreloadEvent(ctx context.Context, req *PreloadEventRequest) (*PreloadEventResponse, error) {
	s.mu.RLock()
	handler, ok := s.preloadHandlers[req.CallbackId]
	s.mu.RUnlock()

	if !ok {
		return &PreloadEventResponse{Success: false, Error: "handler not found"}, nil
	}

	handler()
	return &PreloadEventResponse{Success: true}, nil
}

func (s *CallbackServerImpl) OnActiveEvent(ctx context.Context, req *ActiveEventRequest) (*ActiveEventResponse, error) {
	s.mu.RLock()
	handler, ok := s.activeHandlers[req.CallbackId]
	s.mu.RUnlock()

	if !ok {
		return &ActiveEventResponse{Success: false, Error: "handler not found"}, nil
	}

	handler()
	return &ActiveEventResponse{Success: true}, nil
}

func (s *CallbackServerImpl) OnFrameExitEvent(ctx context.Context, req *FrameExitEventRequest) (*FrameExitEventResponse, error) {
	s.mu.RLock()
	handler, ok := s.frameExitHandlers[req.CallbackId]
	s.mu.RUnlock()

	if !ok {
		return &FrameExitEventResponse{Success: false}, fmt.Errorf("frame exit handler %d not found", req.CallbackId)
	}

	// 创建 FrameExitEvent - 当前 proto 没有传递 Signal 和 Reason，使用空值
	event := FrameExitEvent{
		Signal: "",
		Reason: "",
	}
	handler(event)
	return &FrameExitEventResponse{Success: true}, nil
}

func (s *CallbackServerImpl) OnBroadcastEvent(ctx context.Context, req *BroadcastEventRequest) (*BroadcastEventResponse, error) {
	s.mu.RLock()
	handler, ok := s.broadcastHandlers[req.CallbackId]
	s.mu.RUnlock()

	if !ok {
		return &BroadcastEventResponse{}, fmt.Errorf("broadcast handler %d not found", req.CallbackId)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(req.Data, &data); err != nil {
		return &BroadcastEventResponse{}, err
	}

	broadcast := Broadcast{
		Name: req.Name,
		Data: data,
	}

	result := handler(broadcast)

	resultBytes, err := json.Marshal(result)
	if err != nil {
		return &BroadcastEventResponse{}, err
	}

	return &BroadcastEventResponse{Result: resultBytes}, nil
}

func (s *CallbackServerImpl) OnConsoleCommand(ctx context.Context, req *ConsoleCommandRequest) (*ConsoleCommandResponse, error) {
	s.mu.RLock()
	handler, ok := s.consoleHandlers[req.CallbackId]
	s.mu.RUnlock()

	if !ok {
		return &ConsoleCommandResponse{Success: false, Error: "handler not found"}, nil
	}

	err := handler(req.Args)
	if err != nil {
		return &ConsoleCommandResponse{Success: false, Error: err.Error()}, nil
	}

	return &ConsoleCommandResponse{Success: true}, nil
}
