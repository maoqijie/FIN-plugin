# 快速开始

本指南将帮助你在 5 分钟内创建并运行第一个 FunInterWork 插件。

## 前置条件

- Go 1.25.1 或更高版本
- Linux 或 macOS 系统（Windows 不支持 Go plugin）
- FunInterWork 主程序

## 创建插件

### 1. 使用模板创建

```bash
# 进入插件目录
cd Plugin

# 从模板复制
cp -r ../PluginFramework/templates/bot/info ./my-first-plugin
cd my-first-plugin
```

### 2. 编辑插件代码

创建 `main.go`：

```go
package main

import (
    sdk "github.com/maoqijie/FIN-plugin/sdk"
)

type MyPlugin struct {
    ctx *sdk.Context
}

func (p *MyPlugin) Init(ctx *sdk.Context) error {
    p.ctx = ctx

    // 注册控制台命令
    ctx.RegisterConsoleCommand(sdk.ConsoleCommand{
        Triggers:    []string{"hello"},
        Usage:       "测试插件命令",
        Description: "输出 Hello World",
        Handler: func(args []string) error {
            p.ctx.Logf("Hello World!")
            return nil
        },
    })

    // 监听聊天事件
    ctx.ListenChat(func(evt sdk.ChatEvent) {
        p.ctx.Logf("收到消息: [%s] %s", evt.Sender, evt.Message)
    })

    return nil
}

func (p *MyPlugin) Start() error {
    p.ctx.Logf("插件已启动")
    return nil
}

func (p *MyPlugin) Stop() error {
    p.ctx.Logf("插件已停止")
    return nil
}

// NewPlugin 必须导出此函数
func NewPlugin() sdk.Plugin {
    return &MyPlugin{}
}
```

### 3. 创建 go.mod

```bash
go mod init my-first-plugin
```

`go.mod` 内容：

```go
module my-first-plugin

go 1.25.1

require github.com/maoqijie/FIN-plugin v0.0.0

// 使用本地 SDK（开发时）
replace github.com/maoqijie/FIN-plugin => ../../PluginFramework
```

## 运行插件

### 1. 启动主程序

```bash
# 主程序会自动编译并加载插件
./main --config config.json
```

### 2. 测试插件

在控制台输入：

```bash
# 测试插件命令
hello

# 查看所有插件命令
?
```

### 3. 触发聊天事件

在 Minecraft 中发送消息，控制台会显示：

```
[my-first-plugin] 收到消息: [Steve] Hello!
```

## 下一步

- 查看 [插件结构](plugin-structure.md) 了解插件组织方式
- 查看 [开发流程](development-workflow.md) 了解调试和热重载
- 查看 [API 参考](../api/) 了解更多功能
- 查看 [示例插件](../../templates/) 学习实际案例

## 常见问题

### 插件没有被加载？

1. 检查 `Plugin/<插件名>/` 目录结构是否正确
2. 检查 `go.mod` 中的 replace 路径是否正确
3. 查看主程序日志中的错误信息

### 编译失败？

1. 确保 Go 版本为 1.25.1+
2. 确保 PluginFramework 子模块已正确初始化
3. 检查 import 路径是否正确

### 热重载不生效？

在控制台输入 `reload` 命令重新加载所有插件。
