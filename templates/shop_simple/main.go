package main

import (
	"fmt"
	sdk "github.com/maoqijie/FIN-plugin/sdk"
)

type ShopPlugin struct {
	ctx *sdk.Context
}

func (p *ShopPlugin) Init(ctx *sdk.Context) error {
	p.ctx = ctx
	ctx.Logf("商店插件初始化完成")

	// 注册聊天监听
	ctx.ListenChat(func(event sdk.ChatEvent) {
		ctx.Logf("收到消息: %s 说 %s", event.Sender, event.Message)

		if event.Message == "test" {
			ctx.GameUtils().SayTo(event.Sender, "§a商店插件收到你的消息！")
		}
	})

	return nil
}

func (p *ShopPlugin) Start() error {
	p.ctx.Logf("商店插件启动完成")
	return nil
}

func (p *ShopPlugin) Stop() error {
	p.ctx.Logf("商店插件已停止")
	return nil
}

func NewPlugin() sdk.Plugin {
	return &ShopPlugin{}
}
