package main

import (
	sdk "github.com/maoqijie/FIN-plugin/sdk"
)

type ShopPlugin struct {
	ctx *sdk.Context
}

func (p *ShopPlugin) Init(ctx *sdk.Context) error {
	p.ctx = ctx
	return nil
}

func (p *ShopPlugin) Start() error {
	return nil
}

func (p *ShopPlugin) Stop() error {
	return nil
}

func NewPlugin() sdk.Plugin {
	return &ShopPlugin{}
}
