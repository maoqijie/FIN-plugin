package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	sdk "github.com/maoqijie/FIN-plugin/sdk"
)

type ShopPlugin struct {
	ctx    *sdk.Context
	config *ShopConfig
	// 存储等待玩家输入的状态
	waitingForInput map[string]*PlayerInputState
}

// ShopConfig 商店配置
type ShopConfig struct {
	ScoreboardName string                `json:"货币计分板名"`
	SellItems      map[string]*ShopItem  `json:"出售"`
	BuybackItems   map[string]*ShopItem  `json:"收购"`
}

// ShopItem 商店物品
type ShopItem struct {
	ID    string `json:"ID"`    // 物品标识符（如 minecraft:diamond）
	Price int    `json:"价格"`  // 单价
}

// PlayerInputState 玩家输入状态
type PlayerInputState struct {
	Action     string    // "buy" 或 "sell"
	ItemName   string    // 物品名称
	Item       *ShopItem // 物品信息
	WaitingFor string    // "amount" 表示等待输入数量
	Timestamp  time.Time // 超时控制
}

func (p *ShopPlugin) Init(ctx *sdk.Context) error {
	p.ctx = ctx
	p.waitingForInput = make(map[string]*PlayerInputState)

	// 加载配置
	defaultConfig := &ShopConfig{
		ScoreboardName: "money",
		SellItems: map[string]*ShopItem{
			"钻石": {
				ID:    "minecraft:diamond",
				Price: 100,
			},
			"铁锭": {
				ID:    "minecraft:iron_ingot",
				Price: 10,
			},
			"金锭": {
				ID:    "minecraft:gold_ingot",
				Price: 50,
			},
			"绿宝石": {
				ID:    "minecraft:emerald",
				Price: 200,
			},
		},
		BuybackItems: map[string]*ShopItem{
			"钻石": {
				ID:    "minecraft:diamond",
				Price: 80,
			},
			"铁锭": {
				ID:    "minecraft:iron_ingot",
				Price: 8,
			},
			"金锭": {
				ID:    "minecraft:gold_ingot",
				Price: 40,
			},
			"绿宝石": {
				ID:    "minecraft:emerald",
				Price: 160,
			},
		},
	}

	config, err := p.ctx.Config().GetConfig("config.json", defaultConfig)
	if err != nil {
		return fmt.Errorf("加载配置失败: %v", err)
	}

	p.config = config.(*ShopConfig)

	// 监听聊天消息
	ctx.ListenChat(p.onChat)

	p.ctx.Logf("商店插件已加载，货币计分板: %s", p.config.ScoreboardName)
	return nil
}

func (p *ShopPlugin) Start() error {
	p.ctx.Logf("商店插件已启动")

	// 启动定时清理超时的输入状态
	p.ctx.Utils().NewTimer(30, func() {
		p.cleanupExpiredStates()
	})

	return nil
}

func (p *ShopPlugin) Stop() error {
	p.ctx.Logf("商店插件已停止")
	return nil
}

func (p *ShopPlugin) onChat(event sdk.ChatEvent) {
	playerName := event.PlayerName
	message := strings.TrimSpace(event.Message)

	// 检查是否有等待输入的状态
	if state, exists := p.waitingForInput[playerName]; exists {
		if state.WaitingFor == "amount" {
			p.handleAmountInput(playerName, message, state)
			return
		}
	}

	// 处理商店命令
	switch message {
	case "购买", "buy":
		p.showShop(playerName, "buy")
	case "收购", "sell":
		p.showShop(playerName, "sell")
	default:
		// 检查是否是选择物品
		if state, exists := p.waitingForInput[playerName]; exists {
			if state.WaitingFor == "item" {
				p.handleItemSelection(playerName, message, state)
			}
		}
	}
}

// showShop 显示商店列表
func (p *ShopPlugin) showShop(playerName string, action string) {
	var items map[string]*ShopItem
	var title string

	if action == "buy" {
		items = p.config.SellItems
		title = "§a=== 商店购买列表 ==="
	} else {
		items = p.config.BuybackItems
		title = "§e=== 商店收购列表 ==="
	}

	// 发送标题
	p.ctx.GameUtils().SayTo(playerName, title)

	// 列出所有物品
	index := 1
	for itemName, item := range items {
		msg := fmt.Sprintf("§f%d. §b%s §f- §6%d §f积分", index, itemName, item.Price)
		p.ctx.GameUtils().SayTo(playerName, msg)
		index++
	}

	p.ctx.GameUtils().SayTo(playerName, "§7请输入物品名称进行交易")

	// 设置等待状态
	p.waitingForInput[playerName] = &PlayerInputState{
		Action:     action,
		WaitingFor: "item",
		Timestamp:  time.Now(),
	}
}

// handleItemSelection 处理物品选择
func (p *ShopPlugin) handleItemSelection(playerName string, input string, state *PlayerInputState) {
	var items map[string]*ShopItem
	if state.Action == "buy" {
		items = p.config.SellItems
	} else {
		items = p.config.BuybackItems
	}

	// 查找物品
	item, exists := items[input]
	if !exists {
		p.ctx.GameUtils().SayTo(playerName, "§c物品不存在，请重新输入")
		return
	}

	// 更新状态
	state.ItemName = input
	state.Item = item
	state.WaitingFor = "amount"
	state.Timestamp = time.Now()

	// 询问数量
	if state.Action == "buy" {
		p.ctx.GameUtils().SayTo(playerName, fmt.Sprintf("§a你要购买多少个 §b%s §a? (单价: §6%d §a积分)", input, item.Price))
	} else {
		p.ctx.GameUtils().SayTo(playerName, fmt.Sprintf("§e你要出售多少个 §b%s §e? (单价: §6%d §e积分)", input, item.Price))
	}
}

// handleAmountInput 处理数量输入
func (p *ShopPlugin) handleAmountInput(playerName string, input string, state *PlayerInputState) {
	// 解析数量
	amount, err := strconv.Atoi(input)
	if err != nil || amount <= 0 {
		p.ctx.GameUtils().SayTo(playerName, "§c请输入有效的数量（正整数）")
		return
	}

	// 执行交易
	if state.Action == "buy" {
		p.executeBuy(playerName, state.ItemName, state.Item, amount)
	} else {
		p.executeSell(playerName, state.ItemName, state.Item, amount)
	}

	// 清除等待状态
	delete(p.waitingForInput, playerName)
}

// executeBuy 执行购买
func (p *ShopPlugin) executeBuy(playerName string, itemName string, item *ShopItem, amount int) {
	totalPrice := item.Price * amount

	// 获取玩家余额
	balance, err := p.ctx.GameUtils().GetScore(p.config.ScoreboardName, playerName, 5.0)
	if err != nil {
		p.ctx.GameUtils().SayTo(playerName, "§c无法获取你的余额")
		p.ctx.Logf("获取余额失败: %v", err)
		return
	}

	// 检查余额
	if balance < totalPrice {
		p.ctx.GameUtils().SayTo(playerName, fmt.Sprintf("§c余额不足！需要 §6%d §c积分，你只有 §6%d §c积分", totalPrice, balance))
		return
	}

	// 扣除积分
	cmd := fmt.Sprintf("scoreboard players remove \"%s\" %s %d", playerName, p.config.ScoreboardName, totalPrice)
	p.ctx.GameUtils().SendCommand(cmd)

	// 给予物品
	giveCmd := fmt.Sprintf("give \"%s\" %s %d", playerName, item.ID, amount)
	p.ctx.GameUtils().SendCommand(giveCmd)

	// 发送成功消息
	successMsg := fmt.Sprintf("§a购买成功！花费 §6%d §a积分购买了 §b%d §a个 §b%s", totalPrice, amount, itemName)
	p.ctx.GameUtils().SayTo(playerName, successMsg)

	// 显示剩余余额
	newBalance := balance - totalPrice
	p.ctx.GameUtils().SayTo(playerName, fmt.Sprintf("§7剩余余额: §6%d §7积分", newBalance))

	p.ctx.Logf("%s 购买了 %d 个 %s，花费 %d 积分", playerName, amount, itemName, totalPrice)
}

// executeSell 执行出售
func (p *ShopPlugin) executeSell(playerName string, itemName string, item *ShopItem, amount int) {
	totalPrice := item.Price * amount

	// 获取玩家物品数量
	itemCount, err := p.ctx.GameUtils().GetItem(playerName, item.ID, 0)
	if err != nil {
		p.ctx.GameUtils().SayTo(playerName, "§c无法获取你的物品数量")
		p.ctx.Logf("获取物品数量失败: %v", err)
		return
	}

	// 检查物品数量
	if itemCount < amount {
		p.ctx.GameUtils().SayTo(playerName, fmt.Sprintf("§c物品不足！需要 §b%d §c个，你只有 §b%d §c个", amount, itemCount))
		return
	}

	// 清除物品
	clearCmd := fmt.Sprintf("clear \"%s\" %s 0 %d", playerName, item.ID, amount)
	p.ctx.GameUtils().SendCommand(clearCmd)

	// 增加积分
	addCmd := fmt.Sprintf("scoreboard players add \"%s\" %s %d", playerName, p.config.ScoreboardName, totalPrice)
	p.ctx.GameUtils().SendCommand(addCmd)

	// 发送成功消息
	successMsg := fmt.Sprintf("§a出售成功！出售了 §b%d §a个 §b%s§a，获得 §6%d §a积分", amount, itemName, totalPrice)
	p.ctx.GameUtils().SayTo(playerName, successMsg)

	// 获取新余额
	p.ctx.Utils().Sleep(0.5) // 等待命令执行
	balance, err := p.ctx.GameUtils().GetScore(p.config.ScoreboardName, playerName, 5.0)
	if err == nil {
		p.ctx.GameUtils().SayTo(playerName, fmt.Sprintf("§7当前余额: §6%d §7积分", balance))
	}

	p.ctx.Logf("%s 出售了 %d 个 %s，获得 %d 积分", playerName, amount, itemName, totalPrice)
}

// cleanupExpiredStates 清理超时的输入状态
func (p *ShopPlugin) cleanupExpiredStates() {
	now := time.Now()
	for playerName, state := range p.waitingForInput {
		if now.Sub(state.Timestamp) > 60*time.Second {
			delete(p.waitingForInput, playerName)
			p.ctx.GameUtils().SayTo(playerName, "§c商店操作已超时，请重新开始")
		}
	}
}

func NewPlugin() sdk.Plugin {
	return &ShopPlugin{}
}
