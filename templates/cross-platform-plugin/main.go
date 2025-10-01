package main

import (
	"github.com/hashicorp/go-plugin"
	"github.com/maoqijie/FIN-plugin/sdk"
)

// ExamplePlugin 是示例插件实现
type ExamplePlugin struct {
	ctx *sdk.Context
}

// Init 初始化插件
func (p *ExamplePlugin) Init(ctx *sdk.Context) error {
	p.ctx = ctx

	// 注册控制台命令
	ctx.RegisterConsoleCommand(sdk.ConsoleCommand{
		Name:        "example",
		Triggers:    []string{"example", "ex"},
		Usage:       "example <arg>",
		Description: "示例插件命令",
		Handler: func(args []string) error {
			ctx.LogInfo("示例命令被调用，参数: %v", args)
			return nil
		},
	})

	// 监听玩家加入事件
	ctx.ListenPlayerJoin(func(event sdk.PlayerEvent) {
		ctx.LogSuccess("玩家 %s 加入了游戏", event.Name)
	})

	// 监听聊天消息
	ctx.ListenChat(func(event *sdk.ChatEvent) {
		if event.Message == "hello" {
			// 使用 GameUtils 发送消息
			if gu := ctx.GameUtils(); gu != nil {
				gu.SayTo(event.Sender, "§aHello, "+event.Sender+"!")
			}
		}
	})

	ctx.LogInfo("插件初始化完成")
	return nil
}

// Start 启动插件
func (p *ExamplePlugin) Start() error {
	p.ctx.LogSuccess("插件已启动")
	return nil
}

// Stop 停止插件
func (p *ExamplePlugin) Stop() error {
	p.ctx.LogWarning("插件正在停止...")
	return nil
}

// GetInfo 返回插件信息
func (p *ExamplePlugin) GetInfo() sdk.PluginInfo {
	return sdk.PluginInfo{
		Name:        "example-plugin",
		DisplayName: "示例插件",
		Version:     "1.0.0",
		Description: "跨平台示例插件",
		Author:      "作者名",
	}
}

// main 函数作为 go-plugin 服务器运行
func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: sdk.HandshakeConfig,
		Plugins: map[string]plugin.Plugin{
			"plugin": &sdk.PluginGRPC{Impl: &ExamplePlugin{}},
		},
		// 使用 gRPC 协议
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
