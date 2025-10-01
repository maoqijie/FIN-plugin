    ```go
    results := utils.Gather(
        func() interface{} { return "结果1" },
        func() interface{} { return "结果2" },
        func() interface{} { return "结果3" },
    )
    ```

#### 定时器

- **NewTimer(interval float64, fn func())** - 创建定时器
  - `interval`：执行间隔（秒）
  - `fn`：要定时执行的函数
  - 返回：Timer 实例，支持 `Start()`、`Stop()`、`IsRunning()` 方法
  - 示例：
    ```go
    timer := utils.NewTimer(5.0, func() {
        fmt.Println("每 5 秒执行一次")
    })
    timer.Start()
    defer timer.Stop()
    ```

#### 其他工具方法

- **Sleep(seconds float64)** - 睡眠指定秒数
- **Contains(slice []string, item string)** - 检查字符串切片是否包含元素
- **ContainsInt(slice []int, item int)** - 检查整数切片是否包含元素
- **Max(a, b int)** - 返回较大值
- **Min(a, b int)** - 返回较小值
- **Clamp(value, min, max int)** - 将值限制在指定范围内

#### 完整示例

```go
func (p *plugin) Start() error {
    utils := p.ctx.Utils()

    // 字符串格式化
    msg := utils.SimpleFormat(map[string]string{
        "server": "我的服务器",
        "count": "10",
    }, "欢迎来到 {server}，当前在线 {count} 人")

    // 类型转换
    if count, ok := utils.TryInt("123"); ok {
        p.ctx.Logf("玩家数量: %d", count)
    }

    // 异步回调
    getter, setter := utils.CreateResultCallback(10.0)
    go func() {
        // 模拟异步操作
        utils.Sleep(3.0)
        setter("任务完成")
    }()

    if result, ok := getter(); ok {
        p.ctx.Logf("异步结果: %v", result)
    }

    // 定时任务
    timer := utils.NewTimer(60.0, func() {
        p.ctx.Logf("定时任务执行")
    })
    timer.Start()

    return nil
}
```

### Translator 游戏文本翻译

`sdk.Context.Translator()` 提供 Minecraft 游戏文本翻译功能，类似 ToolDelta 的 mc_translator。使用内置的中文翻译表将游戏文本键翻译为中文。

#### 核心方法

- **Translate(key string, args []interface{}, translateArgs bool)** - 翻译游戏文本
  - `key`：要翻译的消息文本键（如 `"item.diamond.name"`）
  - `args`：可选的翻译参数列表
  - `translateArgs`：是否翻译参数项
  - 示例：
    ```go
    // 简单翻译
    msg := translator.Translate("item.diamond.name", nil, false)
    // 返回: "钻石"

    // 带参数翻译
    msg := translator.Translate("death.attack.anvil", []interface{}{"SkyblueSuper"}, false)
    // 返回: "SkyblueSuper 被坠落的铁砧压扁了"

    // 翻译参数（参数以 % 开头会被翻译）
    msg := translator.Translate(
        "commands.enchant.invalidLevel",
        []interface{}{"enchantment.mending", 6},
        true,
    )
    // 返回: "经验修补 不支持等级 6"
    ```

#### 便捷方法

- **TranslateSimple(key string)** - 简单翻译，不带参数
- **TranslateWithArgs(key string, args ...interface{})** - 带参数翻译
- **TranslateItemName(itemID string)** - 翻译物品名称
- **TranslateBlockName(blockID string)** - 翻译方块名称
- **TranslateEnchantment(enchantID string)** - 翻译附魔名称

#### 管理方法

- **Has(key string)** - 检查是否存在某个翻译键
- **AddTranslation(key, value string)** - 添加自定义翻译
- **AddTranslations(translations map[string]string)** - 批量添加翻译
- **LoadFromLangFile(content string)** - 从 .lang 格式文件加载翻译

#### 颜色代码处理

- **ParseColorCodes(text string, stripCodes bool)** - 解析颜色代码
- **StripColorCodes(text string)** - 移除所有颜色代码（§ 格式）

#### 内置翻译

内置翻译表包含常用的中文翻译：
- 物品名称（钻石、绿宝石、铁锭等）