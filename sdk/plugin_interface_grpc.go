package sdk

import (
	"context"
	"fmt"
	"sync"

	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
)

// 这是插件系统的 gRPC 协议定义
// 使得插件可以跨平台运行（Windows/Linux/macOS/Android）

// PluginGRPC 是插件的 gRPC 实现
type PluginGRPC struct {
	plugin.Plugin
	Impl Plugin
}

func (p *PluginGRPC) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	RegisterPluginServiceServer(s, &GRPCServer{
		Impl:   p.Impl,
		broker: broker,
	})
	return nil
}

func (p *PluginGRPC) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &GRPCClient{
		client: NewPluginServiceClient(c),
		broker: broker,
	}, nil
}

// GRPCServer 是服务端实现
type GRPCServer struct {
	UnimplementedPluginServiceServer
	Impl          Plugin
	broker        *plugin.GRPCBroker
	callbackServer *CallbackServerImpl
	ctxMutex      sync.RWMutex
}

func (s *GRPCServer) Init(ctx context.Context, req *InitRequest) (*InitResponse, error) {
	// 1. 连接到主进程的 ContextService
	conn, err := s.broker.Dial(req.ContextServiceId)
	if err != nil {
		return &InitResponse{Success: false}, err
	}
	contextClient := NewContextServiceClient(conn)

	// 2. 创建 CallbackServer，用于接收主进程的事件回调
	s.callbackServer = NewCallbackServerImpl()
	callbackServiceID := s.broker.NextId()

	go s.broker.AcceptAndServe(callbackServiceID, func(opts []grpc.ServerOption) *grpc.Server {
		server := grpc.NewServer(opts...)
		RegisterCallbackServiceServer(server, s.callbackServer)
		return server
	})

	// 3. 创建 ContextGRPCProxy 并转换为 Context
	pluginName := "grpc-plugin" // 默认名称
	if info := s.Impl.GetInfo(); info.Name != "" {
		pluginName = info.Name
	}
	ctxProxy := NewContextGRPCProxy(pluginName, contextClient, s.callbackServer)
	pluginCtx := ctxProxy.ToContext()

	// 4. 调用插件的 Init
	err = s.Impl.Init(pluginCtx)
	if err != nil {
		return &InitResponse{Success: false}, err
	}

	return &InitResponse{
		Success:           true,
		CallbackServiceId: callbackServiceID,
	}, nil
}

func (s *GRPCServer) Start(ctx context.Context, req *StartRequest) (*StartResponse, error) {
	err := s.Impl.Start()
	if err != nil {
		return &StartResponse{Success: false}, err
	}
	return &StartResponse{Success: true}, nil
}

func (s *GRPCServer) Stop(ctx context.Context, req *StopRequest) (*StopResponse, error) {
	err := s.Impl.Stop()
	if err != nil {
		return &StopResponse{Success: false}, err
	}
	return &StopResponse{Success: true}, nil
}

func (s *GRPCServer) GetInfo(ctx context.Context, req *GetInfoRequest) (*GetInfoResponse, error) {
	info := s.Impl.GetInfo()
	return &GetInfoResponse{
		Name:        info.Name,
		DisplayName: info.DisplayName,
		Version:     info.Version,
		Description: info.Description,
		Author:      info.Author,
	}, nil
}

// GRPCClient 是客户端实现
type GRPCClient struct {
	client         PluginServiceClient
	broker         *plugin.GRPCBroker
	contextServer  *ContextServer
	callbackClient CallbackServiceClient
}

func (c *GRPCClient) Init(ctx *Context) error {
	// 1. 创建 ContextServer，暴露主进程的 Context 给插件
	c.contextServer = NewContextServer(ctx)
	contextServiceID := c.broker.NextId()

	go c.broker.AcceptAndServe(contextServiceID, func(opts []grpc.ServerOption) *grpc.Server {
		server := grpc.NewServer(opts...)
		RegisterContextServiceServer(server, c.contextServer)
		return server
	})

	// 2. 调用插件的 Init
	resp, err := c.client.Init(context.Background(), &InitRequest{
		ContextServiceId: contextServiceID,
	})
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("plugin init failed")
	}

	// 3. 连接到插件的 CallbackService
	conn, err := c.broker.Dial(resp.CallbackServiceId)
	if err != nil {
		return fmt.Errorf("failed to connect to callback service: %v", err)
	}
	c.callbackClient = NewCallbackServiceClient(conn)

	// 4. 将 callbackClient 设置到 contextServer，执行所有待处理的注册
	c.contextServer.SetCallbackClient(c.callbackClient)

	return nil
}

func (c *GRPCClient) Start() error {
	resp, err := c.client.Start(context.Background(), &StartRequest{})
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("plugin start failed")
	}
	return nil
}

func (c *GRPCClient) Stop() error {
	resp, err := c.client.Stop(context.Background(), &StopRequest{})
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("plugin stop failed")
	}
	return nil
}

func (c *GRPCClient) GetInfo() PluginInfo {
	resp, err := c.client.GetInfo(context.Background(), &GetInfoRequest{})
	if err != nil {
		return PluginInfo{}
	}
	return PluginInfo{
		Name:        resp.Name,
		DisplayName: resp.DisplayName,
		Version:     resp.Version,
		Description: resp.Description,
		Author:      resp.Author,
	}
}

// HandshakeConfig 用于握手验证
var HandshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "FUN_INTERWORK_PLUGIN",
	MagicCookieValue: "funinterwork_v1",
}

// PluginMap 是插件映射表
var PluginMap = map[string]plugin.Plugin{
	"plugin": &PluginGRPC{},
}
