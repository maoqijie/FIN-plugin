package sdk

import (
	"fmt"
	"regexp"
	"strings"
)

// Translator 提供游戏文本翻译功能，类似 ToolDelta 的 mc_translator
type Translator struct {
	translations map[string]string
}

// NewTranslator 创建翻译器实例
func NewTranslator() *Translator {
	return &Translator{
		translations: getBuiltinTranslations(),
	}
}

// Translate 将游戏文本翻译为中文
// key: 要翻译的消息文本键（如 "item.diamond.name"）
// args: 可选的翻译参数列表
// translateArgs: 是否翻译参数项，默认 false
//
// 示例:
//   // 简单翻译
//   msg := translator.Translate("item.diamond.name", nil, false)
//   // 返回: "钻石"
//
//   // 带参数翻译
//   msg := translator.Translate("death.attack.anvil", []string{"SkyblueSuper"}, false)
//   // 返回: "SkyblueSuper 被坠落的铁砧压扁了"
//
//   // 翻译参数
//   msg := translator.Translate("commands.enchant.invalidLevel", []interface{}{"enchantment.mending", 6}, true)
//   // 返回: "经验修补 不支持等级 6"
func (t *Translator) Translate(key string, args []interface{}, translateArgs bool) string {
	// 获取翻译文本
	translated, exists := t.translations[key]
	if !exists {
		// 如果没有翻译，返回原始键
		translated = key
	}

	// 如果没有参数，直接返回翻译文本
	if len(args) == 0 {
		return translated
	}

	// 处理参数翻译
	processedArgs := make([]interface{}, len(args))
	for i, arg := range args {
		if translateArgs {
			// 如果参数是字符串且以 % 开头，尝试翻译
			if strArg, ok := arg.(string); ok {
				if strings.HasPrefix(strArg, "%") {
					// 移除 % 前缀并翻译
					argKey := strings.TrimPrefix(strArg, "%")
					if argTranslated, found := t.translations[argKey]; found {
						processedArgs[i] = argTranslated
					} else {
						processedArgs[i] = argKey
					}
				} else {
					processedArgs[i] = arg
				}
			} else {
				processedArgs[i] = arg
			}
		} else {
			processedArgs[i] = arg
		}
	}

	// 替换参数占位符
	result := translated

	// Minecraft 使用 %s 和 %1$s、%2$s 等格式的占位符
	// 先处理位置参数（%1$s, %2$s）
	for i, arg := range processedArgs {
		placeholder := fmt.Sprintf("%%%d$s", i+1)
		result = strings.ReplaceAll(result, placeholder, fmt.Sprint(arg))
	}

	// 处理简单的 %s 占位符（按顺序替换）
	for _, arg := range processedArgs {
		result = strings.Replace(result, "%s", fmt.Sprint(arg), 1)
	}

	return result
}

// TranslateSimple 简化的翻译方法，不支持参数
func (t *Translator) TranslateSimple(key string) string {
	return t.Translate(key, nil, false)
}

// TranslateWithArgs 带参数的翻译方法
func (t *Translator) TranslateWithArgs(key string, args ...interface{}) string {
	return t.Translate(key, args, false)
}

// Has 检查是否存在某个翻译键
func (t *Translator) Has(key string) bool {
	_, exists := t.translations[key]
	return exists
}

// AddTranslation 添加自定义翻译
func (t *Translator) AddTranslation(key, value string) {
	t.translations[key] = value
}

// AddTranslations 批量添加翻译
func (t *Translator) AddTranslations(translations map[string]string) {
	for key, value := range translations {
		t.translations[key] = value
	}
}

// LoadFromLangFile 从 .lang 格式文件加载翻译（Minecraft 语言文件格式）
// 格式: key=value
func (t *Translator) LoadFromLangFile(content string) error {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// 跳过空行和注释
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// 解析 key=value
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		t.translations[key] = value
	}
	return nil
}

// getBuiltinTranslations 返回内置的中文翻译表
// 这里包含常用的 Minecraft 文本翻译
func getBuiltinTranslations() map[string]string {
	return map[string]string{
		// 物品名称
		"item.diamond.name":          "钻石",
		"item.emerald.name":          "绿宝石",
		"item.iron_ingot.name":       "铁锭",
		"item.gold_ingot.name":       "金锭",
		"item.stick.name":            "木棍",
		"item.apple.name":            "苹果",
		"item.golden_apple.name":     "金苹果",
		"item.bow.name":              "弓",
		"item.arrow.name":            "箭",
		"item.diamond_sword.name":    "钻石剑",
		"item.diamond_pickaxe.name":  "钻石镐",
		"item.diamond_axe.name":      "钻石斧",
		"item.diamond_shovel.name":   "钻石锹",
		"item.diamond_hoe.name":      "钻石锄",
		"item.bread.name":            "面包",
		"item.cooked_beef.name":      "熟牛肉",
		"item.cooked_porkchop.name":  "熟猪排",

		// 方块名称
		"tile.dirt.name":             "泥土",
		"tile.grass.name":            "草方块",
		"tile.stone.name":            "石头",
		"tile.wood.name":             "木板",
		"tile.bedrock.name":          "基岩",
		"tile.sand.name":             "沙子",
		"tile.gravel.name":           "沙砾",
		"tile.oreGold.name":          "金矿石",
		"tile.oreIron.name":          "铁矿石",
		"tile.oreDiamond.name":       "钻石矿石",
		"tile.log.name":              "原木",
		"tile.glass.name":            "玻璃",
		"tile.obsidian.name":         "黑曜石",

		// 死亡消息
		"death.attack.anvil":         "%s 被坠落的铁砧压扁了",
		"death.attack.arrow":         "%s 被 %s 射杀",
		"death.attack.cactus":        "%s 被戳死了",
		"death.attack.drown":         "%s 溺水身亡",
		"death.attack.explosion":     "%s 爆炸了",
		"death.attack.fall":          "%s 落地过猛",
		"death.attack.fallingBlock":  "%s 被坠落的方块压扁了",
		"death.attack.fireball":      "%s 被 %s 的火球烧死了",
		"death.attack.lava":          "%s 试图在熔岩里游泳",
		"death.attack.lightningBolt": "%s 被闪电击中",
		"death.attack.mob":           "%s 被 %s 杀死了",
		"death.attack.player":        "%s 被 %s 杀死了",
		"death.attack.starve":        "%s 饿死了",
		"death.attack.wither":        "%s 凋零了",

		// 命令消息
		"commands.generic.syntax":         "无效的命令语法。",
		"commands.generic.player.notFound": "找不到玩家",
		"commands.generic.num.invalid":    "「%s」是无效的数字",
		"commands.generic.num.tooBig":     "你输入的数字（%s）太大，它最大只能为 %s",
		"commands.generic.num.tooSmall":   "你输入的数字（%s）太小，它最小只能为 %s",
		"commands.enchant.noItem":         "目标没有手持物品",
		"commands.enchant.notFound":       "没有 ID 为 %s 的附魔",
		"commands.enchant.invalidLevel":   "%s 不支持等级 %s",
		"commands.enchant.success":        "正在附魔",
		"commands.give.success":           "给予了 %s * %s 给 %s",
		"commands.gamemode.success.self":  "将自己的游戏模式设置成了%s模式",
		"commands.gamemode.success.other": "将 %s 的游戏模式设置成了%s模式",
		"commands.tp.success":             "已将 %s 传送到 %s",
		"commands.kill.successful":        "杀死了 %s",

		// 附魔
		"enchantment.mending":       "经验修补",
		"enchantment.unbreaking":    "耐久",
		"enchantment.sharpness":     "锋利",
		"enchantment.protection":    "保护",
		"enchantment.efficiency":    "效率",
		"enchantment.silk_touch":    "精准采集",
		"enchantment.fortune":       "时运",
		"enchantment.looting":       "抢夺",
		"enchantment.knockback":     "击退",
		"enchantment.fire_aspect":   "火焰附加",
		"enchantment.flame":         "火矢",
		"enchantment.power":         "力量",
		"enchantment.punch":         "冲击",
		"enchantment.infinity":      "无限",
		"enchantment.thorns":        "荆棘",
		"enchantment.aqua_affinity": "水下速掘",
		"enchantment.respiration":   "水下呼吸",
		"enchantment.depth_strider": "深海探索者",

		// 游戏模式
		"gameMode.survival":  "生存",
		"gameMode.creative":  "创造",
		"gameMode.adventure": "冒险",
		"gameMode.spectator": "旁观",

		// 其他常用
		"multiplayer.player.joined": "%s 加入了游戏",
		"multiplayer.player.left":   "%s 离开了游戏",
		"chat.type.text":            "<%s> %s",
		"chat.type.announcement":    "[%s] %s",
	}
}

// TranslateItemName 翻译物品名称（便捷方法）
func (t *Translator) TranslateItemName(itemID string) string {
	key := fmt.Sprintf("item.%s.name", itemID)
	return t.TranslateSimple(key)
}

// TranslateBlockName 翻译方块名称（便捷方法）
func (t *Translator) TranslateBlockName(blockID string) string {
	key := fmt.Sprintf("tile.%s.name", blockID)
	return t.TranslateSimple(key)
}

// TranslateEnchantment 翻译附魔名称（便捷方法）
func (t *Translator) TranslateEnchantment(enchantID string) string {
	key := fmt.Sprintf("enchantment.%s", enchantID)
	return t.TranslateSimple(key)
}

// ParseColorCodes 解析并替换颜色代码（Minecraft 格式 §）
// 这个方法可以移除或转换颜色代码
func (t *Translator) ParseColorCodes(text string, stripCodes bool) string {
	if stripCodes {
		// 移除所有颜色代码（§ + 单个字符）
		re := regexp.MustCompile(`§.`)
		return re.ReplaceAllString(text, "")
	}
	// 可以在这里添加颜色代码转换逻辑（如转为 ANSI 颜色）
	return text
}

// StripColorCodes 移除所有颜色代码
func (t *Translator) StripColorCodes(text string) string {
	return t.ParseColorCodes(text, true)
}
