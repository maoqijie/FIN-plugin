# 插件数据管理

插件通常需要保存配置、玩家数据、日志等信息。SDK 提供了便捷的数据目录管理方法，类似 ToolDelta 的 `data_path` 和 `format_data_path`。

## 核心方法

### DataPath() - 获取数据目录

自动创建并返回插件专属数据文件夹路径。

```go
dataPath := ctx.DataPath()
// 返回: "plugins/{插件名}/"
```

**特点**：
- 自动根据插件名称创建目录
- 默认路径为 `plugins/{插件名}/`
- 调用时自动确保目录存在
- 目录权限为 0755

**示例**：
```go
func (p *Plugin) Init(ctx *sdk.Context) error {
    // 获取插件数据目录
    dataPath := ctx.DataPath()
    ctx.Logf("数据目录: %s", dataPath)
    // 输出: plugins/my-plugin/

    return nil
}
```

### FormatDataPath(...path) - 格式化路径

方便地生成插件内部文件路径，支持多个路径片段。

```go
path := ctx.FormatDataPath(path ...string) string
```

**参数**：
- `path`: 可变参数，相对于插件数据目录的路径片段

**返回**：
- 完整的文件路径

**特点**：
- 支持多个路径片段
- 自动创建父目录
- 返回完整的文件路径
- 父目录权限为 0755

**示例**：
```go
// 单个路径片段
configPath := ctx.FormatDataPath("config.json")
// 返回: "plugins/my-plugin/config.json"

// 多个路径片段（子目录）
playerPath := ctx.FormatDataPath("players", "steve.json")
// 返回: "plugins/my-plugin/players/steve.json"

// 多层子目录
backupPath := ctx.FormatDataPath("backups", "2024-10", "backup.json")
// 返回: "plugins/my-plugin/backups/2024-10/backup.json"
```

## 使用场景

### 1. 保存配置文件

```go
func (p *Plugin) saveConfig() error {
    configPath := p.ctx.FormatDataPath("config.json")

    config := map[string]interface{}{
        "enable": true,
        "level":  5,
        "mode":   "normal",
    }

    // 序列化为 JSON
    data, err := json.MarshalIndent(config, "", "  ")
    if err != nil {
        return err
    }

    // 写入文件
    return os.WriteFile(configPath, data, 0644)
}

func (p *Plugin) loadConfig() (map[string]interface{}, error) {
    configPath := p.ctx.FormatDataPath("config.json")

    // 读取文件
    data, err := os.ReadFile(configPath)
    if err != nil {
        return nil, err
    }

    // 反序列化
    var config map[string]interface{}
    err = json.Unmarshal(data, &config)
    return config, err
}
```

### 2. 管理玩家数据

```go
type PlayerData struct {
    Name       string `json:"name"`
    Score      int    `json:"score"`
    Level      int    `json:"level"`
    LastOnline string `json:"last_online"`
}

func (p *Plugin) savePlayerData(playerName string, data PlayerData) error {
    // 每个玩家一个文件，存放在 players/ 子目录
    playerPath := p.ctx.FormatDataPath("players", playerName+".json")

    jsonData, err := json.MarshalIndent(data, "", "  ")
    if err != nil {
        return err
    }

    return os.WriteFile(playerPath, jsonData, 0644)
}

func (p *Plugin) loadPlayerData(playerName string) (*PlayerData, error) {
    playerPath := p.ctx.FormatDataPath("players", playerName+".json")

    jsonData, err := os.ReadFile(playerPath)
    if err != nil {
        return nil, err
    }

    var data PlayerData
    err = json.Unmarshal(jsonData, &data)
    return &data, err
}
```

### 3. 创建备份

```go
import "time"

func (p *Plugin) createBackup() error {
    // 按日期组织备份
    today := time.Now().Format("2006-01-02")
    backupPath := p.ctx.FormatDataPath("backups", today, "backup.json")

    // 收集要备份的数据
    backupData := map[string]interface{}{
        "timestamp": time.Now().Unix(),
        "players":   p.getAllPlayers(),
        "scores":    p.getAllScores(),
    }

    jsonData, err := json.MarshalIndent(backupData, "", "  ")
    if err != nil {
        return err
    }

    return os.WriteFile(backupPath, jsonData, 0644)
}
```

### 4. 配合 TempJSON 使用

```go
func (p *Plugin) useCachedData() {
    // 使用插件数据目录创建 TempJSON 实例
    tj := p.ctx.TempJSON(p.ctx.DataPath())

    // 快速写入缓存
    playerData := map[string]interface{}{
        "name":  "Steve",
        "score": 100,
    }
    tj.LoadAndWrite("cache/player.json", playerData, true, 60.0)

    // 快速读取缓存
    data, _ := tj.LoadAndRead("cache/player.json", false, nil, 0)
}
```

### 5. 配合 Config 使用

```go
func (p *Plugin) useConfig() {
    // Config 默认使用 plugins/{插件名}/ 目录
    cfg := p.ctx.Config()

    // 获取配置
    config, _ := cfg.GetConfig("settings.json", map[string]interface{}{
        "enable": true,
    })

    // 也可以使用 FormatDataPath 获取配置路径
    configPath := p.ctx.FormatDataPath("settings.json")
    p.ctx.Logf("配置文件: %s", configPath)
}
```

## 目录组织示例

使用数据管理方法后，插件数据会自动组织成：

```
plugins/
└── my-plugin/                 # 插件数据目录
    ├── config.json            # 配置文件
    ├── players/               # 玩家数据子目录
    │   ├── steve.json
    │   ├── alex.json
    │   └── notch.json
    ├── cache/                 # 缓存数据
    │   └── player.json
    ├── backups/               # 备份目录
    │   ├── 2024-10-01/
    │   │   └── backup.json
    │   └── 2024-10-02/
    │       └── backup.json
    └── logs/                  # 日志目录
        ├── 2024-10-01.log
        └── 2024-10-02.log
```

## 完整示例

```go
package main

import (
    "encoding/json"
    "os"
    sdk "github.com/maoqijie/FIN-plugin/sdk"
)

type MyPlugin struct {
    ctx *sdk.Context
}

type GameData struct {
    Players []string          `json:"players"`
    Scores  map[string]int    `json:"scores"`
}

func (p *MyPlugin) Init(ctx *sdk.Context) error {
    p.ctx = ctx

    // 1. 获取数据目录
    dataPath := ctx.DataPath()
    ctx.Logf("数据目录: %s", dataPath)

    // 2. 加载或创建游戏数据
    if err := p.loadOrCreateGameData(); err != nil {
        return err
    }

    return nil
}

func (p *MyPlugin) loadOrCreateGameData() error {
    dataPath := p.ctx.FormatDataPath("game_data.json")

    // 检查文件是否存在
    if _, err := os.Stat(dataPath); os.IsNotExist(err) {
        // 创建默认数据
        defaultData := GameData{
            Players: []string{},
            Scores:  make(map[string]int),
        }
        return p.saveGameData(defaultData)
    }

    // 加载现有数据
    _, err := p.loadGameData()
    return err
}

func (p *MyPlugin) saveGameData(data GameData) error {
    dataPath := p.ctx.FormatDataPath("game_data.json")

    jsonData, err := json.MarshalIndent(data, "", "  ")
    if err != nil {
        return err
    }

    return os.WriteFile(dataPath, jsonData, 0644)
}

func (p *MyPlugin) loadGameData() (*GameData, error) {
    dataPath := p.ctx.FormatDataPath("game_data.json")

    jsonData, err := os.ReadFile(dataPath)
    if err != nil {
        return nil, err
    }

    var data GameData
    err = json.Unmarshal(jsonData, &data)
    return &data, err
}

func (p *MyPlugin) Start() error {
    return nil
}

func (p *MyPlugin) Stop() error {
    return nil
}

func NewPlugin() sdk.Plugin {
    return &MyPlugin{}
}
```

## 最佳实践

### 1. 使用 FormatDataPath 而不是手动拼接

```go
❌ 不推荐：手动拼接路径
path := "plugins/" + p.ctx.PluginName() + "/data.json"

✅ 推荐：使用 FormatDataPath
path := p.ctx.FormatDataPath("data.json")
```

### 2. 组织数据到子目录

```go
// 按类型组织数据
configPath := p.ctx.FormatDataPath("config", "settings.json")
playerPath := p.ctx.FormatDataPath("players", "steve.json")
backupPath := p.ctx.FormatDataPath("backups", "2024-10-02.json")
logPath := p.ctx.FormatDataPath("logs", "app.log")
```

### 3. 检查文件是否存在

```go
path := p.ctx.FormatDataPath("data.json")

if _, err := os.Stat(path); os.IsNotExist(err) {
    // 文件不存在，创建默认数据
    p.createDefaultData()
} else {
    // 文件存在，加载数据
    p.loadData()
}
```

### 4. 错误处理

```go
path := p.ctx.FormatDataPath("data.json")

if err := os.WriteFile(path, data, 0644); err != nil {
    p.ctx.Logf("保存数据失败: %v", err)
    return fmt.Errorf("保存失败: %w", err)
}
```

### 5. 配合其他工具使用

```go
// 使用 TempJSON 进行缓存管理
tj := p.ctx.TempJSON(p.ctx.DataPath())
tj.LoadAndWrite("cache.json", data, true, 60.0)

// 使用 Config 进行配置管理
cfg := p.ctx.Config()
config, _ := cfg.GetConfig("config.json", defaultConfig)
```

## 注意事项

1. **目录自动创建**
   - `DataPath()` 和 `FormatDataPath()` 会自动创建必要的目录
   - 无需手动调用 `os.MkdirAll()`

2. **插件名称唯一性**
   - 数据目录基于插件名称
   - 确保插件名称唯一，避免数据冲突

3. **文件权限**
   - 目录权限默认为 0755
   - 建议文件权限使用 0644

4. **路径安全**
   - 避免使用 `..` 等特殊路径
   - 防止越界访问其他插件数据

5. **数据备份**
   - 定期备份重要数据
   - 使用时间戳组织备份文件

## 与 ToolDelta 对比

### ToolDelta

```python
class MyPlugin(Plugin):
    name = "我的插件"

    def __init__(self):
        # 创建数据目录
        self.make_data_path()

        # 获取数据路径
        data_path = self.data_path

        # 格式化文件路径
        config_path = self.format_data_path("config.json")
        player_path = self.format_data_path("players", "steve.json")
```

### FunInterwork

```go
func (p *MyPlugin) Init(ctx *sdk.Context) error {
    // 获取数据目录（自动创建）
    dataPath := ctx.DataPath()

    // 格式化文件路径
    configPath := ctx.FormatDataPath("config.json")
    playerPath := ctx.FormatDataPath("players", "steve.json")

    return nil
}
```

**主要区别**：
- Go 版本无需手动调用创建目录方法
- `DataPath()` 自动确保目录存在
- 方法命名更符合 Go 风格
- 返回完整路径，更便于直接使用

## 示例代码

完整示例位于 `templates/data_management/`：
- 展示所有数据管理方法
- 演示实际数据读写
- 展示配合其他工具使用
