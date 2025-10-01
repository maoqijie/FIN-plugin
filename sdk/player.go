package sdk

import (
	"fmt"
	"sync"
)

// Player 玩家信息对象，类似 ToolDelta 的 Player
type Player struct {
	Name            string // 玩家名称
	UUID            string // 玩家 UUID
	XUID            string // 玩家 XUID
	EntityUniqueID  int64  // 实体唯一 ID
	EntityRuntimeID uint64 // 实体运行时 ID
	Online          bool   // 是否在线

	// 内部使用
	gameUtils *GameUtils
}

// PlayerManager 玩家信息管理器，类似 ToolDelta 的 PlayerInfoMaintainer
type PlayerManager struct {
	mu              sync.RWMutex
	players         map[string]*Player // key: 玩家名称
	playersByUUID   map[string]*Player // key: UUID
	playersByUniqueID map[int64]*Player // key: EntityUniqueID
	botInfo         *Player
	gameUtils       *GameUtils
}

// NewPlayerManager 创建玩家管理器
func NewPlayerManager(gameUtils *GameUtils) *PlayerManager {
	return &PlayerManager{
		players:           make(map[string]*Player),
		playersByUUID:     make(map[string]*Player),
		playersByUniqueID: make(map[int64]*Player),
		gameUtils:         gameUtils,
	}
}

// AddPlayer 添加或更新玩家信息（内部方法）
func (pm *PlayerManager) AddPlayer(name, uuid, xuid string, uniqueID int64, runtimeID uint64) *Player {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	player := &Player{
		Name:            name,
		UUID:            uuid,
		XUID:            xuid,
		EntityUniqueID:  uniqueID,
		EntityRuntimeID: runtimeID,
		Online:          true,
		gameUtils:       pm.gameUtils,
	}

	pm.players[name] = player
	if uuid != "" {
		pm.playersByUUID[uuid] = player
	}
	if uniqueID != 0 {
		pm.playersByUniqueID[uniqueID] = player
	}

	return player
}

// RemovePlayer 移除玩家（内部方法）
func (pm *PlayerManager) RemovePlayer(name string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if player, exists := pm.players[name]; exists {
		player.Online = false
		delete(pm.players, name)
		if player.UUID != "" {
			delete(pm.playersByUUID, player.UUID)
		}
		if player.EntityUniqueID != 0 {
			delete(pm.playersByUniqueID, player.EntityUniqueID)
		}
	}
}

// SetBotInfo 设置机器人信息（内部方法）
func (pm *PlayerManager) SetBotInfo(name, uuid, xuid string, uniqueID int64, runtimeID uint64) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.botInfo = &Player{
		Name:            name,
		UUID:            uuid,
		XUID:            xuid,
		EntityUniqueID:  uniqueID,
		EntityRuntimeID: runtimeID,
		Online:          true,
		gameUtils:       pm.gameUtils,
	}
}

// GetAllPlayers 获取所有在线玩家列表
//
// 返回: 玩家对象列表
//
// 示例:
//   players := pm.GetAllPlayers()
//   for _, player := range players {
//       ctx.Logf("玩家: %s (在线: %v)", player.Name, player.Online)
//   }
func (pm *PlayerManager) GetAllPlayers() []*Player {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	players := make([]*Player, 0, len(pm.players))
	for _, player := range pm.players {
		players = append(players, player)
	}
	return players
}

// GetBotInfo 获取机器人信息
//
// 返回: 机器人玩家对象，如果未设置则返回 nil
//
// 示例:
//   bot := pm.GetBotInfo()
//   if bot != nil {
//       ctx.Logf("机器人名称: %s", bot.Name)
//   }
func (pm *PlayerManager) GetBotInfo() *Player {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.botInfo
}

// GetPlayerByName 根据玩家名称获取玩家对象
//
// name: 玩家名称
// 返回: 玩家对象，如果不存在则返回 nil
//
// 示例:
//   player := pm.GetPlayerByName("Steve")
//   if player != nil {
//       ctx.Logf("找到玩家: %s", player.Name)
//   }
func (pm *PlayerManager) GetPlayerByName(name string) *Player {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.players[name]
}

// GetPlayerByUUID 根据 UUID 获取玩家对象
//
// uuid: 玩家 UUID
// 返回: 玩家对象，如果不存在则返回 nil
//
// 示例:
//   player := pm.GetPlayerByUUID("123e4567-e89b-12d3-a456-426614174000")
func (pm *PlayerManager) GetPlayerByUUID(uuid string) *Player {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.playersByUUID[uuid]
}

// GetPlayerByUniqueID 根据实体唯一 ID 获取玩家对象
//
// uniqueID: 实体唯一 ID
// 返回: 玩家对象，如果不存在则返回 nil
//
// 示例:
//   player := pm.GetPlayerByUniqueID(1234567890)
func (pm *PlayerManager) GetPlayerByUniqueID(uniqueID int64) *Player {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.playersByUniqueID[uniqueID]
}

// GetPlayerCount 获取在线玩家数量
func (pm *PlayerManager) GetPlayerCount() int {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return len(pm.players)
}

// Player 方法

// Show 向玩家发送聊天消息
//
// text: 要发送的消息
//
// 示例:
//   player.Show("欢迎来到服务器！")
//   player.Show("§a你的分数: §e100")
func (p *Player) Show(text string) error {
	if p.gameUtils == nil {
		return fmt.Errorf("GameUtils 未初始化")
	}
	return p.gameUtils.Tellraw(p.Name, text)
}

// SetTitle 向玩家显示标题消息
//
// title: 主标题
// subtitle: 副标题（可选）
//
// 示例:
//   player.SetTitle("欢迎", "Welcome to Server")
//   player.SetTitle("游戏开始", "")
func (p *Player) SetTitle(title string, subtitle ...string) error {
	if p.gameUtils == nil {
		return fmt.Errorf("GameUtils 未初始化")
	}

	// 构建 title 命令
	var cmd string
	if len(subtitle) > 0 && subtitle[0] != "" {
		// 先发送副标题，再发送主标题
		p.gameUtils.SendCommand(fmt.Sprintf("title %s subtitle %s", p.Name, subtitle[0]))
		cmd = fmt.Sprintf("title %s title %s", p.Name, title)
	} else {
		cmd = fmt.Sprintf("title %s title %s", p.Name, title)
	}

	return p.gameUtils.SendCommand(cmd)
}

// SetActionBar 向玩家显示 ActionBar 消息
//
// text: 要显示的文本
//
// 示例:
//   player.SetActionBar("当前血量: 20/20")
func (p *Player) SetActionBar(text string) error {
	if p.gameUtils == nil {
		return fmt.Errorf("GameUtils 未初始化")
	}
	return p.gameUtils.Title(text)
}

// GetPos 获取玩家坐标
//
// 返回: 坐标信息、错误
//
// 示例:
//   pos, err := player.GetPos()
//   if err == nil {
//       ctx.Logf("玩家位置: %.2f, %.2f, %.2f", pos.X, pos.Y, pos.Z)
//   }
func (p *Player) GetPos() (*Position, error) {
	if p.gameUtils == nil {
		return nil, fmt.Errorf("GameUtils 未初始化")
	}
	return p.gameUtils.GetPos(p.Name)
}

// GetPosXYZ 获取玩家简单坐标
//
// 返回: x, y, z 坐标、错误
//
// 示例:
//   x, y, z, err := player.GetPosXYZ()
func (p *Player) GetPosXYZ() (float32, float32, float32, error) {
	if p.gameUtils == nil {
		return 0, 0, 0, fmt.Errorf("GameUtils 未初始化")
	}
	return p.gameUtils.GetPosXYZ(p.Name)
}

// GetScore 获取玩家在指定计分板的分数
//
// scbName: 计分板名称
// timeout: 超时时间（秒），默认 30 秒
//
// 返回: 分数、错误
//
// 示例:
//   score, err := player.GetScore("money", 10.0)
//   if err == nil {
//       ctx.Logf("玩家金币: %d", score)
//   }
func (p *Player) GetScore(scbName string, timeout ...float64) (int, error) {
	if p.gameUtils == nil {
		return 0, fmt.Errorf("GameUtils 未初始化")
	}

	t := 30.0
	if len(timeout) > 0 {
		t = timeout[0]
	}

	return p.gameUtils.GetScore(scbName, p.Name, t)
}

// GetItemCount 获取玩家背包中指定物品的数量
//
// itemName: 物品名称（如 "diamond"）
// itemSpecialID: 物品特殊 ID（默认 0）
//
// 返回: 物品数量、错误
//
// 示例:
//   count, err := player.GetItemCount("diamond", 0)
//   if err == nil {
//       ctx.Logf("玩家拥有 %d 个钻石", count)
//   }
func (p *Player) GetItemCount(itemName string, itemSpecialID ...int) (int, error) {
	if p.gameUtils == nil {
		return 0, fmt.Errorf("GameUtils 未初始化")
	}

	sid := 0
	if len(itemSpecialID) > 0 {
		sid = itemSpecialID[0]
	}

	return p.gameUtils.GetItem(p.Name, itemName, sid)
}

// IsOp 检查玩家是否有管理员权限
//
// 返回: 是否为管理员、错误
//
// 示例:
//   isOp, err := player.IsOp()
//   if err == nil && isOp {
//       ctx.Logf("%s 是管理员", player.Name)
//   }
func (p *Player) IsOp() (bool, error) {
	if p.gameUtils == nil {
		return false, fmt.Errorf("GameUtils 未初始化")
	}
	return p.gameUtils.IsOp(p.Name)
}

// Teleport 传送玩家到指定坐标
//
// x, y, z: 目标坐标
//
// 示例:
//   player.Teleport(100, 64, 200)
func (p *Player) Teleport(x, y, z float32) error {
	if p.gameUtils == nil {
		return fmt.Errorf("GameUtils 未初始化")
	}
	cmd := fmt.Sprintf("tp %s %.2f %.2f %.2f", p.Name, x, y, z)
	return p.gameUtils.SendCommand(cmd)
}

// TeleportTo 传送玩家到另一个玩家位置
//
// targetPlayer: 目标玩家名称
//
// 示例:
//   player.TeleportTo("Steve")
func (p *Player) TeleportTo(targetPlayer string) error {
	if p.gameUtils == nil {
		return fmt.Errorf("GameUtils 未初始化")
	}
	cmd := fmt.Sprintf("tp %s %s", p.Name, targetPlayer)
	return p.gameUtils.SendCommand(cmd)
}

// SetGameMode 设置玩家游戏模式
//
// mode: 游戏模式（0=生存, 1=创造, 2=冒险, 3=旁观）
//
// 示例:
//   player.SetGameMode(1) // 创造模式
func (p *Player) SetGameMode(mode int) error {
	if p.gameUtils == nil {
		return fmt.Errorf("GameUtils 未初始化")
	}
	cmd := fmt.Sprintf("gamemode %d %s", mode, p.Name)
	return p.gameUtils.SendCommand(cmd)
}

// GiveItem 给予玩家物品
//
// itemName: 物品名称
// amount: 数量
// data: 物品数据值（可选，默认 0）
//
// 示例:
//   player.GiveItem("diamond", 64, 0)
func (p *Player) GiveItem(itemName string, amount int, data ...int) error {
	if p.gameUtils == nil {
		return fmt.Errorf("GameUtils 未初始化")
	}

	dataValue := 0
	if len(data) > 0 {
		dataValue = data[0]
	}

	cmd := fmt.Sprintf("give %s %s %d %d", p.Name, itemName, amount, dataValue)
	return p.gameUtils.SendCommand(cmd)
}

// ClearItem 清除玩家物品
//
// itemName: 物品名称（可选，为空则清除所有）
// maxCount: 最大清除数量（可选，默认 -1 表示全部）
//
// 示例:
//   player.ClearItem("diamond", 10) // 清除 10 个钻石
//   player.ClearItem("", -1)        // 清除所有物品
func (p *Player) ClearItem(itemName string, maxCount ...int) error {
	if p.gameUtils == nil {
		return fmt.Errorf("GameUtils 未初始化")
	}

	count := -1
	if len(maxCount) > 0 {
		count = maxCount[0]
	}

	var cmd string
	if itemName == "" {
		cmd = fmt.Sprintf("clear %s", p.Name)
	} else if count == -1 {
		cmd = fmt.Sprintf("clear %s %s", p.Name, itemName)
	} else {
		cmd = fmt.Sprintf("clear %s %s %d", p.Name, itemName, count)
	}

	return p.gameUtils.SendCommand(cmd)
}

// AddEffect 给予玩家药水效果
//
// effect: 效果名称（如 "speed", "regeneration"）
// duration: 持续时间（秒）
// amplifier: 效果等级（0 = 等级 I，1 = 等级 II）
// hideParticles: 是否隐藏粒子效果
//
// 示例:
//   player.AddEffect("speed", 30, 1, false) // 速度 II，持续 30 秒
func (p *Player) AddEffect(effect string, duration int, amplifier int, hideParticles bool) error {
	if p.gameUtils == nil {
		return fmt.Errorf("GameUtils 未初始化")
	}

	hideFlag := "false"
	if hideParticles {
		hideFlag = "true"
	}

	cmd := fmt.Sprintf("effect %s %s %d %d %s", p.Name, effect, duration, amplifier, hideFlag)
	return p.gameUtils.SendCommand(cmd)
}

// ClearEffects 清除玩家所有药水效果
//
// 示例:
//   player.ClearEffects()
func (p *Player) ClearEffects() error {
	if p.gameUtils == nil {
		return fmt.Errorf("GameUtils 未初始化")
	}
	cmd := fmt.Sprintf("effect %s clear", p.Name)
	return p.gameUtils.SendCommand(cmd)
}

// Kill 杀死玩家
//
// 示例:
//   player.Kill()
func (p *Player) Kill() error {
	if p.gameUtils == nil {
		return fmt.Errorf("GameUtils 未初始化")
	}
	cmd := fmt.Sprintf("kill %s", p.Name)
	return p.gameUtils.SendCommand(cmd)
}

// Kick 踢出玩家
//
// reason: 踢出原因（可选）
//
// 示例:
//   player.Kick("违反服务器规则")
func (p *Player) Kick(reason ...string) error {
	if p.gameUtils == nil {
		return fmt.Errorf("GameUtils 未初始化")
	}

	var cmd string
	if len(reason) > 0 && reason[0] != "" {
		cmd = fmt.Sprintf("kick %s %s", p.Name, reason[0])
	} else {
		cmd = fmt.Sprintf("kick %s", p.Name)
	}

	return p.gameUtils.SendCommand(cmd)
}
