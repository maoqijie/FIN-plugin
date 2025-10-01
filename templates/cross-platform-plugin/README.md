# 跨平台插件模板

这是一个基于 Hashicorp go-plugin 的跨平台插件模板，支持 Windows、Linux、macOS 和 Android。

## 快速开始

### 1. 创建新插件

```bash
# 复制模板
cp -r templates/cross-platform-plugin plugins/my-plugin
cd plugins/my-plugin

# 修改插件名称
# 编辑 plugin.yaml，修改 name、displayName 等字段
# 编辑 main.go，修改 GetInfo() 返回的信息
```

### 2. 开发插件

在 `main.go` 中实现插件逻辑：

```go
// Init 方法中注册事件和命令
func (p *ExamplePlugin) Init(ctx *sdk.Context) error {
    p.ctx = ctx

    // 注册控制台命令
    ctx.RegisterConsoleCommand(sdk.ConsoleCommand{
        Name: "mycmd",
        Handler: func(args []string) error {
            ctx.LogInfo("命令被调用")
            return nil
        },
    })

    // 监听游戏事件
    ctx.ListenPlayerJoin(func(event sdk.PlayerEvent) {
        ctx.LogSuccess("玩家 %s 加入", event.Name)
    })

    return nil
}
```

### 3. 本地测试

```bash
# 安装依赖
go mod tidy

# 编译当前平台
go build -o my-plugin .

# 或使用构建脚本编译所有平台
./build.sh
```

### 4. 部署插件

#### 方法 1: 单平台部署

```bash
# 将编译好的可执行文件和 plugin.yaml 放在同一目录
plugins/
└── my-plugin/
    ├── plugin.yaml
    └── my-plugin        # Linux/macOS
    # 或 my-plugin.exe   # Windows
```

#### 方法 2: 多平台部署

```bash
# 使用 build.sh 构建所有平台
./build.sh

# 部署时包含所有平台的可执行文件
plugins/
└── my-plugin/
    ├── plugin.yaml
    ├── my-plugin           # Linux/macOS 可执行文件
    └── my-plugin.exe       # Windows 可执行文件
```

plugin.yaml 中配置平台路径：

```yaml
platform:
  windows_amd64: my-plugin.exe
  linux_amd64: my-plugin
  darwin_amd64: my-plugin
  darwin_arm64: my-plugin
```

### 5. 加载插件

在主程序控制台中执行：

```bash
reload
```

## SDK 功能

### 控制台命令

```go
ctx.RegisterConsoleCommand(sdk.ConsoleCommand{
    Name:        "mycmd",
    Triggers:    []string{"mycmd", "mc"},
    Usage:       "mycmd <arg>",
    Description: "我的命令",
    Handler: func(args []string) error {
        return nil
    },
})
```

### 游戏事件监听

```go
// 玩家加入
ctx.ListenPlayerJoin(func(event sdk.PlayerEvent) {
    ctx.LogInfo("玩家 %s 加入", event.Name)
})

// 玩家离开
ctx.ListenPlayerLeave(func(event sdk.PlayerEvent) {
    ctx.LogInfo("玩家 %s 离开", event.Name)
})

// 聊天消息
ctx.ListenChat(func(event *sdk.ChatEvent) {
    ctx.LogInfo("%s: %s", event.Sender, event.Message)
})
```

### 游戏控制

```go
// 获取 GameUtils
gu := ctx.GameUtils()

// 发送消息
gu.SayTo("玩家名", "§a你好！")
gu.SendText("全体消息")

// 执行命令
gu.SendCommand("/give @a diamond 64")
```

### 日志输出

```go
ctx.LogInfo("信息日志")      // 蓝色
ctx.LogSuccess("成功日志")   // 绿色
ctx.LogWarning("警告日志")   // 黄色
ctx.LogError("错误日志")     // 红色
```

### 数据存储

```go
// 配置文件
config := ctx.Config()
config.SetDefault("key", "value")
value := config.GetString("key")

// 临时数据
temp := ctx.TempJSON()
temp.Set("data", myData)
myData := temp.Get("data")
```

## 构建说明

### 构建脚本

`build.sh` 会自动构建以下平台：

- Windows (amd64, arm64)
- Linux (amd64, arm64)
- macOS (amd64, arm64/M1)
- Android (arm64)

### 手动构建

```bash
# Windows AMD64
GOOS=windows GOARCH=amd64 go build -o my-plugin.exe

# Linux AMD64
GOOS=linux GOARCH=amd64 go build -o my-plugin

# macOS ARM64 (M1/M2)
GOOS=darwin GOARCH=arm64 go build -o my-plugin
```

## 注意事项

1. **插件必须有 main() 函数**，作为独立可执行文件运行
2. **使用 plugin.Serve() 启动 gRPC 服务器**
3. **CGO_ENABLED=0** 确保跨平台兼容
4. **实现所有 Plugin 接口方法**：Init, Start, Stop, GetInfo

## 性能

- 启动时间：+50-100ms（启动独立进程）
- 调用延迟：~0.1-0.5ms（gRPC 通信）
- 内存开销：每个插件 +10-20MB（独立进程）

## 问题排查

### 插件无法加载

1. 检查可执行文件是否存在且有执行权限
2. 检查 plugin.yaml 中的 platform 配置是否正确
3. 查看主程序日志中的错误信息

### gRPC 连接失败

1. 确保使用了正确的 HandshakeConfig
2. 检查 plugin.Serve() 配置是否正确
3. 查看插件进程是否正常启动

### 编译错误

1. 确保 go.mod 中的依赖版本正确
2. 执行 `go mod tidy` 更新依赖
3. 检查 SDK 路径是否正确
