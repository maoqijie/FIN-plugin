package main

import (
	"encoding/json"
	"fmt"
	"os"

	sdk "github.com/maoqijie/FIN-plugin/sdk"
)

// DataManagementPlugin 展示插件数据管理的示例
type DataManagementPlugin struct {
	ctx *sdk.Context
}

// PlayerData 玩家数据结构
type PlayerData struct {
	Name       string `json:"name"`
	Score      int    `json:"score"`
	Level      int    `json:"level"`
	LastOnline string `json:"last_online"`
}

// Init 插件初始化
func (p *DataManagementPlugin) Init(ctx *sdk.Context) error {
	p.ctx = ctx

	// 注册控制台命令
	ctx.RegisterConsoleCommand(sdk.ConsoleCommand{
		Triggers:    []string{"datatest", "dt"},
		Usage:       "测试插件数据管理功能",
		Description: "演示如何使用 DataPath 和 FormatDataPath",
		Handler:     p.handleDataTest,
	})

	return nil
}

// Start 插件启动
func (p *DataManagementPlugin) Start() error {
	p.ctx.Logf("数据管理插件已启动")

	// 演示 DataPath 用法
	p.demoDataPath()

	// 演示 FormatDataPath 用法
	p.demoFormatDataPath()

	// 演示实际数据读写
	p.demoDataReadWrite()

	return nil
}

// Stop 插件停止
func (p *DataManagementPlugin) Stop() error {
	p.ctx.Logf("数据管理插件已停止")
	return nil
}

// demoDataPath 演示 DataPath 方法
func (p *DataManagementPlugin) demoDataPath() {
	// 获取插件数据目录（自动创建）
	dataPath := p.ctx.DataPath()
	p.ctx.Logf("插件数据目录: %s", dataPath)

	// 目录会自动创建，例如: "plugins/data_management/"
}

// demoFormatDataPath 演示 FormatDataPath 方法
func (p *DataManagementPlugin) demoFormatDataPath() {
	// 获取配置文件路径
	configPath := p.ctx.FormatDataPath("config.json")
	p.ctx.Logf("配置文件路径: %s", configPath)

	// 获取子目录文件路径
	playerDataPath := p.ctx.FormatDataPath("players", "steve.json")
	p.ctx.Logf("玩家数据路径: %s", playerDataPath)

	// 获取多层子目录文件路径
	backupPath := p.ctx.FormatDataPath("backups", "2024-10", "backup.json")
	p.ctx.Logf("备份文件路径: %s", backupPath)
}

// demoDataReadWrite 演示实际数据读写
func (p *DataManagementPlugin) demoDataReadWrite() {
	// 创建示例数据
	playerData := PlayerData{
		Name:       "Steve",
		Score:      100,
		Level:      10,
		LastOnline: "2024-10-02",
	}

	// 保存数据
	playerPath := p.ctx.FormatDataPath("players", "steve.json")
	if err := p.savePlayerData(playerPath, playerData); err != nil {
		p.ctx.Logf("保存玩家数据失败: %v", err)
		return
	}
	p.ctx.Logf("玩家数据已保存到: %s", playerPath)

	// 读取数据
	loadedData, err := p.loadPlayerData(playerPath)
	if err != nil {
		p.ctx.Logf("加载玩家数据失败: %v", err)
		return
	}
	p.ctx.Logf("加载的玩家数据: %+v", loadedData)
}

// savePlayerData 保存玩家数据到文件
func (p *DataManagementPlugin) savePlayerData(path string, data PlayerData) error {
	// 序列化为 JSON
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化失败: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(path, jsonData, 0644); err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	return nil
}

// loadPlayerData 从文件加载玩家数据
func (p *DataManagementPlugin) loadPlayerData(path string) (*PlayerData, error) {
	// 读取文件
	jsonData, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %w", err)
	}

	// 反序列化
	var data PlayerData
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return nil, fmt.Errorf("反序列化失败: %w", err)
	}

	return &data, nil
}

// handleDataTest 处理数据测试命令
func (p *DataManagementPlugin) handleDataTest(args []string) error {
	p.ctx.Logf("=== 插件数据管理演示 ===")

	// 1. 演示获取数据目录
	dataPath := p.ctx.DataPath()
	p.ctx.Logf("1. 数据目录: %s", dataPath)

	// 2. 演示格式化路径
	configPath := p.ctx.FormatDataPath("config.json")
	p.ctx.Logf("2. 配置路径: %s", configPath)

	// 3. 演示创建和读取数据
	testData := PlayerData{
		Name:  "TestPlayer",
		Score: 999,
		Level: 99,
	}

	testPath := p.ctx.FormatDataPath("test_data.json")
	if err := p.savePlayerData(testPath, testData); err != nil {
		return fmt.Errorf("保存测试数据失败: %w", err)
	}
	p.ctx.Logf("3. 测试数据已保存")

	loadedData, err := p.loadPlayerData(testPath)
	if err != nil {
		return fmt.Errorf("加载测试数据失败: %w", err)
	}
	p.ctx.Logf("4. 测试数据已加载: %+v", loadedData)

	// 4. 演示配合 TempJSON 使用
	p.demoWithTempJSON()

	return nil
}

// demoWithTempJSON 演示配合 TempJSON 使用
func (p *DataManagementPlugin) demoWithTempJSON() {
	// 使用 TempJSON 管理插件数据
	tj := p.ctx.TempJSON(p.ctx.DataPath())

	// 保存数据
	playerData := map[string]interface{}{
		"name":  "Alex",
		"score": 200,
		"level": 20,
	}

	// 使用 FormatDataPath 生成路径，配合 TempJSON
	relativePath := "cached_players/alex.json"
	fullPath := p.ctx.FormatDataPath(relativePath)

	// 使用相对路径（去掉数据目录前缀）
	tj.LoadAndWrite(relativePath, playerData, true, 60.0)

	p.ctx.Logf("5. 使用 TempJSON 缓存数据到: %s", fullPath)
}

// NewPlugin 插件入口点
func NewPlugin() sdk.Plugin {
	return &DataManagementPlugin{}
}
