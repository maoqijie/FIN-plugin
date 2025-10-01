package sdk

import (
	"fmt"
	"sync"
)

// PluginAPIVersion 插件 API 版本
type PluginAPIVersion struct {
	Major int // 主版本号
	Minor int // 次版本号
	Patch int // 修订号
}

// String 返回版本字符串
func (v PluginAPIVersion) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// Compare 比较两个版本
// 返回: 1 表示 v > other, 0 表示相等, -1 表示 v < other
func (v PluginAPIVersion) Compare(other PluginAPIVersion) int {
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

// IsCompatible 检查版本是否兼容（主版本号相同）
func (v PluginAPIVersion) IsCompatible(other PluginAPIVersion) bool {
	return v.Major == other.Major
}

// PluginAPIInfo 插件 API 信息
type PluginAPIInfo struct {
	Name    string             // API 名称
	Version PluginAPIVersion   // API 版本
	Plugin  Plugin             // 插件实例
}

// PluginAPIRegistry 插件 API 注册表
type PluginAPIRegistry struct {
	mu       sync.RWMutex
	registry map[string]*PluginAPIInfo // key: API 名称
}

// NewPluginAPIRegistry 创建插件 API 注册表
func NewPluginAPIRegistry() *PluginAPIRegistry {
	return &PluginAPIRegistry{
		registry: make(map[string]*PluginAPIInfo),
	}
}

// Register 注册插件 API
func (r *PluginAPIRegistry) Register(name string, version PluginAPIVersion, plugin Plugin) error {
	if name == "" {
		return fmt.Errorf("API 名称不能为空")
	}
	if plugin == nil {
		return fmt.Errorf("插件实例不能为空")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.registry[name]; exists {
		return fmt.Errorf("API '%s' 已被注册", name)
	}

	r.registry[name] = &PluginAPIInfo{
		Name:    name,
		Version: version,
		Plugin:  plugin,
	}

	return nil
}

// Unregister 注销插件 API
func (r *PluginAPIRegistry) Unregister(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.registry, name)
}

// Get 获取插件 API
func (r *PluginAPIRegistry) Get(name string) (Plugin, PluginAPIVersion, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	info, exists := r.registry[name]
	if !exists {
		return nil, PluginAPIVersion{}, fmt.Errorf("API '%s' 未注册", name)
	}

	return info.Plugin, info.Version, nil
}

// GetWithVersion 获取指定版本的插件 API（检查兼容性）
func (r *PluginAPIRegistry) GetWithVersion(name string, requiredVersion PluginAPIVersion) (Plugin, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	info, exists := r.registry[name]
	if !exists {
		return nil, fmt.Errorf("API '%s' 未注册", name)
	}

	// 检查主版本号兼容性
	if !info.Version.IsCompatible(requiredVersion) {
		return nil, fmt.Errorf("API '%s' 版本不兼容：需要 %s，实际 %s",
			name, requiredVersion.String(), info.Version.String())
	}

	// 检查次版本号（向后兼容）
	if info.Version.Compare(requiredVersion) < 0 {
		return nil, fmt.Errorf("API '%s' 版本过低：需要 %s，实际 %s",
			name, requiredVersion.String(), info.Version.String())
	}

	return info.Plugin, nil
}

// List 列出所有已注册的 API
func (r *PluginAPIRegistry) List() []PluginAPIInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list := make([]PluginAPIInfo, 0, len(r.registry))
	for _, info := range r.registry {
		list = append(list, *info)
	}

	return list
}

// Clear 清空所有注册的 API
func (r *PluginAPIRegistry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.registry = make(map[string]*PluginAPIInfo)
}

// Has 检查 API 是否已注册
func (r *PluginAPIRegistry) Has(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, exists := r.registry[name]
	return exists
}
