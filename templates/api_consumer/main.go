package main

import (
	"fmt"

	sdk "github.com/maoqijie/FIN-plugin/sdk"
)

// APIConsumerPlugin 使用其他插件 API 的示例插件
type APIConsumerPlugin struct {
	ctx *sdk.Context
}

// ExampleAPI 定义 API 接口（可选，用于类型安全和 IDE 提示）
// 实际使用时，可以从 API 插件包导入，或定义接口契约
type ExampleAPI interface {
	Greet(name string) string
	SendMessage(message string) error
	GetServerStatus() map[string]interface{}
	CalculateDistance(x1, y1, z1, x2, y2, z2 float32) float32
}

// Init 插件初始化
func (p *APIConsumerPlugin) Init(ctx *sdk.Context) error {
	p.ctx = ctx

	// 在 Preload 阶段获取 API（确保依赖的 API 插件已加载）
	ctx.ListenPreload(func() {
		p.onPreload()
	})

	// 注册控制台命令测试 API
	ctx.RegisterConsoleCommand(sdk.ConsoleCommand{
		Name:        "test-api",
		Triggers:    []string{"testapi", "tapi"},
		Description: "测试 API 插件功能",
		Usage:       "测试调用前置插件的各种方法",
		Handler:     p.handleTestAPI,
	})

	return nil
}

// onPreload 预加载回调
func (p *APIConsumerPlugin) onPreload() {
	// 方式 1: 获取 API 不检查版本
	api, version, err := p.ctx.GetPluginAPI("example-api")
	if err != nil {
		p.ctx.Logf("§c获取 example-api 失败: %v", err)
		p.ctx.Logf("§e请确保 example-api 插件已加载")
		return
	}

	p.ctx.Logf("§a成功获取 example-api，版本: %s", version.String())

	// 方式 2: 使用类型断言调用方法
	// 注意: 这里需要知道实际的插件类型，或使用接口定义
	// 如果有 example-api 的源码，可以直接导入其类型

	// 使用反射或类型断言调用方法
	// 这里演示直接调用（需要在实际环境中根据具体类型调整）
	p.ctx.Logf("§aAPI 插件实例获取成功，可以通过类型断言调用方法")

	// 列出所有可用的 API
	apis := p.ctx.ListPluginAPIs()
	p.ctx.Logf("§b当前已注册的 API 插件:")
	for _, apiInfo := range apis {
		p.ctx.Logf("  - %s (版本 %s)", apiInfo.Name, apiInfo.Version.String())
	}
}

// handleTestAPI 测试 API 命令处理器
func (p *APIConsumerPlugin) handleTestAPI(args []string) error {
	// 方式 3: 使用版本检查获取 API
	api, err := p.ctx.GetPluginAPIWithVersion("example-api", sdk.PluginAPIVersion{
		Major: 0,
		Minor: 0,
		Patch: 1,
	})
	if err != nil {
		return fmt.Errorf("获取 API 失败: %w", err)
	}

	p.ctx.Logf("§a成功获取 example-api (版本检查通过)")

	// 使用类型断言调用方法
	// 注意: 实际使用时需要根据 API 插件的实际类型进行断言
	// 这里演示的是概念，实际需要导入 API 插件的包

	// 示例: 如果能导入 API 插件类型
	// if exampleAPI, ok := api.(*ExampleAPIPlugin); ok {
	//     message := exampleAPI.Greet("World")
	//     p.ctx.Logf("API 返回: %s", message)
	//
	//     exampleAPI.SendMessage("§e来自 API 消费者插件的消息")
	//
	//     status := exampleAPI.GetServerStatus()
	//     p.ctx.Logf("服务器状态: %+v", status)
	//
	//     distance := exampleAPI.CalculateDistance(0, 0, 0, 10, 10, 10)
	//     p.ctx.Logf("距离: %.2f", distance)
	// }

	// 如果无法导入类型，可以使用接口
	// 在实际项目中，API 提供者应该定义一个接口供消费者使用
	p.ctx.Logf("§a已获取 API 实例: %T", api)
	p.ctx.Logf("§e提示: 在实际使用中，需要通过类型断言或接口来调用具体方法")

	return nil
}

// Start 插件启动
func (p *APIConsumerPlugin) Start() error {
	p.ctx.Logf("API 消费者插件已启动")
	return nil
}

// Stop 插件停止
func (p *APIConsumerPlugin) Stop() error {
	p.ctx.Logf("API 消费者插件已停止")
	return nil
}

// NewPlugin 插件入口点
func NewPlugin() sdk.Plugin {
	return &APIConsumerPlugin{}
}
