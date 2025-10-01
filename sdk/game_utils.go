package sdk

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"
)

// Position 表示三维坐标
type Position struct {
	X         float32
	Y         float32
	Z         float32
	Dimension uint8
	YRot      float32
}

// InventorySlot 表示背包中的一个物品槽位
type InventorySlot struct {
	Slot     int    // 槽位编号
	ItemID   string // 物品 ID (如 "minecraft:diamond")
	ItemName string // 物品名称
	Count    int    // 物品数量
	Aux      int    // 物品附加值（如工具耐久度）
}

// GameUtils 提供高级游戏交互接口，类似 ToolDelta 的 game_utils
type GameUtils struct {
	gi interface{} // 存储 *game_interface.GameInterface
}

// NewGameUtils 创建 GameUtils 实例
func NewGameUtils(gi interface{}) *GameUtils {
	return &GameUtils{gi: gi}
}

// GetTarget 获取匹配目标选择器的玩家名称列表
// target: 目标选择器（如 "@a", "@p", "PlayerName"）
// timeout: 超时时间（秒），默认 5 秒
func (g *GameUtils) GetTarget(target string, timeout float64) ([]string, error) {
	if timeout <= 0 {
		timeout = 5.0
	}

	// 使用反射调用 gi.Querytarget().DoQuerytarget(target)
	giVal := reflect.ValueOf(g.gi)
	if !giVal.IsValid() || giVal.IsNil() {
		return nil, fmt.Errorf("gameInterface 未初始化")
	}

	querytargetMethod := giVal.MethodByName("Querytarget")
	if !querytargetMethod.IsValid() {
		return nil, fmt.Errorf("gameInterface 不支持 Querytarget 方法")
	}

	querytargetVal := querytargetMethod.Call(nil)
	if len(querytargetVal) == 0 || !querytargetVal[0].IsValid() {
		return nil, fmt.Errorf("Querytarget 返回值无效")
	}

	doQueryMethod := querytargetVal[0].MethodByName("DoQuerytarget")
	if !doQueryMethod.IsValid() {
		return nil, fmt.Errorf("Querytarget 不支持 DoQuerytarget 方法")
	}

	results := doQueryMethod.Call([]reflect.Value{reflect.ValueOf(target)})
	if len(results) != 2 {
		return nil, fmt.Errorf("DoQuerytarget 返回值数量不正确")
	}

	if !results[1].IsNil() {
		return nil, fmt.Errorf("查询目标失败: %v", results[1].Interface())
	}

	// 解析返回的切片
	if results[0].Kind() != reflect.Slice {
		return []string{}, nil
	}

	names := []string{}
	for i := 0; i < results[0].Len(); i++ {
		item := results[0].Index(i)
		// 如果是指针，需要解引用
		if item.Kind() == reflect.Ptr {
			item = item.Elem()
		}

		// 提取 EntityName 字段
		nameField := item.FieldByName("EntityName")
		if nameField.IsValid() && nameField.Kind() == reflect.String {
			name := nameField.String()
			if name != "" {
				names = append(names, name)
			}
		}
	}

	return names, nil
}

// GetPos 获取玩家的详细坐标信息
// target: 目标玩家名称或选择器
// 返回: Position 包含坐标、维度、视角等信息
func (g *GameUtils) GetPos(target string) (*Position, error) {
	giVal := reflect.ValueOf(g.gi)
	if !giVal.IsValid() || giVal.IsNil() {
		return nil, fmt.Errorf("gameInterface 未初始化")
	}

	querytargetMethod := giVal.MethodByName("Querytarget")
	if !querytargetMethod.IsValid() {
		return nil, fmt.Errorf("gameInterface 不支持 Querytarget 方法")
	}

	querytargetVal := querytargetMethod.Call(nil)
	if len(querytargetVal) == 0 || !querytargetVal[0].IsValid() {
		return nil, fmt.Errorf("Querytarget 返回值无效")
	}

	doQueryMethod := querytargetVal[0].MethodByName("DoQuerytarget")
	if !doQueryMethod.IsValid() {
		return nil, fmt.Errorf("Querytarget 不支持 DoQuerytarget 方法")
	}

	results := doQueryMethod.Call([]reflect.Value{reflect.ValueOf(target)})
	if len(results) != 2 {
		return nil, fmt.Errorf("DoQuerytarget 返回值数量不正确")
	}

	if !results[1].IsNil() {
		return nil, fmt.Errorf("查询玩家坐标失败: %v", results[1].Interface())
	}

	if results[0].Kind() != reflect.Slice || results[0].Len() == 0 {
		return nil, fmt.Errorf("玩家不存在或不在线")
	}

	// 解析第一个结果
	firstResult := results[0].Index(0)
	// 如果是指针，需要解引用
	if firstResult.Kind() == reflect.Ptr {
		firstResult = firstResult.Elem()
	}

	// 提取 Position 字段
	posField := firstResult.FieldByName("Position")
	if !posField.IsValid() {
		return nil, fmt.Errorf("无法提取 Position 字段")
	}

	x := float32(posField.FieldByName("X").Float())
	y := float32(posField.FieldByName("Y").Float())
	z := float32(posField.FieldByName("Z").Float())
	dimension := uint8(firstResult.FieldByName("Dimension").Uint())
	yRot := float32(firstResult.FieldByName("YRot").Float())

	return &Position{
		X:         x,
		Y:         y,
		Z:         z,
		Dimension: dimension,
		YRot:      yRot,
	}, nil
}

// GetPosXYZ 获取玩家的简单坐标值
// target: 目标玩家名称或选择器
// 返回: (x, y, z) 坐标元组
func (g *GameUtils) GetPosXYZ(target string) (float32, float32, float32, error) {
	pos, err := g.GetPos(target)
	if err != nil {
		return 0, 0, 0, err
	}
	return pos.X, pos.Y, pos.Z, nil
}

// GetItem 统计玩家背包中特定物品的数量
// target: 目标玩家名称或选择器
// itemName: 物品的 Minecraft ID（如 "minecraft:diamond"）
// itemSpecialID: 物品特殊 ID（默认 -1 表示忽略）
func (g *GameUtils) GetItem(target, itemName string, itemSpecialID int) (int, error) {
	giVal := reflect.ValueOf(g.gi)
	if !giVal.IsValid() || giVal.IsNil() {
		return 0, fmt.Errorf("gameInterface 未初始化")
	}

	commandsMethod := giVal.MethodByName("Commands")
	if !commandsMethod.IsValid() {
		return 0, fmt.Errorf("gameInterface 不支持 Commands 方法")
	}

	commandsVal := commandsMethod.Call(nil)
	if len(commandsVal) == 0 || !commandsVal[0].IsValid() {
		return 0, fmt.Errorf("Commands 返回值无效")
	}

	// 使用 clear 命令的测试模式来统计物品数量
	cmd := fmt.Sprintf("clear %s %s %d 0", target, itemName, itemSpecialID)
	sendMethod := commandsVal[0].MethodByName("SendWSCommandWithResp")
	if !sendMethod.IsValid() {
		return 0, fmt.Errorf("Commands 不支持 SendWSCommandWithResp 方法")
	}

	results := sendMethod.Call([]reflect.Value{reflect.ValueOf(cmd)})
	if len(results) != 2 {
		return 0, fmt.Errorf("SendWSCommandWithResp 返回值数量不正确")
	}

	if !results[1].IsNil() {
		return 0, fmt.Errorf("查询物品数量失败: %v", results[1].Interface())
	}

	// 从 CommandOutput 提取 SuccessCount
	output := results[0]
	if output.Kind() == reflect.Ptr {
		output = output.Elem()
	}
	successCount := output.FieldByName("SuccessCount").Int()

	return int(successCount), nil
}

// GetScore 获取计分板中目标的分数
// scbName: 计分板名称
// target: 目标名称
// timeout: 超时时间（秒），默认 30 秒
func (g *GameUtils) GetScore(scbName, target string, timeout float64) (int, error) {
	if timeout <= 0 {
		timeout = 30.0
	}

	giVal := reflect.ValueOf(g.gi)
	if !giVal.IsValid() || giVal.IsNil() {
		return 0, fmt.Errorf("gameInterface 未初始化")
	}

	commandsMethod := giVal.MethodByName("Commands")
	if !commandsMethod.IsValid() {
		return 0, fmt.Errorf("gameInterface 不支持 Commands 方法")
	}

	commandsVal := commandsMethod.Call(nil)
	if len(commandsVal) == 0 || !commandsVal[0].IsValid() {
		return 0, fmt.Errorf("Commands 返回值无效")
	}

	// 使用 scoreboard players test 命令获取分数
	cmd := fmt.Sprintf("scoreboard players test %s %s * *", target, scbName)
	timeoutDuration := time.Duration(timeout * float64(time.Second))

	sendMethod := commandsVal[0].MethodByName("SendWSCommandWithTimeout")
	if !sendMethod.IsValid() {
		return 0, fmt.Errorf("Commands 不支持 SendWSCommandWithTimeout 方法")
	}

	results := sendMethod.Call([]reflect.Value{
		reflect.ValueOf(cmd),
		reflect.ValueOf(timeoutDuration),
	})
	if len(results) != 3 {
		return 0, fmt.Errorf("SendWSCommandWithTimeout 返回值数量不正确")
	}

	// 检查是否超时
	if results[1].Bool() {
		return 0, fmt.Errorf("获取分数超时")
	}

	if !results[2].IsNil() {
		return 0, fmt.Errorf("计分板或目标不存在: %v", results[2].Interface())
	}

	// 从 CommandOutput.OutputMessages[0].Parameters[0] 提取分数
	output := results[0]
	if output.Kind() == reflect.Ptr {
		output = output.Elem()
	}
	outputMsgs := output.FieldByName("OutputMessages")
	if outputMsgs.Len() == 0 {
		return 0, fmt.Errorf("无法获取分数")
	}

	firstMsg := outputMsgs.Index(0)
	if firstMsg.Kind() == reflect.Ptr {
		firstMsg = firstMsg.Elem()
	}
	parameters := firstMsg.FieldByName("Parameters")
	if parameters.Len() == 0 {
		return 0, fmt.Errorf("无法获取分数参数")
	}

	// 通常分数在第一个参数中
	scoreStr := parameters.Index(0).String()
	var score int
	if _, err := fmt.Sscanf(scoreStr, "%d", &score); err != nil {
		return 0, fmt.Errorf("解析分数失败: %w", err)
	}

	return score, nil
}

// IsCmdSuccess 检查命令是否执行成功
// cmd: 要执行的 Minecraft 命令
// timeout: 超时时间（秒），默认 30 秒
func (g *GameUtils) IsCmdSuccess(cmd string, timeout float64) (bool, error) {
	if timeout <= 0 {
		timeout = 30.0
	}

	giVal := reflect.ValueOf(g.gi)
	if !giVal.IsValid() || giVal.IsNil() {
		return false, fmt.Errorf("gameInterface 未初始化")
	}

	commandsMethod := giVal.MethodByName("Commands")
	if !commandsMethod.IsValid() {
		return false, fmt.Errorf("gameInterface 不支持 Commands 方法")
	}

	commandsVal := commandsMethod.Call(nil)
	if len(commandsVal) == 0 || !commandsVal[0].IsValid() {
		return false, fmt.Errorf("Commands 返回值无效")
	}

	timeoutDuration := time.Duration(timeout * float64(time.Second))

	sendMethod := commandsVal[0].MethodByName("SendWSCommandWithTimeout")
	if !sendMethod.IsValid() {
		return false, fmt.Errorf("Commands 不支持 SendWSCommandWithTimeout 方法")
	}

	results := sendMethod.Call([]reflect.Value{
		reflect.ValueOf(cmd),
		reflect.ValueOf(timeoutDuration),
	})
	if len(results) != 3 {
		return false, fmt.Errorf("SendWSCommandWithTimeout 返回值数量不正确")
	}

	// 检查是否超时
	if results[1].Bool() {
		return false, fmt.Errorf("命令执行超时")
	}

	// 检查是否有错误
	if !results[2].IsNil() {
		return false, nil
	}

	// 检查 SuccessCount
	output := results[0]
	if output.Kind() == reflect.Ptr {
		output = output.Elem()
	}
	successCount := output.FieldByName("SuccessCount").Int()

	return successCount > 0, nil
}

// IsOp 检查玩家是否拥有管理员权限
// playerName: 玩家名称
func (g *GameUtils) IsOp(playerName string) (bool, error) {
	// 尝试执行一个需要 OP 权限的命令来判断
	// 使用 tag 命令测试权限（需要 OP 才能操作）
	testCmd := fmt.Sprintf("tag \"%s\" list", playerName)
	success, err := g.IsCmdSuccess(testCmd, 5.0)
	if err != nil {
		return false, fmt.Errorf("检查管理员权限失败: %w", err)
	}

	return success, nil
}

// TakeItemOutItemFrame 从展示框中取出物品
// x, y, z: 展示框的坐标
func (g *GameUtils) TakeItemOutItemFrame(x, y, z int) error {
	giVal := reflect.ValueOf(g.gi)
	if !giVal.IsValid() || giVal.IsNil() {
		return fmt.Errorf("gameInterface 未初始化")
	}

	commandsMethod := giVal.MethodByName("Commands")
	if !commandsMethod.IsValid() {
		return fmt.Errorf("gameInterface 不支持 Commands 方法")
	}

	commandsVal := commandsMethod.Call(nil)
	if len(commandsVal) == 0 || !commandsVal[0].IsValid() {
		return fmt.Errorf("Commands 返回值无效")
	}

	// 使用 kill 命令移除展示框中的物品实体
	cmd := fmt.Sprintf("kill @e[type=item_frame,x=%d,y=%d,z=%d,r=1]", x, y, z)

	sendMethod := commandsVal[0].MethodByName("SendWSCommandWithResp")
	if !sendMethod.IsValid() {
		return fmt.Errorf("Commands 不支持 SendWSCommandWithResp 方法")
	}

	results := sendMethod.Call([]reflect.Value{reflect.ValueOf(cmd)})
	if len(results) != 2 {
		return fmt.Errorf("SendWSCommandWithResp 返回值数量不正确")
	}

	if !results[1].IsNil() {
		return fmt.Errorf("移除展示框物品失败: %v", results[1].Interface())
	}

	return nil
}

// SendCommand 发送游戏命令（封装常用命令发送功能）
// cmd: Minecraft 命令
func (g *GameUtils) SendCommand(cmd string) error {
	giVal := reflect.ValueOf(g.gi)
	if !giVal.IsValid() || giVal.IsNil() {
		return fmt.Errorf("gameInterface 未初始化")
	}

	commandsMethod := giVal.MethodByName("Commands")
	if !commandsMethod.IsValid() {
		return fmt.Errorf("gameInterface 不支持 Commands 方法")
	}

	commandsVal := commandsMethod.Call(nil)
	if len(commandsVal) == 0 || !commandsVal[0].IsValid() {
		return fmt.Errorf("Commands 返回值无效")
	}

	sendMethod := commandsVal[0].MethodByName("SendWSCommand")
	if !sendMethod.IsValid() {
		return fmt.Errorf("Commands 不支持 SendWSCommand 方法")
	}

	results := sendMethod.Call([]reflect.Value{reflect.ValueOf(cmd)})
	if len(results) != 1 {
		return fmt.Errorf("SendWSCommand 返回值数量不正确")
	}

	if !results[0].IsNil() {
		return fmt.Errorf("发送命令失败: %v", results[0].Interface())
	}

	return nil
}

// SendChat 让机器人在聊天栏发言
// message: 聊天消息内容
func (g *GameUtils) SendChat(message string) error {
	giVal := reflect.ValueOf(g.gi)
	if !giVal.IsValid() || giVal.IsNil() {
		return fmt.Errorf("gameInterface 未初始化")
	}

	commandsMethod := giVal.MethodByName("Commands")
	if !commandsMethod.IsValid() {
		return fmt.Errorf("gameInterface 不支持 Commands 方法")
	}

	commandsVal := commandsMethod.Call(nil)
	if len(commandsVal) == 0 || !commandsVal[0].IsValid() {
		return fmt.Errorf("Commands 返回值无效")
	}

	sendChatMethod := commandsVal[0].MethodByName("SendChat")
	if !sendChatMethod.IsValid() {
		return fmt.Errorf("Commands 不支持 SendChat 方法")
	}

	results := sendChatMethod.Call([]reflect.Value{reflect.ValueOf(message)})
	if len(results) != 1 {
		return fmt.Errorf("SendChat 返回值数量不正确")
	}

	if !results[0].IsNil() {
		return fmt.Errorf("发送聊天消息失败: %v", results[0].Interface())
	}

	return nil
}

// Title 以 actionbar 形式向所有玩家显示消息
// message: 要显示的消息
func (g *GameUtils) Title(message string) error {
	giVal := reflect.ValueOf(g.gi)
	if !giVal.IsValid() || giVal.IsNil() {
		return fmt.Errorf("gameInterface 未初始化")
	}

	commandsMethod := giVal.MethodByName("Commands")
	if !commandsMethod.IsValid() {
		return fmt.Errorf("gameInterface 不支持 Commands 方法")
	}

	commandsVal := commandsMethod.Call(nil)
	if len(commandsVal) == 0 || !commandsVal[0].IsValid() {
		return fmt.Errorf("Commands 返回值无效")
	}

	titleMethod := commandsVal[0].MethodByName("Title")
	if !titleMethod.IsValid() {
		return fmt.Errorf("Commands 不支持 Title 方法")
	}

	results := titleMethod.Call([]reflect.Value{reflect.ValueOf(message)})
	if len(results) != 1 {
		return fmt.Errorf("Title 返回值数量不正确")
	}

	if !results[0].IsNil() {
		return fmt.Errorf("显示标题失败: %v", results[0].Interface())
	}

	return nil
}

// Tellraw 使用 tellraw 命令向指定玩家发送 JSON 格式消息
// selector: 玩家选择器（如 "@a", "@p", "PlayerName"）
// message: 消息内容（会自动包装为 rawtext 格式）
func (g *GameUtils) Tellraw(selector, message string) error {
	payload := map[string]interface{}{
		"rawtext": []map[string]string{
			{"text": message},
		},
	}
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("构造 tellraw 消息失败: %w", err)
	}

	cmd := fmt.Sprintf("tellraw %s %s", selector, string(jsonBytes))
	return g.SendCommand(cmd)
}

// SendCommandWithResponse 发送命令并等待响应
// cmd: Minecraft 命令
// timeout: 超时时间（秒），默认 30 秒
// 返回: 命令输出结果、是否超时、错误
//
// 示例:
//   output, timedOut, err := utils.SendCommandWithResponse("testfor @a", 10.0)
//   if err == nil && !timedOut {
//       // 处理输出
//   }
func (g *GameUtils) SendCommandWithResponse(cmd string, timeout ...float64) (interface{}, bool, error) {
	t := 30.0
	if len(timeout) > 0 {
		t = timeout[0]
	}

	giVal := reflect.ValueOf(g.gi)
	if !giVal.IsValid() || giVal.IsNil() {
		return nil, false, fmt.Errorf("gameInterface 未初始化")
	}

	commandsMethod := giVal.MethodByName("Commands")
	if !commandsMethod.IsValid() {
		return nil, false, fmt.Errorf("gameInterface 不支持 Commands 方法")
	}

	commandsVal := commandsMethod.Call(nil)
	if len(commandsVal) == 0 || !commandsVal[0].IsValid() {
		return nil, false, fmt.Errorf("Commands 返回值无效")
	}

	sendMethod := commandsVal[0].MethodByName("SendWSCommandWithTimeout")
	if !sendMethod.IsValid() {
		return nil, false, fmt.Errorf("Commands 不支持 SendWSCommandWithTimeout 方法")
	}

	timeoutDuration := time.Duration(t * float64(time.Second))
	results := sendMethod.Call([]reflect.Value{
		reflect.ValueOf(cmd),
		reflect.ValueOf(timeoutDuration),
	})
	if len(results) != 3 {
		return nil, false, fmt.Errorf("SendWSCommandWithTimeout 返回值数量不正确")
	}

	timedOut := results[1].Bool()
	var err error
	if !results[2].IsNil() {
		err = results[2].Interface().(error)
	}

	return results[0].Interface(), timedOut, err
}

// SayTo 向指定目标发送聊天消息（使用 tellraw）
// target: 目标玩家选择器或名称
// text: 消息内容
//
// 示例:
//   utils.SayTo("@a", "欢迎来到服务器！")
//   utils.SayTo("Steve", "你好！")
func (g *GameUtils) SayTo(target, text string) error {
	return g.Tellraw(target, text)
}

// PlayerTitle 向指定玩家显示标题
// target: 目标玩家选择器或名称
// text: 标题文本
//
// 示例:
//   utils.PlayerTitle("@a", "游戏开始")
func (g *GameUtils) PlayerTitle(target, text string) error {
	cmd := fmt.Sprintf("title %s title %s", target, text)
	return g.SendCommand(cmd)
}

// PlayerSubtitle 向指定玩家显示副标题
// target: 目标玩家选择器或名称
// text: 副标题文本
//
// 示例:
//   utils.PlayerSubtitle("@a", "Good Luck")
func (g *GameUtils) PlayerSubtitle(target, text string) error {
	cmd := fmt.Sprintf("title %s subtitle %s", target, text)
	return g.SendCommand(cmd)
}

// PlayerActionbar 向指定玩家显示 ActionBar 消息
// target: 目标玩家选择器或名称
// text: 消息文本
//
// 示例:
//   utils.PlayerActionbar("@a", "当前血量: 20/20")
func (g *GameUtils) PlayerActionbar(target, text string) error {
	cmd := fmt.Sprintf("title %s actionbar %s", target, text)
	return g.SendCommand(cmd)
}

// SendWOCommand 发送高权限控制台命令（Settings 通道）
// cmd: 控制台命令
//
// 注意: 这个方法需要主程序支持 Settings 通道
//
// 示例:
//   utils.SendWOCommand("list")
func (g *GameUtils) SendWOCommand(cmd string) error {
	giVal := reflect.ValueOf(g.gi)
	if !giVal.IsValid() || giVal.IsNil() {
		return fmt.Errorf("gameInterface 未初始化")
	}

	commandsMethod := giVal.MethodByName("Commands")
	if !commandsMethod.IsValid() {
		return fmt.Errorf("gameInterface 不支持 Commands 方法")
	}

	commandsVal := commandsMethod.Call(nil)
	if len(commandsVal) == 0 || !commandsVal[0].IsValid() {
		return fmt.Errorf("Commands 返回值无效")
	}

	// 尝试调用 SendSettings 方法
	sendMethod := commandsVal[0].MethodByName("SendSettings")
	if !sendMethod.IsValid() {
		// 如果不支持，回退到普通命令
		return g.SendCommand(cmd)
	}

	results := sendMethod.Call([]reflect.Value{reflect.ValueOf(cmd)})
	if len(results) > 0 && !results[0].IsNil() {
		return fmt.Errorf("发送控制台命令失败: %v", results[0].Interface())
	}

	return nil
}

// SendPacket 发送游戏网络数据包
// packetID: 数据包 ID
// packet: 数据包内容（map 或结构体）
//
// 注意: 这是低级 API，需要了解 Minecraft 协议
//
// 示例:
//   utils.SendPacket(0x09, map[string]interface{}{
//       "text": "Hello",
//   })
func (g *GameUtils) SendPacket(packetID uint32, packet interface{}) error {
	giVal := reflect.ValueOf(g.gi)
	if !giVal.IsValid() || giVal.IsNil() {
		return fmt.Errorf("gameInterface 未初始化")
	}

	// 尝试调用 SendPacket 方法
	sendMethod := giVal.MethodByName("SendPacket")
	if !sendMethod.IsValid() {
		return fmt.Errorf("gameInterface 不支持 SendPacket 方法")
	}

	results := sendMethod.Call([]reflect.Value{
		reflect.ValueOf(packetID),
		reflect.ValueOf(packet),
	})
	if len(results) > 0 && !results[0].IsNil() {
		return fmt.Errorf("发送数据包失败: %v", results[0].Interface())
	}

	return nil
}

// GetInventory 查询玩家背包信息
// selector: 目标玩家名称或选择器
// 返回: 背包槽位列表，包含物品 ID、数量等信息
//
// 注意: 这个方法需要服务器支持背包查询命令或数据包
//
// 示例:
//   slots, err := ctx.GameUtils().GetInventory("Steve")
//   if err == nil {
//       for _, slot := range slots {
//           ctx.Logf("槽位 %d: %s x%d", slot.Slot, slot.ItemID, slot.Count)
//       }
//   }
func (g *GameUtils) GetInventory(selector string) ([]InventorySlot, error) {
	giVal := reflect.ValueOf(g.gi)
	if !giVal.IsValid() || giVal.IsNil() {
		return nil, fmt.Errorf("gameInterface 未初始化")
	}

	// 使用 replaceitem 测试每个槽位来获取背包信息
	// 这是一种间接方法，因为 Minecraft 基岩版没有直接的背包查询命令
	// 更好的实现需要拦截 InventoryContent 数据包

	// 返回空实现，提示需要数据包拦截
	return nil, fmt.Errorf("GetInventory 需要通过拦截 InventoryContent 数据包实现，当前版本暂不支持")
}

// GetBlock 获取指定坐标的方块类型
// x, y, z: 方块坐标
// 返回: 方块 ID (如 "minecraft:stone")
//
// 示例:
//   blockID, err := ctx.GameUtils().GetBlock(100, 64, 100)
//   if err == nil {
//       ctx.Logf("方块类型: %s", blockID)
//   }
func (g *GameUtils) GetBlock(x, y, z int) (string, error) {
	giVal := reflect.ValueOf(g.gi)
	if !giVal.IsValid() || giVal.IsNil() {
		return "", fmt.Errorf("gameInterface 未初始化")
	}

	commandsMethod := giVal.MethodByName("Commands")
	if !commandsMethod.IsValid() {
		return "", fmt.Errorf("gameInterface 不支持 Commands 方法")
	}

	commandsVal := commandsMethod.Call(nil)
	if len(commandsVal) == 0 || !commandsVal[0].IsValid() {
		return "", fmt.Errorf("Commands 返回值无效")
	}

	// 使用 testforblock 命令来检测方块
	// 由于没有直接的查询命令，这里使用 setblock 的 keep 模式来检测
	cmd := fmt.Sprintf("testforblock %d %d %d air", x, y, z)

	sendMethod := commandsVal[0].MethodByName("SendWSCommandWithResp")
	if !sendMethod.IsValid() {
		return "", fmt.Errorf("Commands 不支持 SendWSCommandWithResp 方法")
	}

	results := sendMethod.Call([]reflect.Value{reflect.ValueOf(cmd)})
	if len(results) != 2 {
		return "", fmt.Errorf("SendWSCommandWithResp 返回值数量不正确")
	}

	// 如果成功，说明是空气方块
	if results[1].IsNil() {
		return "minecraft:air", nil
	}

	// 无法直接获取方块 ID，需要通过其他方式
	// 返回提示信息
	return "", fmt.Errorf("GetBlock 需要通过拦截 BlockUpdate 数据包或使用结构方块实现，当前版本暂不支持精确查询")
}

// EffectOptions 药水效果配置选项
type EffectOptions struct {
	Duration      int  // 持续时间（秒）
	Level         int  // 效果等级（0 表示 I 级）
	HideParticles bool // 是否隐藏粒子效果
}

// SetEffect 给玩家添加药水效果
// target: 目标玩家选择器或名称
// effectID: 效果 ID（如 1=速度, 2=缓慢, 5=力量, 10=再生）
// opts: 效果配置选项
//
// 常用效果 ID:
//   1  - speed (速度)
//   2  - slowness (缓慢)
//   3  - haste (急迫)
//   4  - mining_fatigue (挖掘疲劳)
//   5  - strength (力量)
//   6  - instant_health (瞬间治疗)
//   8  - jump_boost (跳跃提升)
//   10 - regeneration (再生)
//   11 - resistance (抗性提升)
//   12 - fire_resistance (抗火)
//   13 - water_breathing (水下呼吸)
//   14 - invisibility (隐身)
//
// 示例:
//   ctx.GameUtils().SetEffect("@a", 1, EffectOptions{
//       Duration:      60,
//       Level:         1,
//       HideParticles: true,
//   })
func (g *GameUtils) SetEffect(target string, effectID int, opts EffectOptions) error {
	giVal := reflect.ValueOf(g.gi)
	if !giVal.IsValid() || giVal.IsNil() {
		return fmt.Errorf("gameInterface 未初始化")
	}

	commandsMethod := giVal.MethodByName("Commands")
	if !commandsMethod.IsValid() {
		return fmt.Errorf("gameInterface 不支持 Commands 方法")
	}

	commandsVal := commandsMethod.Call(nil)
	if len(commandsVal) == 0 || !commandsVal[0].IsValid() {
		return fmt.Errorf("Commands 返回值无效")
	}

	// 设置默认值
	if opts.Duration <= 0 {
		opts.Duration = 30 // 默认 30 秒
	}
	if opts.Level < 0 {
		opts.Level = 0
	}

	// 构造 effect 命令
	hideParticle := ""
	if opts.HideParticles {
		hideParticle = " true"
	}

	cmd := fmt.Sprintf("effect \"%s\" %d %d %d%s", target, effectID, opts.Duration, opts.Level, hideParticle)

	sendMethod := commandsVal[0].MethodByName("SendWSCommand")
	if !sendMethod.IsValid() {
		return fmt.Errorf("Commands 不支持 SendWSCommand 方法")
	}

	results := sendMethod.Call([]reflect.Value{reflect.ValueOf(cmd)})
	if len(results) != 1 {
		return fmt.Errorf("SendWSCommand 返回值数量不正确")
	}

	if !results[0].IsNil() {
		return fmt.Errorf("添加药水效果失败: %v", results[0].Interface())
	}

	return nil
}

// ClearEffect 清除玩家的药水效果
// target: 目标玩家选择器或名称
// effectID: 效果 ID（可选，如果不指定则清除所有效果）
//
// 示例:
//   // 清除特定效果
//   ctx.GameUtils().ClearEffect("Steve", 1)
//   // 清除所有效果
//   ctx.GameUtils().ClearEffect("Steve", -1)
func (g *GameUtils) ClearEffect(target string, effectID int) error {
	giVal := reflect.ValueOf(g.gi)
	if !giVal.IsValid() || giVal.IsNil() {
		return fmt.Errorf("gameInterface 未初始化")
	}

	commandsMethod := giVal.MethodByName("Commands")
	if !commandsMethod.IsValid() {
		return fmt.Errorf("gameInterface 不支持 Commands 方法")
	}

	commandsVal := commandsMethod.Call(nil)
	if len(commandsVal) == 0 || !commandsVal[0].IsValid() {
		return fmt.Errorf("Commands 返回值无效")
	}

	var cmd string
	if effectID < 0 {
		// 清除所有效果
		cmd = fmt.Sprintf("effect \"%s\" clear", target)
	} else {
		// 清除特定效果
		cmd = fmt.Sprintf("effect \"%s\" clear %d", target, effectID)
	}

	sendMethod := commandsVal[0].MethodByName("SendWSCommand")
	if !sendMethod.IsValid() {
		return fmt.Errorf("Commands 不支持 SendWSCommand 方法")
	}

	results := sendMethod.Call([]reflect.Value{reflect.ValueOf(cmd)})
	if len(results) != 1 {
		return fmt.Errorf("SendWSCommand 返回值数量不正确")
	}

	if !results[0].IsNil() {
		return fmt.Errorf("清除药水效果失败: %v", results[0].Interface())
	}

	return nil
}
