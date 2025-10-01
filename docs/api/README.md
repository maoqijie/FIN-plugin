# API 快速参考

本页面提供所有 API 的快速索引。详细文档请查看各模块的文档文件。

## Context - 上下文

**文件位置**：`sdk/plugin.go`

### 信息获取

| 方法 | 返回类型 | 说明 |
|------|---------|------|
| `PluginName()` | `string` | 获取插件名称 |
| `BotInfo()` | `BotInfo` | 获取机器人信息（昵称、XUID、实体ID） |
| `ServerInfo()` | `ServerInfo` | 获取租赁服信息（服务器代码、口令） |
| `QQInfo()` | `QQInfo` | 获取 QQ 适配器信息（WS地址、Token） |
| `InterworkInfo()` | `InterworkInfo` | 获取互通群组信息 |

### 功能模块

| 方法 | 返回类型 | 说明 |
|------|---------|------|
| `GameUtils()` | `*GameUtils` | 游戏交互接口 |
| `PlayerManager()` | `*PlayerManager` | 玩家管理器 |
| `PacketWaiter()` | `*PacketWaiter` | 数据包等待器 |
| `Utils()` | `*Utils` | 实用工具 |
| `Translator()` | `*Translator` | 文本翻译器 |
| `Console()` | `*Console` | 控制台输出 |
| `Config()` | `*Config` | 配置管理 |
| `TempJSON()` | `*TempJSON` | JSON 缓存 |

### 事件监听

| 方法 | 参数 | 说明 |
|------|------|------|
| `ListenPreload(handler)` | `func()` | 插件预加载事件 |
| `ListenActive(handler)` | `func()` | 激活事件 |
| `ListenPlayerJoin(handler)` | `func(PlayerEvent)` | 玩家加入 |
| `ListenPlayerLeave(handler)` | `func(PlayerEvent)` | 玩家离开 |
| `ListenChat(handler)` | `func(ChatEvent)` | 聊天消息 |
| `ListenFrameExit(handler)` | `func(FrameExitEvent)` | 框架退出 |
| `ListenPacket(handler, ids...)` | `func(PacketEvent), []uint32` | 监听指定数据包 |
| `ListenPacketAll(handler)` | `func(PacketEvent)` | 监听所有数据包（不推荐） |

### 插件 API

| 方法 | 说明 |
|------|------|
| `RegisterPluginAPI(name, version, plugin)` | 注册为 API 插件 |
| `GetPluginAPI(name)` | 获取 API 插件 |
| `GetPluginAPIWithVersion(name, version)` | 获取指定版本的 API |
| `ListPluginAPIs()` | 列出所有 API 插件 |

### 数据管理

| 方法 | 说明 |
|------|------|
| `DataPath()` | 获取插件数据目录（自动创建） |
| `FormatDataPath(path...)` | 格式化数据文件路径 |

### 其他

| 方法 | 说明 |
|------|------|
| `RegisterConsoleCommand(cmd)` | 注册控制台命令 |
| `Logf(format, args...)` | 输出日志 |

---

## GameUtils - 游戏交互

**文件位置**：`sdk/game_utils.go`

### 查询方法

| 方法 | 返回值 | 说明 |
|------|--------|------|
| `GetTarget(target, timeout)` | `[]string, error` | 获取目标选择器匹配的玩家 |
| `GetPos(target)` | `*Position, error` | 获取玩家详细坐标 |
| `GetPosXYZ(target)` | `float32, float32, float32, error` | 获取玩家简单坐标 |
| `GetItem(target, item, specialID)` | `int, error` | 获取物品数量 |
| `GetScore(scbName, target, timeout)` | `int, error` | 获取计分板分数 |
| `IsCmdSuccess(cmd, timeout)` | `bool, error` | 检查命令是否成功 |
| `IsOp(playerName)` | `bool, error` | 检查是否为管理员 |

### 命令发送

| 方法 | 说明 |
|------|------|
| `SendCommand(cmd)` | 发送游戏命令 |
| `SendCommandWithResponse(cmd, timeout)` | 发送命令并等待响应 |
| `SendWOCommand(cmd)` | 发送高权限控制台命令 |
| `SendPacket(packetID, packet)` | 发送网络数据包 |

### 消息发送

| 方法 | 说明 |
|------|------|
| `SendChat(message)` | 机器人发言 |
| `Title(message)` | ActionBar 全体消息 |
| `Tellraw(selector, message)` | Tellraw 消息 |
| `SayTo(target, text)` | 向目标发送消息 |
| `PlayerTitle(target, text)` | 显示标题 |
| `PlayerSubtitle(target, text)` | 显示副标题 |
| `PlayerActionbar(target, text)` | 显示 ActionBar |

---

## PlayerManager - 玩家管理

**文件位置**：`sdk/player.go`

### 查询方法

| 方法 | 返回值 | 说明 |
|------|--------|------|
| `GetAllPlayers()` | `[]*Player` | 获取所有在线玩家 |
| `GetBotInfo()` | `*Player` | 获取机器人信息 |
| `GetPlayerByName(name)` | `*Player` | 根据名称查找玩家 |
| `GetPlayerByUUID(uuid)` | `*Player` | 根据 UUID 查找 |
| `GetPlayerByUniqueID(id)` | `*Player` | 根据实体 ID 查找 |
| `GetPlayerCount()` | `int` | 获取在线人数 |

### Player 对象方法

#### 交互
- `Show(text)` - 发送消息
- `SetTitle(title, subtitle)` - 显示标题
- `SetActionBar(text)` - 显示 ActionBar

#### 查询
- `GetPos()` - 获取坐标
- `GetPosXYZ()` - 获取简单坐标
- `GetScore(scbName, timeout)` - 获取分数
- `GetItemCount(itemName, specialID)` - 获取物品数量
- `IsOp()` - 检查权限

#### 操作
- `Teleport(x, y, z)` - 传送
- `TeleportTo(playerName)` - 传送到玩家
- `SetGameMode(mode)` - 设置游戏模式
- `GiveItem(itemName, amount, data)` - 给予物品
- `ClearItem(itemName, maxCount)` - 清除物品
- `AddEffect(effect, duration, amplifier, hideParticles)` - 给予效果
- `ClearEffects()` - 清除所有效果
- `Kill()` - 杀死玩家
- `Kick(reason)` - 踢出玩家

---

## Utils - 实用工具

**文件位置**：`sdk/utils.go`

### 字符串与格式化
- `SimpleFormat(data, template)` - 占位符替换
- `ToPlayerSelector(playerName)` - 玩家名转选择器

### 类型转换
- `TryInt(input)` - 转换为整数

### 列表操作
- `FillListIndex(list, reference, defaultValue)` - 填充列表
- `FillStringList(list, referenceLen, defaultValue)` - 填充字符串列表
- `FillIntList(list, referenceLen, defaultValue)` - 填充整数列表

### 异步与并发
- `CreateResultCallback(timeout)` - 创建回调锁
- `RunAsync(fn)` - 异步执行
- `RunAsyncWithResult(fn)` - 异步执行并返回结果
- `Gather(fns...)` - 并行执行多个函数

### 定时器
- `NewTimer(interval, fn)` - 创建定时器

### 其他
- `Sleep(seconds)` - 睡眠
- `Contains(slice, item)` - 检查包含
- `Max(a, b)` / `Min(a, b)` / `Clamp(value, min, max)` - 数值操作

---

## Config - 配置管理

**文件位置**：`sdk/config.go`

### 配置读写
- `GetConfig(fileName, defaultConfig)` - 获取简单配置
- `SaveConfig(fileName, config)` - 保存配置
- `GetPluginConfigAndVersion(fileName, defaultConfig, defaultVersion, validateFunc)` - 获取带版本的配置
- `UpgradePluginConfig(fileName, config, newVersion)` - 升级配置

### 配置验证
- `CheckAuto(value, rule)` - 自动验证单个值
- `ValidateConfig(config, standard)` - 批量验证配置

### 其他
- `GetConfigPath(fileName)` - 获取配置文件路径
- `ConfigExists(fileName)` - 检查配置是否存在
- `DeleteConfig(fileName)` - 删除配置

---

## TempJSON - JSON 缓存

**文件位置**：`sdk/tempjson.go`

### 快捷方法（推荐）
- `LoadAndRead(filePath, createIfNotExists, defaultData, timeout)` - 快速读取
- `LoadAndWrite(filePath, data, createIfNotExists, timeout)` - 快速写入

### 缓存管理
- `Load(filePath, createIfNotExists, defaultData, timeout)` - 加载到缓存
- `Read(filePath, copyData)` - 从缓存读取
- `Write(filePath, data)` - 写入缓存
- `Unload(filePath)` - 卸载缓存

### 批量操作
- `SaveAll()` - 保存所有缓存
- `UnloadAll()` - 卸载所有缓存
- `GetCachedPaths()` - 获取缓存列表
- `IsCached(filePath)` - 检查缓存状态

---

## Translator - 文本翻译

**文件位置**：`sdk/translator.go`

### 翻译方法
- `Translate(key, params, translateParams)` - 完整翻译
- `TranslateSimple(key)` - 简单翻译
- `TranslateWithArgs(key, args...)` - 带参数翻译

### 便捷翻译
- `TranslateItemName(itemID)` - 翻译物品名称
- `TranslateBlockName(blockID)` - 翻译方块名称
- `TranslateEnchantment(enchID)` - 翻译附魔名称

### 管理方法
- `AddTranslation(key, value)` - 添加翻译
- `AddTranslations(translations)` - 批量添加
- `LoadFromLangFile(filePath)` - 从 .lang 文件加载

### 颜色处理
- `ParseColorCodes(text)` - 解析颜色代码
- `StripColorCodes(text)` - 移除颜色代码

---

## Console - 控制台输出

**文件位置**：`sdk/console.go`

### 基础输出
- `PrintInf(text, addPrefix)` - 信息（蓝色）
- `PrintSuc(text, addPrefix)` - 成功（绿色）
- `PrintWar(text, addPrefix)` - 警告（黄色）
- `PrintErr(text, addPrefix)` - 错误（红色）
- `PrintLoad(text, addPrefix)` - 加载（紫色）

### 格式化
- `FmtInfo(text, info)` - 格式化提示
- `CleanPrint(text)` - 转换 Minecraft 颜色代码
- `CleanFmt(text)` - 仅转换颜色代码

### 高级输出
- `PrintWithColor(text, color)` - 指定颜色打印
- `PrintRainbow(text)` - 彩虹文本
- `PrintBox(text, boxChar)` - 边框文本
- `PrintProgress(current, total, barWidth)` - 进度条
- `PrintTable(headers, rows)` - 表格

### 光标控制
- `ClearLine()` - 清除当前行
- `MoveCursorUp(n)` / `MoveCursorDown(n)` - 移动光标
- `HideCursor()` / `ShowCursor()` - 隐藏/显示光标
- `ClearScreen()` - 清屏

---

## PacketWaiter - 数据包等待

**文件位置**：`sdk/packet_handler.go`

- `WaitNextPacket(packetID, timeout)` - 等待指定数据包
- `WaitNextPacketAny(timeout)` - 等待任意数据包（不推荐）
- `NotifyPacket(packetID, packet)` - 通知数据包到达（内部方法）
- `Clear()` - 清理所有等待器

---

## 详细文档

以上为 API 快速参考，详细说明、示例和最佳实践请查看各模块文档：

- [Context](context.md) - 上下文方法详解
- [事件监听](events.md) - 事件系统完整指南
- [GameUtils](game-utils.md) - 游戏交互详细文档
- [插件数据](data-management.md) - 数据目录管理详细指南
- [Utils](utils.md) - 工具方法使用指南
- [Translator](translator.md) - 文本翻译完整文档
- [Console](console.md) - 控制台输出详解
- [Config](config.md) - 配置管理完整指南
- [TempJSON](tempjson.md) - JSON 缓存详细文档
- [PlayerManager](player-manager.md) - 玩家管理完整指南
- [示例代码](../../templates/) - 实际可运行的示例
