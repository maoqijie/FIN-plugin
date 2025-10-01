package main

import (
	"fmt"

	sdk "github.com/maoqijie/FIN-plugin/sdk"
)

// ExampleAPIPlugin 示例 API 插件（前置插件）
// 此插件提供 API 供其他插件调用
type ExampleAPIPlugin struct {
	ctx *sdk.Context
}

// Init 插件初始化
func (p *ExampleAPIPlugin) Init(ctx *sdk.Context) error {
	p.ctx = ctx

	// 注册为 API 插件，其他插件可以通过 "example-api" 名称获取此插件实例
	err := ctx.RegisterPluginAPI("example-api", sdk.PluginAPIVersion{
		Major: 0,
		Minor: 0,
		Patch: 1,
	}, p)
	if err != nil {
		return fmt.Errorf("注册 API 插件失败: %w", err)
	}

	ctx.Logf("示例 API 插件已注册，版本 0.0.1")
	return nil
}

// Start 插件启动
func (p *ExampleAPIPlugin) Start() error {
	p.ctx.Logf("示例 API 插件已启动")
	return nil
}

// Stop 插件停止
func (p *ExampleAPIPlugin) Stop() error {
	p.ctx.Logf("示例 API 插件已停止")
	return nil
}

// ===== 以下是提供给其他插件的公共 API 方法 =====

// Greet 向指定对象问候（示例方法）
func (p *ExampleAPIPlugin) Greet(name string) string {
	message := fmt.Sprintf("Hello, %s!", name)
	p.ctx.Logf("Greet 被调用: %s", message)
	return message
}

// SendMessage 向所有玩家发送消息（示例方法）
func (p *ExampleAPIPlugin) SendMessage(message string) error {
	utils := p.ctx.GameUtils()
	if utils == nil {
		return fmt.Errorf("GameUtils 未初始化")
	}
	return utils.SayTo("@a", message)
}

// GetServerStatus 获取服务器状态信息（示例方法）
func (p *ExampleAPIPlugin) GetServerStatus() map[string]interface{} {
	serverInfo := p.ctx.ServerInfo()
	botInfo := p.ctx.BotInfo()

	return map[string]interface{}{
		"bot_name":     botInfo.Name,
		"server_code":  serverInfo.Code,
		"has_passcode": serverInfo.PasscodeSet,
	}
}

// CalculateDistance 计算两点之间的距离（示例方法）
func (p *ExampleAPIPlugin) CalculateDistance(x1, y1, z1, x2, y2, z2 float32) float32 {
	dx := x2 - x1
	dy := y2 - y1
	dz := z2 - z1
	return float32(sqrt(float64(dx*dx + dy*dy + dz*dz)))
}

// sqrt 计算平方根
func sqrt(x float64) float64 {
	if x == 0 {
		return 0
	}
	z := x
	for i := 0; i < 10; i++ {
		z = (z + x/z) / 2
	}
	return z
}

// NewPlugin 插件入口点
func NewPlugin() sdk.Plugin {
	return &ExampleAPIPlugin{}
}
