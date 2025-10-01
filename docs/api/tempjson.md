## TempJSON - 缓存式 JSON 文件管理

`TempJSON` 提供高性能的 JSON 文件缓存管理，通过内存缓存减少磁盘 I/O，类似 ToolDelta 的 tempjson 模块。

### 获取 TempJSON 实例

```go
func (p *plugin) Init(ctx *sdk.Context) error {
    p.ctx = ctx

    // 获取 TempJSON 管理器（默认当前目录）
    tj := ctx.TempJSON()

    // 或指定默认目录
    tj := ctx.TempJSON("data")

    return nil
}
```

### 核心概念

TempJSON 通过内存缓存加速 JSON 文件操作：
- **加载（Load）**：将文件读入内存缓存
- **读写（Read/Write）**：在内存中操作数据
- **卸载（Unload）**：保存修改并从缓存中移除
- **自动卸载**：设置超时自动卸载，释放内存

### 快捷方法（推荐）

#### LoadAndRead - 快速读取

最常用的方法，适合一次性读取操作。

```go
func (p *plugin) Start() error {
    tj := p.ctx.TempJSON()

    // 快速读取（读取后立即卸载）
    data, err := tj.LoadAndRead("user_data.json", false, map[string]interface{}{
        "score": 0,
        "level": 1,
    }, 0) // timeout=0 表示立即卸载
    if err != nil {
        return err
    }

    // 类型断言
    dataMap := data.(map[string]interface{})
    score := int(dataMap["score"].(float64)) // JSON 数字默认为 float64
    level := int(dataMap["level"].(float64))

    p.ctx.Logf("用户分数: %d, 等级: %d", score, level)

    return nil
}
```

#### LoadAndWrite - 快速写入

适合一次性写入操作，写入后立即保存到磁盘。

```go
// 更新用户数据
userData := map[string]interface{}{
    "score": 1000,
    "level": 5,
    "last_login": time.Now().Format(time.RFC3339),
}

err := tj.LoadAndWrite("user_data.json", userData, false, 0)
if err != nil {
    return err
}
// 数据已自动保存到磁盘
```

### 持久缓存方法

适合需要频繁读写的场景。

#### Load - 加载到缓存

```go
tj := p.ctx.TempJSON()

// 加载文件到缓存（30 秒后自动卸载）
err := tj.Load("player_stats.json", false, map[string]interface{}{
    "players": []interface{}{},
}, 30.0)
if err != nil {
    return err
}

// 文件现在在内存中，可以多次读写
```

#### Read - 从缓存读取

```go
// 深拷贝读取（默认，安全）
data, err := tj.Read("player_stats.json", true)
if err != nil {
    return err
}

// 浅拷贝读取（性能更好，但要小心修改）
data, err := tj.Read("player_stats.json", false)
```

#### Write - 写入缓存

```go
// 修改数据
stats := map[string]interface{}{
    "players": []interface{}{
        map[string]interface{}{
            "name":  "玩家A",
            "score": 100,
        },
    },
}

// 写入缓存（不会立即保存到磁盘）
err := tj.Write("player_stats.json", stats)
```

#### Unload - 卸载缓存

```go
// 保存修改并从缓存中移除
err := tj.Unload("player_stats.json")
```

### 实际应用示例

#### 示例 1：玩家数据管理

```go
type plugin struct {
    ctx *sdk.Context
    tj  *sdk.TempJSON
}

func (p *plugin) Init(ctx *sdk.Context) error {
    p.ctx = ctx
    p.tj = ctx.TempJSON("plugin_data")
    return nil
}

// 获取玩家分数
func (p *plugin) GetPlayerScore(playerName string) (int, error) {
    // 快速读取
    data, err := p.tj.LoadAndRead("scores.json", false, map[string]interface{}{}, 0)
    if err != nil {
        return 0, err
    }

    scores := data.(map[string]interface{})
    if score, exists := scores[playerName]; exists {
        return int(score.(float64)), nil
    }

    return 0, nil
}

// 设置玩家分数
func (p *plugin) SetPlayerScore(playerName string, score int) error {
    // 读取现有数据
    data, err := p.tj.LoadAndRead("scores.json", false, map[string]interface{}{}, 30.0)
    if err != nil {
        return err
    }

    scores := data.(map[string]interface{})
    scores[playerName] = score

    // 写入并保存
    return p.tj.LoadAndWrite("scores.json", scores, false, 0)
}
```

#### 示例 2：高频读写场景

```go
func (p *plugin) Start() error {
    tj := p.ctx.TempJSON("cache")

    // 加载到缓存（60 秒后自动卸载）
    err := tj.Load("game_state.json", false, map[string]interface{}{
        "running": false,
        "players": []interface{}{},
    }, 60.0)
    if err != nil {
        return err
    }

    // 多次读写操作（在内存中进行，速度快）
    for i := 0; i < 100; i++ {
        // 读取
        data, _ := tj.Read("game_state.json", true)
        state := data.(map[string]interface{})

        // 修改
        state["tick"] = i
        players := state["players"].([]interface{})
        players = append(players, fmt.Sprintf("player_%d", i))
        state["players"] = players

        // 写入缓存
        tj.Write("game_state.json", state)
    }

    // 手动保存到磁盘
    return tj.Unload("game_state.json")
}
```

#### 示例 3：批量操作

```go
func (p *plugin) ProcessAllPlayerData() error {
    tj := p.ctx.TempJSON()

    // 加载多个文件到缓存
    files := []string{"players.json", "scores.json", "achievements.json"}
    for _, file := range files {
        err := tj.Load(file, false, map[string]interface{}{}, 30.0)
        if err != nil {
            p.ctx.Logf("加载 %s 失败: %v", file, err)
        }
    }

    // 批量处理
    for _, file := range files {
        data, err := tj.Read(file, true)
        if err != nil {
            continue
        }

        // 处理数据...
        processedData := p.processData(data)

        // 写回缓存
        tj.Write(file, processedData)
    }

    // 保存所有修改
    return tj.SaveAll()
}
```

### 高级功能

#### SaveAll - 保存所有缓存

保存所有已修改的文件到磁盘，但保持在缓存中。

```go
// 保存所有修改（不卸载）
err := tj.SaveAll()
```

#### UnloadAll - 卸载所有缓存

保存并卸载所有缓存的文件。

```go
// 插件停止时清理
func (p *plugin) Stop() error {
    return p.tj.UnloadAll()
}
```

#### GetCachedPaths - 获取缓存列表

```go
// 查看哪些文件在缓存中
paths := tj.GetCachedPaths()
for _, path := range paths {
    p.ctx.Logf("已缓存: %s", path)
}
```

#### IsCached - 检查缓存状态

```go
if tj.IsCached("data.json") {
    p.ctx.Logf("文件在缓存中")
}
```

### 性能优化建议

1. **使用快捷方法**：对于一次性操作，使用 `LoadAndRead` 和 `LoadAndWrite`
2. **合理设置超时**：频繁访问的文件设置较长超时（30-60 秒）
3. **及时卸载**：不再使用的文件及时卸载，释放内存
4. **批量操作**：多个文件需要处理时，一起加载到缓存后再批量处理
5. **深拷贝 vs 浅拷贝**：
   - 读取后要修改：使用深拷贝（`deepCopy=true`）
   - 只读取不修改：可使用浅拷贝（`deepCopy=false`）提升性能

### 注意事项

1. **JSON 数字类型**：JSON 解析后数字默认为 `float64`，需要类型转换
   ```go
   count := int(data["count"].(float64))
   ```

2. **并发安全**：TempJSON 内部已实现并发安全，可在多个 goroutine 中使用

3. **文件路径**：支持相对路径和绝对路径，相对路径基于 `defaultDir`

4. **自动卸载**：设置 `timeout > 0` 会启动定时器，超时后自动保存并卸载

5. **内存管理**：长期运行的插件应定期调用 `UnloadAll` 或设置合理的超时
