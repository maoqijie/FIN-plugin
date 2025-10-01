# ToolDelta 功能对比与待实现清单

## 已实现的核心功能 ✅

### 基础插件系统
- [x] 插件生命周期（Init, Start, Stop）
- [x] 插件热加载/重载
- [x] 插件配置管理
- [x] 插件数据目录管理
- [x] 控制台命令注册

### 事件监听
- [x] ListenPreload - 预加载事件
- [x] ListenActive - 激活事件
- [x] ListenPlayerJoin - 玩家加入
- [x] ListenPlayerLeave - 玩家离开
- [x] ListenChat - 聊天消息
- [x] ListenFrameExit - 框架退出
- [x] ListenPacket - 数据包监听

### 游戏工具 API
- [x] SendCommand - 发送命令
- [x] SendChat - 发送聊天消息
- [x] SayTo - 向玩家发送消息
- [x] GetScore - 获取计分板分数
- [x] GetItem - 获取玩家物品数量
- [x] GetPosition - 获取玩家坐标
- [x] GetPosXYZ - 获取简单坐标
- [x] TestSelector - 测试目标选择器

### 插件间通信
- [x] RegisterPluginAPI - 注册插件 API
- [x] GetPluginAPI - 获取插件 API
- [x] ListPluginAPIs - 列出所有 API

### 其他
- [x] 消息拦截机制（CancelMessage）
- [x] 日志输出（ctx.Logf）

---

## 待实现的功能 ⏳

### 1. 高级游戏工具 API

#### 1.1 目标选择器相关
```python
# ToolDelta 实现
def getTarget(sth: str, timeout: float = 5) -> list[str]:
    """获取符合目标选择器实体的列表"""
```

**建议实现**:
```go
// SDK 方法
func (g *GameUtils) GetTargets(selector string, timeout float64) ([]string, error)

// 示例
targets, err := ctx.GameUtils().GetTargets("@a[r=10]", 5.0)
```

#### 1.2 方块和区域操作
```python
# ToolDelta 实现
def getBlockTile(x: int, y: int, z: int) -> str:
    """获取指定坐标的方块类型"""

def getTickingAreaList() -> dict:
    """获取常加载区块列表"""
```

**建议实现**:
```go
func (g *GameUtils) GetBlock(x, y, z int) (string, error)
func (g *GameUtils) GetTickingAreas() (map[string]interface{}, error)
```

#### 1.3 玩家背包操作
```python
# ToolDelta 实现
def queryPlayerInventory(selector: str) -> dict:
    """查询玩家背包信息"""
```

**建议实现**:
```go
type InventorySlot struct {
    Slot     int
    ItemID   string
    ItemName string
    Count    int
    Aux      int
}

func (g *GameUtils) GetInventory(selector string) ([]InventorySlot, error)
```

#### 1.4 效果管理
```python
# ToolDelta 实现
def set_player_effect(
    target: str,
    effect_id: int,
    duration: int = 1000000,
    level: int = 0,
    hideParticle: bool = True,
) -> None:
    """给玩家添加药水效果"""
```

**建议实现**:
```go
type EffectOptions struct {
    Duration      int
    Level         int
    HideParticles bool
}

func (g *GameUtils) SetEffect(target string, effectID int, opts EffectOptions) error
```

#### 1.5 展示框操作
```python
# ToolDelta 实现
def take_item_out_item_frame(pos: tuple[float, float, float]) -> None:
    """从展示框取出物品"""
```

**建议实现**:
```go
func (g *GameUtils) TakeItemFromFrame(x, y, z float32) error
```

---

### 2. 事件优先级系统 ⭐

ToolDelta 的所有事件监听都支持优先级参数：

```python
# ToolDelta 实现
def ListenChat(self, cb: Callable[[Chat], Any], priority: int = 0):
    """priority 越高越先执行"""
```

**建议实现**:
```go
// SDK 修改
type ChatHandlerEntry struct {
    Handler  ChatHandler
    Priority int
}

func (c *Context) ListenChatWithPriority(handler ChatHandler, priority int) error

// 示例
ctx.ListenChatWithPriority(func(event *sdk.ChatEvent) {
    // 处理逻辑
}, 100) // 高优先级
```

**实现要点**:
- 在 PluginManager 中按优先级排序处理器
- 高优先级先执行，可以拦截事件阻止低优先级处理

---

### 3. 广播事件系统 ⭐⭐

ToolDelta 的广播机制允许插件间异步通信：

```python
# ToolDelta 实现
def ListenInternalBroadcast(
    self,
    broadcast_name: str,
    cb: Callable[[InternalBroadcast], Any],
    priority: int = 0,
):
    """监听广播事件"""

def BroadcastEvent(self, evt: InternalBroadcast):
    """广播事件给所有监听者"""
```

**建议实现**:
```go
// SDK 类型定义
type Broadcast struct {
    Name string
    Data map[string]interface{}
}

type BroadcastHandler func(Broadcast) interface{}

// Context 方法
func (c *Context) ListenBroadcast(name string, handler BroadcastHandler) error
func (c *Context) Broadcast(broadcast Broadcast) []interface{}

// 示例：插件 A 广播事件
ctx.Broadcast(Broadcast{
    Name: "player.teleport",
    Data: map[string]interface{}{
        "player": "Steve",
        "from": [3]float32{100, 64, 100},
        "to": [3]float32{200, 64, 200},
    },
})

// 示例：插件 B 监听事件
ctx.ListenBroadcast("player.teleport", func(evt Broadcast) interface{} {
    player := evt.Data["player"].(string)
    // 处理传送逻辑
    return nil // 或返回数据
})
```

**使用场景**:
- 传送系统插件广播传送事件
- 经济插件广播交易事件
- 权限插件广播权限变更事件

---

### 4. 消息等待机制 ⭐

ToolDelta 提供了等待玩家回复的便捷方法：

```python
# ToolDelta 实现
def waitMsg(playername: str, timeout: float = 30) -> str | None:
    """等待玩家发送消息"""
```

**建议实现**:
```go
// SDK 方法
func (c *Context) WaitMessage(playerName string, timeout time.Duration) (string, error)

// 示例：等待玩家输入
ctx.GameUtils().SayTo(player, "请输入你的选择：")
choice, err := ctx.WaitMessage(player, 30*time.Second)
if err != nil {
    ctx.GameUtils().SayTo(player, "§c超时未回复")
    return
}
```

**实现方式**:
- 使用 channel 和超时机制
- 在聊天事件处理器中检查等待队列
- 支持超时自动清理

---

### 5. 二进制数据包监听 ⭐

ToolDelta 区分了字典数据包和二进制数据包：

```python
# ToolDelta 实现
def ListenBytesPacket(
    self,
    pkID: PacketIDS | list[PacketIDS],
    cb: BytesPacketListener,
    priority: int = 0,
):
    """监听二进制数据包"""
```

**建议实现**:
```go
// SDK 类型
type BytesPacketHandler func(packetID uint32, data []byte) bool

// Context 方法
func (c *Context) ListenBytesPacket(handler BytesPacketHandler, packetIDs ...uint32) error

// 示例
ctx.ListenBytesPacket(func(packetID uint32, data []byte) bool {
    // 处理原始字节数据
    return false // true 表示拦截
}, packet.IDSomePacket)
```

---

### 6. 多重命令类型支持

ToolDelta 支持多种命令发送方式：

```python
# ToolDelta 实现
def sendcmd(cmd: str, waitForResp: bool = False, timeout: float = 30)
    """通过玩家身份发送命令"""

def sendwscmd(cmd: str, timeout: float = 30)
    """通过 WebSocket 身份发送命令"""

def sendwocmd(cmd: str)
    """通过 Settings 通道发送命令"""
```

**当前实现**: 已有 SendCommand, SendWSCommand 等
**建议**: 文档化不同命令类型的用途和区别

---

### 7. 打印系统增强

ToolDelta 提供了多种日志级别：

```python
# ToolDelta 实现
self.print       # 普通输出
self.print_inf   # 信息
self.print_suc   # 成功
self.print_war   # 警告
self.print_err   # 错误
```

**建议实现**:
```go
// Context 方法
func (c *Context) LogInfo(format string, args ...interface{})
func (c *Context) LogSuccess(format string, args ...interface{})
func (c *Context) LogWarning(format string, args ...interface{})
func (c *Context) LogError(format string, args ...interface{})

// 实现不同颜色的输出
// INFO: 蓝色背景
// SUCCESS: 绿色背景
// WARNING: 黄色背景
// ERROR: 红色背景
```

---

### 8. 权限检查 API

```python
# ToolDelta 实现
def is_op(playername: str) -> bool:
    """检查玩家是否为 OP"""
```

**建议实现**:
```go
func (g *GameUtils) IsOp(playerName string) (bool, error)

// 更强大的权限系统
type PermissionLevel int

const (
    PermissionMember PermissionLevel = 0
    PermissionOperator PermissionLevel = 1
    PermissionHost PermissionLevel = 2
    PermissionOwner PermissionLevel = 3
    PermissionInternal PermissionLevel = 4
)

func (g *GameUtils) GetPermissionLevel(playerName string) (PermissionLevel, error)
```

---

### 9. 多目标分数获取

```python
# ToolDelta 实现
def getMultiScore(scoreboardNameToGet: str, targetNameToGet: str) -> int | dict:
    """获取单个或多个目标的分数"""
```

**建议实现**:
```go
func (g *GameUtils) GetMultiScore(scoreboard string, targets []string) (map[string]int, error)

// 示例
scores, err := ctx.GameUtils().GetMultiScore("money", []string{"Steve", "Alex", "Bob"})
// 返回: {"Steve": 1000, "Alex": 500, "Bob": 200}
```

---

### 10. 通知服务器方法

```python
# ToolDelta 实现
def notifyToServer(
    message: str = "",
    sound: str = "random.orb",
    subtitle: str | None = None,
) -> None:
    """向所有玩家显示通知"""
```

**建议实现**:
```go
type NotificationOptions struct {
    Message  string
    Sound    string
    Subtitle string
}

func (g *GameUtils) NotifyAll(opts NotificationOptions) error

// 示例
ctx.GameUtils().NotifyAll(NotificationOptions{
    Message:  "§a新玩家加入服务器",
    Sound:    "random.orb",
    Subtitle: "欢迎！",
})
```

---

## 实现优先级建议

### 高优先级 ⭐⭐⭐
1. **事件优先级系统** - 对插件执行顺序至关重要
2. **WaitMessage 机制** - 简化交互式插件开发
3. **GetTargets 方法** - 扩展目标选择器功能

### 中优先级 ⭐⭐
4. **广播事件系统** - 增强插件间通信
5. **日志级别增强** - 改善调试体验
6. **GetInventory 方法** - 背包管理功能

### 低优先级 ⭐
7. **GetBlock 方法** - 方块检测功能
8. **SetEffect 方法** - 药水效果管理
9. **二进制数据包监听** - 高级数据包处理
10. **其他辅助方法**

---

## 架构改进建议

### 1. 统一错误处理
```go
// 定义标准错误类型
var (
    ErrTimeout       = errors.New("operation timeout")
    ErrPlayerOffline = errors.New("player is offline")
    ErrCommandFailed = errors.New("command execution failed")
)
```

### 2. Context 超时支持
```go
// 所有需要等待的操作支持 context
func (g *GameUtils) GetScoreWithContext(ctx context.Context, scoreboard, target string) (int, error)
```

### 3. 事件拦截增强
```go
// 事件处理器返回值表示是否拦截
type ChatHandler func(*ChatEvent) bool

// true: 拦截事件，停止传播
// false: 继续传播给其他处理器
```

---

## 文档完善计划

1. ✅ 商店插件教程（已完成）
2. ⏳ 事件系统完整文档
3. ⏳ GameUtils API 参考手册
4. ⏳ 插件间通信最佳实践
5. ⏳ 常见插件开发模式
6. ⏳ 性能优化指南

---

## 总结

当前 FunInterwork PluginFramework 已经实现了 ToolDelta 的核心功能（约 60-70%）。待实现的功能主要集中在：

1. **高级游戏操作 API** - 背包、方块、效果等
2. **事件系统增强** - 优先级、广播、拦截
3. **便捷交互方法** - WaitMessage、多级日志
4. **插件间通信** - 广播事件系统

建议按照优先级逐步实现，同时持续完善文档和示例代码。
