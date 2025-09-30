package main

import (
	"fmt"
	"strings"

	"github.com/FunInterwork/PluginFramework/sdk"
)

// InfoPlugin 演示如何注册控制台命令与读取上下文信息。
type InfoPlugin struct {
	ctx *sdk.Context
}

func (p *InfoPlugin) Init(ctx *sdk.Context) error {
	p.ctx = ctx

	return ctx.RegisterConsoleCommand(sdk.ConsoleCommand{
		Triggers:     []string{"info", "botinfo"},
		ArgumentHint: "[详细]",
		Usage:        "查看机器人与服务器运行状态",
		Description:  "输出机器人、租赁服与互通配置的实时信息",
		Handler:      p.handleInfoCommand,
	})
}

func (p *InfoPlugin) Start() error {
	p.ctx.Logf("Info 插件已启动")
	return nil
}

func (p *InfoPlugin) Stop() error {
	p.ctx.Logf("Info 插件已停止")
	return nil
}

func (p *InfoPlugin) handleInfoCommand(args []string) error {
	bot := p.ctx.BotInfo()
	fmt.Printf("机器人昵称: %s\n", bot.Name)

	server := p.ctx.ServerInfo()
	fmt.Printf("租赁服号: %s\n", server.Code)

	if len(args) > 0 && strings.EqualFold(args[0], "详细") {
		interwork := p.ctx.InterworkInfo()
		fmt.Printf("已关联群组: %d 个\n", len(interwork.LinkedGroups))
	}

	return nil
}

// NewPlugin 由主程序调用以获取插件实例。
func NewPlugin() sdk.Plugin {
	return &InfoPlugin{}
}

