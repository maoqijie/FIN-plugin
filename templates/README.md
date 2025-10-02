# 插件模板示例

本目录包含各种插件开发模板和示例代码，帮助你快速上手 FIN 插件开发。

## 📦 可用模板

### 基础示例

#### 1. [bot/info](bot/info/) - 控制台命令示例
展示如何注册和处理控制台命令。

**功能**：
- 注册控制台命令 `info`
- 显示机器人、服务器、QQ 适配器信息
- 演示 `Context` 的基本用法

**适合场景**：需要添加管理命令的插件

---

#### 2. [data_management](data_management/) - 数据目录管理
展示如何管理插件数据文件和目录。

**功能**：
- 使用 `DataPath()` 获取数据目录
- 使用 `FormatDataPath()` 格式化文件路径
- 演示配置文件、玩家数据、备份管理
- 配合 `TempJSON` 使用示例

**适合场景**：需要保存配置或玩家数据的插件

---

### 高级示例

#### 3. [api_plugin](api_plugin/) - 前置插件（API 提供者）
展示如何创建可被其他插件调用的 API 插件。

**功能**：
- 注册为 API 插件
- 提供公共方法给其他插件
- 版本管理和兼容性控制

**适合场景**：需要提供通用功能给其他插件的基础插件

---

#### 4. [api_consumer](api_consumer/) - API 消费者
展示如何使用其他插件提供的 API。

**功能**：
- 查找并获取 API 插件
- 调用其他插件的方法
- 版本检查和兼容性处理

**适合场景**：需要依赖其他插件功能的插件

---

### 完整应用示例

#### 5. [shop](shop/) - 商店系统 ⭐
完整的游戏内商店系统，参考 ToolDelta 设计。

**功能**：
- 双向交易系统（购买/出售）
- 交互式菜单界面
- 积分货币系统
- 配置化物品管理
- 完整的交易流程和验证
- 超时自动清理

**技术亮点**：
- 状态机管理（等待用户输入）
- 配置文件驱动
- 计分板集成
- 物品和货币验证
- 多玩家并发支持

**适合场景**：
- 经济系统插件
- 游戏内交易系统
- 需要复杂交互流程的插件

---

## 🚀 快速开始

### 1. 选择模板

根据你的需求选择合适的模板：

- **第一次开发插件**？从 `bot/info` 开始
- **需要保存数据**？查看 `data_management`
- **开发复杂功能**？参考 `shop` 示例
- **提供通用功能**？学习 `api_plugin`

### 2. 复制模板

```bash
# 复制到 Plugin/grpc 目录
cp -r templates/shop Plugin/grpc/my-shop

# 或在控制台使用 create 命令创建新插件
```

### 3. 修改代码

编辑 `main.go`，实现你的功能逻辑。

### 4. 编译运行

```bash
cd Plugin/grpc/my-shop
go build -buildmode=plugin -o main.so main.go
```

或者直接将 `.go` 文件放入插件目录，框架会自动编译。

## 📚 学习路径

### 初级（必学）

1. **bot/info** - 了解插件结构和生命周期
2. **data_management** - 学习数据管理

### 中级

3. **api_plugin + api_consumer** - 理解插件间协作

### 高级

4. **shop** - 研究完整的应用架构

## 🎯 模板对比

| 模板 | 复杂度 | 代码量 | 主要 API | 学习时间 |
|------|--------|--------|----------|----------|
| bot/info | ⭐ | 50 行 | Context, Console | 10 分钟 |
| data_management | ⭐⭐ | 200 行 | DataPath, TempJSON, Config | 20 分钟 |
| api_plugin | ⭐⭐ | 100 行 | RegisterPluginAPI | 15 分钟 |
| api_consumer | ⭐⭐ | 80 行 | GetPluginAPI | 15 分钟 |
| shop | ⭐⭐⭐⭐ | 350 行 | GameUtils, Config, ListenChat | 60 分钟 |

## 📖 详细文档

每个模板目录都包含：

- `main.go` - 完整的源代码
- `README.md` - 详细的使用说明
- `go.mod` - 依赖配置

查看具体模板的 README 了解更多细节。

## 🔧 开发建议

### 代码结构

```go
package main

import sdk "github.com/maoqijie/FIN-plugin/sdk"

type MyPlugin struct {
    ctx *sdk.Context
}

func (p *MyPlugin) Init(ctx *sdk.Context) error {
    p.ctx = ctx
    // 初始化代码
    return nil
}

func (p *MyPlugin) Start() error {
    // 启动代码
    return nil
}

func (p *MyPlugin) Stop() error {
    // 停止代码
    return nil
}

func NewPlugin() sdk.Plugin {
    return &MyPlugin{}
}
```

### 常用模式

#### 1. 监听事件

```go
func (p *MyPlugin) Init(ctx *sdk.Context) error {
    // 监听玩家加入
    ctx.ListenPlayerJoin(func(event sdk.PlayerEvent) {
        p.ctx.Logf("玩家 %s 加入了游戏", event.PlayerName)
    })

    // 监听聊天
    ctx.ListenChat(func(event sdk.ChatEvent) {
        p.ctx.Logf("玩家 %s 说: %s", event.PlayerName, event.Message)
    })

    return nil
}
```

#### 2. 配置管理

```go
type MyConfig struct {
    Enable bool   `json:"enable"`
    Value  int    `json:"value"`
}

func (p *MyPlugin) Init(ctx *sdk.Context) error {
    config, err := ctx.Config().GetConfig("config.json", &MyConfig{
        Enable: true,
        Value:  100,
    })
    if err != nil {
        return err
    }

    p.config = config.(*MyConfig)
    return nil
}
```

#### 3. 数据保存

```go
func (p *MyPlugin) saveData() error {
    path := p.ctx.FormatDataPath("data.json")
    data := map[string]interface{}{
        "players": p.players,
        "scores":  p.scores,
    }

    jsonData, _ := json.Marshal(data)
    return os.WriteFile(path, jsonData, 0644)
}
```

#### 4. 游戏交互

```go
func (p *MyPlugin) handlePlayer(playerName string) {
    // 发送消息
    p.ctx.GameUtils().SayTo(playerName, "§a欢迎！")

    // 获取坐标
    pos, _ := p.ctx.GameUtils().GetPos(playerName)

    // 获取分数
    score, _ := p.ctx.GameUtils().GetScore("money", playerName, 5.0)

    // 传送玩家
    p.ctx.GameUtils().SendCommand(fmt.Sprintf(
        "tp \"%s\" 100 64 100", playerName,
    ))
}
```

## 🤝 贡献

欢迎提交新的示例模板！

提交要求：
1. 代码清晰注释
2. 完整的 README 文档
3. 测试通过
4. 遵循现有代码风格

## 📄 许可证

所有模板遵循主仓库的许可证条款。
