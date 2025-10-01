package sdk

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
)

// Config 提供插件配置文件管理功能，类似 ToolDelta 的配置管理
type Config struct {
	pluginName string
	configDir  string
}

// NewConfig 创建配置管理器
// pluginName: 插件名称
// configDir: 配置文件目录（可选，默认为 "./plugins/{pluginName}"）
func NewConfig(pluginName string, configDir ...string) *Config {
	dir := filepath.Join("plugins", pluginName)
	if len(configDir) > 0 && configDir[0] != "" {
		dir = configDir[0]
	}
	return &Config{
		pluginName: pluginName,
		configDir:  dir,
	}
}

// ConfigVersion 配置文件版本信息
type ConfigVersion struct {
	Major int // 主版本号
	Minor int // 次版本号
	Patch int // 修订版本号
}

// String 返回版本字符串 "major.minor.patch"
func (v ConfigVersion) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// ParseVersion 解析版本字符串
func ParseVersion(versionStr string) (ConfigVersion, error) {
	parts := strings.Split(versionStr, ".")
	if len(parts) != 3 {
		return ConfigVersion{}, fmt.Errorf("版本格式错误，应为 major.minor.patch")
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return ConfigVersion{}, fmt.Errorf("主版本号解析失败: %v", err)
	}

	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return ConfigVersion{}, fmt.Errorf("次版本号解析失败: %v", err)
	}

	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return ConfigVersion{}, fmt.Errorf("修订版本号解析失败: %v", err)
	}

	return ConfigVersion{Major: major, Minor: minor, Patch: patch}, nil
}

// Compare 比较版本大小
// 返回值: 1 表示 v > other，0 表示相等，-1 表示 v < other
func (v ConfigVersion) Compare(other ConfigVersion) int {
	if v.Major != other.Major {
		if v.Major > other.Major {
			return 1
		}
		return -1
	}
	if v.Minor != other.Minor {
		if v.Minor > other.Minor {
			return 1
		}
		return -1
	}
	if v.Patch != other.Patch {
		if v.Patch > other.Patch {
			return 1
		}
		return -1
	}
	return 0
}

// ConfigWithVersion 配置文件内容（包含版本）
type ConfigWithVersion struct {
	Config  map[string]interface{} `json:"config"`
	Version string                 `json:"version"`
}

// GetPluginConfigAndVersion 获取插件配置文件和版本
// fileName: 配置文件名（如 "config.json"）
// defaultConfig: 默认配置内容
// defaultVersion: 默认版本
// validateFunc: 可选的配置验证函数
//
// 返回: 配置内容、版本、错误
//
// 示例:
//   config, version, err := cfg.GetPluginConfigAndVersion(
//       "config.json",
//       map[string]interface{}{
//           "enable": true,
//           "max_count": 10,
//       },
//       ConfigVersion{Major: 1, Minor: 0, Patch: 0},
//       nil,
//   )
func (c *Config) GetPluginConfigAndVersion(
	fileName string,
	defaultConfig map[string]interface{},
	defaultVersion ConfigVersion,
	validateFunc func(map[string]interface{}) error,
) (map[string]interface{}, ConfigVersion, error) {
	// 确保配置目录存在
	if err := os.MkdirAll(c.configDir, 0755); err != nil {
		return nil, ConfigVersion{}, fmt.Errorf("创建配置目录失败: %v", err)
	}

	configPath := filepath.Join(c.configDir, fileName)

	// 检查配置文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// 配置文件不存在，创建默认配置
		if err := c.SaveConfigWithVersion(fileName, defaultConfig, defaultVersion); err != nil {
			return nil, ConfigVersion{}, fmt.Errorf("创建默认配置文件失败: %v", err)
		}
		return defaultConfig, defaultVersion, nil
	}

	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, ConfigVersion{}, fmt.Errorf("读取配置文件失败: %v", err)
	}

	// 解析配置文件
	var configWithVersion ConfigWithVersion
	if err := json.Unmarshal(data, &configWithVersion); err != nil {
		return nil, ConfigVersion{}, fmt.Errorf("解析配置文件失败: %v", err)
	}

	// 解析版本
	version, err := ParseVersion(configWithVersion.Version)
	if err != nil {
		// 版本解析失败，使用默认版本
		version = defaultVersion
	}

	// 验证配置
	if validateFunc != nil {
		if err := validateFunc(configWithVersion.Config); err != nil {
			return nil, ConfigVersion{}, fmt.Errorf("配置验证失败: %v", err)
		}
	}

	return configWithVersion.Config, version, nil
}

// SaveConfigWithVersion 保存配置文件和版本
func (c *Config) SaveConfigWithVersion(
	fileName string,
	config map[string]interface{},
	version ConfigVersion,
) error {
	// 确保配置目录存在
	if err := os.MkdirAll(c.configDir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %v", err)
	}

	configPath := filepath.Join(c.configDir, fileName)

	// 构建配置结构
	configWithVersion := ConfigWithVersion{
		Config:  config,
		Version: version.String(),
	}

	// 序列化为 JSON
	data, err := json.MarshalIndent(configWithVersion, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}

	// 写入文件
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %v", err)
	}

	return nil
}

// UpgradePluginConfig 升级插件配置文件
// fileName: 配置文件名
// newConfig: 新配置内容
// newVersion: 新版本号
//
// 示例:
//   err := cfg.UpgradePluginConfig(
//       "config.json",
//       map[string]interface{}{
//           "enable": true,
//           "max_count": 20,
//           "new_feature": "enabled",
//       },
//       ConfigVersion{Major: 1, Minor: 1, Patch: 0},
//   )
func (c *Config) UpgradePluginConfig(
	fileName string,
	newConfig map[string]interface{},
	newVersion ConfigVersion,
) error {
	return c.SaveConfigWithVersion(fileName, newConfig, newVersion)
}

// ValidatorFunc 配置验证函数类型
type ValidatorFunc func(interface{}) error

// TypeValidator 类型验证器
type TypeValidator struct {
	validators map[string]ValidatorFunc
}

// NewTypeValidator 创建类型验证器
func NewTypeValidator() *TypeValidator {
	return &TypeValidator{
		validators: make(map[string]ValidatorFunc),
	}
}

// CheckAuto 自动验证配置值是否符合标准
// standard: 标准类型或验证器
// val: 要验证的值
// fromKey: 键名（用于错误提示）
//
// 支持的标准类型:
//   - "int": 整数
//   - "str": 字符串
//   - "bool": 布尔值
//   - "float": 浮点数
//   - "list": 列表/数组
//   - "dict": 字典/对象
//   - "pint": 正整数（大于 0）
//   - "nnint": 非负整数（大于等于 0）
//
// 示例:
//   err := CheckAuto("int", 123, "max_count")
//   err := CheckAuto("pint", 10, "port")
//   err := CheckAuto([]string{"a", "b", "c"}, "b", "mode")
func CheckAuto(standard interface{}, val interface{}, fromKey string) error {
	if fromKey == "" {
		fromKey = "value"
	}

	// 如果 standard 是字符串，表示类型名称
	if typeStr, ok := standard.(string); ok {
		switch typeStr {
		case "int":
			if !isInt(val) {
				return fmt.Errorf("%s 必须是整数，当前值: %v", fromKey, val)
			}
		case "str":
			if _, ok := val.(string); !ok {
				return fmt.Errorf("%s 必须是字符串，当前值: %v", fromKey, val)
			}
		case "bool":
			if _, ok := val.(bool); !ok {
				return fmt.Errorf("%s 必须是布尔值，当前值: %v", fromKey, val)
			}
		case "float":
			if !isFloat(val) {
				return fmt.Errorf("%s 必须是浮点数，当前值: %v", fromKey, val)
			}
		case "list":
			if !isList(val) {
				return fmt.Errorf("%s 必须是列表，当前值: %v", fromKey, val)
			}
		case "dict":
			if !isDict(val) {
				return fmt.Errorf("%s 必须是字典，当前值: %v", fromKey, val)
			}
		case "pint":
			// 正整数（大于 0）
			if !isInt(val) {
				return fmt.Errorf("%s 必须是正整数，当前值: %v", fromKey, val)
			}
			intVal := toInt(val)
			if intVal <= 0 {
				return fmt.Errorf("%s 必须是正整数（大于 0），当前值: %d", fromKey, intVal)
			}
		case "nnint":
			// 非负整数（大于等于 0）
			if !isInt(val) {
				return fmt.Errorf("%s 必须是非负整数，当前值: %v", fromKey, val)
			}
			intVal := toInt(val)
			if intVal < 0 {
				return fmt.Errorf("%s 必须是非负整数（大于等于 0），当前值: %d", fromKey, intVal)
			}
		default:
			return fmt.Errorf("未知的验证类型: %s", typeStr)
		}
		return nil
	}

	// 如果 standard 是切片，表示枚举值
	standardVal := reflect.ValueOf(standard)
	if standardVal.Kind() == reflect.Slice {
		// 检查值是否在枚举列表中
		for i := 0; i < standardVal.Len(); i++ {
			if reflect.DeepEqual(standardVal.Index(i).Interface(), val) {
				return nil
			}
		}
		return fmt.Errorf("%s 的值必须是 %v 中的一个，当前值: %v", fromKey, standard, val)
	}

	// 如果 standard 是 map，表示嵌套配置验证
	if standardMap, ok := standard.(map[string]interface{}); ok {
		valMap, ok := val.(map[string]interface{})
		if !ok {
			return fmt.Errorf("%s 必须是字典类型", fromKey)
		}
		return ValidateConfig(valMap, standardMap)
	}

	return fmt.Errorf("不支持的验证标准: %T", standard)
}

// ValidateConfig 验证配置是否符合标准
// config: 要验证的配置
// standard: 标准配置模板
//
// 示例:
//   standard := map[string]interface{}{
//       "enable": "bool",
//       "port": "pint",
//       "mode": []string{"normal", "advanced"},
//   }
//   err := ValidateConfig(config, standard)
func ValidateConfig(config map[string]interface{}, standard map[string]interface{}) error {
	for key, standardVal := range standard {
		configVal, exists := config[key]
		if !exists {
			return fmt.Errorf("缺少配置项: %s", key)
		}

		if err := CheckAuto(standardVal, configVal, key); err != nil {
			return err
		}
	}
	return nil
}

// GetConfig 获取简单配置文件（不包含版本）
// fileName: 配置文件名
// defaultConfig: 默认配置
//
// 示例:
//   config, err := cfg.GetConfig("settings.json", map[string]interface{}{
//       "enable": true,
//   })
func (c *Config) GetConfig(fileName string, defaultConfig map[string]interface{}) (map[string]interface{}, error) {
	configPath := filepath.Join(c.configDir, fileName)

	// 确保配置目录存在
	if err := os.MkdirAll(c.configDir, 0755); err != nil {
		return nil, fmt.Errorf("创建配置目录失败: %v", err)
	}

	// 检查配置文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// 配置文件不存在，创建默认配置
		if err := c.SaveConfig(fileName, defaultConfig); err != nil {
			return nil, fmt.Errorf("创建默认配置文件失败: %v", err)
		}
		return defaultConfig, nil
	}

	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	// 解析配置文件
	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	return config, nil
}

// SaveConfig 保存简单配置文件（不包含版本）
func (c *Config) SaveConfig(fileName string, config map[string]interface{}) error {
	configPath := filepath.Join(c.configDir, fileName)

	// 确保配置目录存在
	if err := os.MkdirAll(c.configDir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %v", err)
	}

	// 序列化为 JSON
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}

	// 写入文件
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %v", err)
	}

	return nil
}

// GetConfigPath 获取配置文件的完整路径
func (c *Config) GetConfigPath(fileName string) string {
	return filepath.Join(c.configDir, fileName)
}

// ConfigExists 检查配置文件是否存在
func (c *Config) ConfigExists(fileName string) bool {
	configPath := filepath.Join(c.configDir, fileName)
	_, err := os.Stat(configPath)
	return err == nil
}

// DeleteConfig 删除配置文件
func (c *Config) DeleteConfig(fileName string) error {
	configPath := filepath.Join(c.configDir, fileName)
	if err := os.Remove(configPath); err != nil {
		return fmt.Errorf("删除配置文件失败: %v", err)
	}
	return nil
}

// 辅助函数

func isInt(val interface{}) bool {
	switch val.(type) {
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64:
		return true
	case float64:
		// JSON 解析时数字默认为 float64
		f := val.(float64)
		return f == float64(int64(f))
	default:
		return false
	}
}

func toInt(val interface{}) int64 {
	switch v := val.(type) {
	case int:
		return int64(v)
	case int8:
		return int64(v)
	case int16:
		return int64(v)
	case int32:
		return int64(v)
	case int64:
		return v
	case uint:
		return int64(v)
	case uint8:
		return int64(v)
	case uint16:
		return int64(v)
	case uint32:
		return int64(v)
	case uint64:
		return int64(v)
	case float64:
		return int64(v)
	default:
		return 0
	}
}

func isFloat(val interface{}) bool {
	switch val.(type) {
	case float32, float64:
		return true
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64:
		return true
	default:
		return false
	}
}

func isList(val interface{}) bool {
	v := reflect.ValueOf(val)
	return v.Kind() == reflect.Slice || v.Kind() == reflect.Array
}

func isDict(val interface{}) bool {
	v := reflect.ValueOf(val)
	return v.Kind() == reflect.Map
}
