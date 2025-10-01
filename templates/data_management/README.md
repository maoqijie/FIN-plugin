# 插件数据管理示例

这是一个展示如何管理插件数据的示例插件。

## 功能说明

此插件演示如何使用 SDK 提供的数据目录管理功能：

- `DataPath()` - 获取插件专属数据目录
- `FormatDataPath()` - 格式化数据文件路径
- 配合 `TempJSON` 进行高效数据缓存

## 核心方法

### DataPath() - 获取数据目录

自动创建并返回插件专属数据文件夹路径。

```go
// 获取插件数据目录
dataPath := ctx.DataPath()
// 返回: "plugins/{插件名}/"
```

特点：
- 自动根据插件名称创建目录
- 默认路径为 `plugins/{插件名}/`
- 调用时自动确保目录存在

### FormatDataPath(...path) - 格式化路径

方便地生成插件内部文件路径。

```go
// 获取配置文件路径
configPath := ctx.FormatDataPath("config.json")
// 返回: "plugins/my-plugin/config.json"

// 获取子目录文件路径
playerPath := ctx.FormatDataPath("players", "steve.json")
// 返回: "plugins/my-plugin/players/steve.json"

// 获取多层子目录路径
backupPath := ctx.FormatDataPath("backups", "2024-10", "backup.json")
// 返回: "plugins/my-plugin/backups/2024-10/backup.json"
```

特点：
- 支持多个路径片段
- 自动创建父目录
- 返回完整的文件路径

## 使用场景

### 1. 保存配置文件

```go
func (p *Plugin) saveConfig() error {
    configPath := p.ctx.FormatDataPath("config.json")

    config := map[string]interface{}{
        "enable": true,
        "level": 5,
    }

    // 序列化并保存
    data, _ := json.Marshal(config)
    return os.WriteFile(configPath, data, 0644)
}
```

### 2. 管理玩家数据

```go
func (p *Plugin) savePlayerData(playerName string, data PlayerData) error {
    // 每个玩家一个文件
    playerPath := p.ctx.FormatDataPath("players", playerName + ".json")

    jsonData, _ := json.MarshalIndent(data, "", "  ")
    return os.WriteFile(playerPath, jsonData, 0644)
}
```

### 3. 创建备份

```go
func (p *Plugin) createBackup() error {
    // 按日期组织备份
    today := time.Now().Format("2006-01-02")
    backupPath := p.ctx.FormatDataPath("backups", today, "backup.json")

    // 保存备份
    return p.saveBackup(backupPath)
}
```

### 4. 配合 TempJSON 使用

```go
func (p *Plugin) useTempJSON() {
    // 创建 TempJSON 实例，使用插件数据目录
    tj := p.ctx.TempJSON(p.ctx.DataPath())

    // 使用缓存读写
    data := map[string]interface{}{"score": 100}
    tj.LoadAndWrite("scores.json", data, true, 60.0)
}
```

### 5. 配合 Config 使用

```go
func (p *Plugin) useConfig() {
    // Config 默认使用 plugins/{插件名}/ 目录
    cfg := p.ctx.Config()

    config, _ := cfg.GetConfig("settings.json", defaultConfig)

    // 也可以使用 FormatDataPath 获取配置路径
    configPath := p.ctx.FormatDataPath("settings.json")
}
```

## 目录结构示例

使用这些方法后，插件数据会组织成：

```
plugins/
└── data_management/           # 插件数据目录
    ├── config.json            # 配置文件
    ├── test_data.json         # 测试数据
    ├── players/               # 玩家数据子目录
    │   ├── steve.json
    │   ├── alex.json
    │   └── ...
    ├── cached_players/        # 缓存数据
    │   └── alex.json
    └── backups/               # 备份目录
        ├── 2024-10/
        │   └── backup.json
        └── 2024-11/
            └── backup.json
```

## 运行示例

### 1. 启动插件

插件启动时会自动演示所有功能：

```
[data_management] 数据管理插件已启动
[data_management] 插件数据目录: plugins/data_management/
[data_management] 配置文件路径: plugins/data_management/config.json
[data_management] 玩家数据路径: plugins/data_management/players/steve.json
[data_management] 玩家数据已保存到: plugins/data_management/players/steve.json
[data_management] 加载的玩家数据: {Name:Steve Score:100 Level:10 LastOnline:2024-10-02}
```

### 2. 测试命令

在控制台输入 `datatest` 或 `dt`：

```bash
datatest
```

输出：
```
=== 插件数据管理演示 ===
1. 数据目录: plugins/data_management/
2. 配置路径: plugins/data_management/config.json
3. 测试数据已保存
4. 测试数据已加载: {Name:TestPlayer Score:999 Level:99}
5. 使用 TempJSON 缓存数据到: plugins/data_management/cached_players/alex.json
```

## 最佳实践

1. **使用 FormatDataPath 而不是手动拼接路径**
   ```go
   ❌ path := "plugins/" + p.ctx.PluginName() + "/data.json"
   ✅ path := p.ctx.FormatDataPath("data.json")
   ```

2. **组织数据到子目录**
   ```go
   // 按类型组织
   configPath := p.ctx.FormatDataPath("config", "settings.json")
   playerPath := p.ctx.FormatDataPath("players", "steve.json")
   backupPath := p.ctx.FormatDataPath("backups", "2024-10-02.json")
   ```

3. **配合其他工具使用**
   - 使用 `TempJSON` 进行缓存管理
   - 使用 `Config` 进行配置管理
   - 使用标准库 `os` 包进行文件操作

4. **错误处理**
   ```go
   path := p.ctx.FormatDataPath("data.json")
   if err := os.WriteFile(path, data, 0644); err != nil {
       p.ctx.Logf("保存失败: %v", err)
       return err
   }
   ```

## 注意事项

1. **目录自动创建**：调用 `DataPath()` 和 `FormatDataPath()` 会自动创建必要的目录
2. **插件名称**：数据目录基于插件名称，确保插件名称唯一
3. **权限**：目录权限默认为 0755，文件建议使用 0644
4. **路径安全**：避免使用 `..` 等特殊路径，以防越界访问

## 与 ToolDelta 对比

```python
# ToolDelta
self.make_data_path()                           # 创建数据目录
path = self.data_path                           # 获取数据路径
file_path = self.format_data_path("data.json") # 格式化路径
```

```go
// FunInterwork
dataPath := ctx.DataPath()                      // 获取数据目录
filePath := ctx.FormatDataPath("data.json")     // 格式化路径
```

主要区别：
- Go 版本自动创建目录，无需手动调用
- 方法命名更符合 Go 风格
- 返回完整路径，更便于使用
