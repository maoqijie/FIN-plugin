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
		_ = results[0].Index(i)
		// TODO: 从结构体中提取玩家名称
		// 当前返回空切片，需要解析 TargetQueryingInfo 结构
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
	outputMsgs := output.FieldByName("OutputMessages")
	if outputMsgs.Len() == 0 {
		return 0, fmt.Errorf("无法获取分数")
	}

	firstMsg := outputMsgs.Index(0)
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
