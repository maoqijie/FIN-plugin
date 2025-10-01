package sdk

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// TempJSON 提供缓存式 JSON 文件管理，类似 ToolDelta 的 tempjson
// 通过内存缓存减少磁盘 I/O，提升性能
type TempJSON struct {
	mu           sync.RWMutex
	cache        map[string]*jsonCache
	defaultDir   string
	globalMu     sync.Mutex // 全局锁，用于定时器清理
}

// jsonCache JSON 文件缓存条目
type jsonCache struct {
	data      interface{}   // 缓存的 JSON 数据
	timer     *time.Timer   // 自动卸载定时器
	mu        sync.RWMutex  // 单个缓存项的读写锁
	modified  bool          // 是否被修改过
}

// NewTempJSON 创建 TempJSON 实例
// defaultDir: 默认 JSON 文件目录（可选）
func NewTempJSON(defaultDir ...string) *TempJSON {
	dir := "."
	if len(defaultDir) > 0 && defaultDir[0] != "" {
		dir = defaultDir[0]
	}
	return &TempJSON{
		cache:      make(map[string]*jsonCache),
		defaultDir: dir,
	}
}

// Load 加载 JSON 文件到内存缓存
// path: JSON 文件路径
// needFileExists: 是否要求文件必须存在
// defaultContent: 文件不存在时使用的默认内容
// timeout: 自动卸载超时时间（秒），0 表示不自动卸载
//
// 示例:
//   tj := sdk.NewTempJSON()
//   err := tj.Load("data.json", false, map[string]interface{}{"count": 0}, 30.0)
func (t *TempJSON) Load(path string, needFileExists bool, defaultContent interface{}, timeout float64) error {
	// 转换为绝对路径
	absPath := t.getAbsolutePath(path)

	// 检查是否已经在缓存中
	t.mu.RLock()
	if entry, exists := t.cache[absPath]; exists {
		t.mu.RUnlock()
		// 重置定时器
		if timeout > 0 {
			entry.mu.Lock()
			if entry.timer != nil {
				entry.timer.Stop()
			}
			entry.timer = time.AfterFunc(time.Duration(timeout*float64(time.Second)), func() {
				t.Unload(path)
			})
			entry.mu.Unlock()
		}
		return nil
	}
	t.mu.RUnlock()

	// 检查文件是否存在
	_, err := os.Stat(absPath)
	fileExists := err == nil

	if needFileExists && !fileExists {
		return fmt.Errorf("文件不存在: %s", absPath)
	}

	var data interface{}

	if fileExists {
		// 读取文件内容
		fileData, err := os.ReadFile(absPath)
		if err != nil {
			return fmt.Errorf("读取文件失败: %v", err)
		}

		// 解析 JSON
		if err := json.Unmarshal(fileData, &data); err != nil {
			return fmt.Errorf("解析 JSON 失败: %v", err)
		}
	} else {
		// 使用默认内容
		if defaultContent != nil {
			data = defaultContent
			// 创建目录并保存默认内容
			if err := t.saveToFile(absPath, data); err != nil {
				return fmt.Errorf("保存默认内容失败: %v", err)
			}
		} else {
			data = make(map[string]interface{})
		}
	}

	// 创建缓存条目
	entry := &jsonCache{
		data:     data,
		modified: false,
	}

	// 设置自动卸载定时器
	if timeout > 0 {
		entry.timer = time.AfterFunc(time.Duration(timeout*float64(time.Second)), func() {
			t.Unload(path)
		})
	}

	// 添加到缓存
	t.mu.Lock()
	t.cache[absPath] = entry
	t.mu.Unlock()

	return nil
}

// Unload 卸载 JSON 文件缓存（如果有修改则保存到磁盘）
// path: JSON 文件路径
//
// 示例:
//   tj.Unload("data.json")
func (t *TempJSON) Unload(path string) error {
	absPath := t.getAbsolutePath(path)

	t.mu.Lock()
	entry, exists := t.cache[absPath]
	if !exists {
		t.mu.Unlock()
		return nil // 不在缓存中，直接返回
	}
	delete(t.cache, absPath)
	t.mu.Unlock()

	// 停止定时器
	entry.mu.Lock()
	if entry.timer != nil {
		entry.timer.Stop()
	}

	// 如果有修改，保存到磁盘
	modified := entry.modified
	data := entry.data
	entry.mu.Unlock()

	if modified {
		if err := t.saveToFile(absPath, data); err != nil {
			return fmt.Errorf("保存文件失败: %v", err)
		}
	}

	return nil
}

// Read 从缓存中读取 JSON 数据
// path: JSON 文件路径
// deepCopy: 是否深拷贝（默认 true，防止外部修改缓存）
//
// 返回: JSON 数据、错误
//
// 示例:
//   data, err := tj.Read("data.json", true)
//   if err != nil {
//       // 处理错误
//   }
//   dataMap := data.(map[string]interface{})
func (t *TempJSON) Read(path string, deepCopy bool) (interface{}, error) {
	absPath := t.getAbsolutePath(path)

	t.mu.RLock()
	entry, exists := t.cache[absPath]
	t.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("文件未加载到缓存: %s", path)
	}

	entry.mu.RLock()
	defer entry.mu.RUnlock()

	if deepCopy {
		// 深拷贝：通过 JSON 序列化/反序列化
		return deepCopyJSON(entry.data)
	}

	return entry.data, nil
}

// Write 写入 JSON 数据到缓存
// path: JSON 文件路径
// obj: 要写入的数据
//
// 示例:
//   tj.Write("data.json", map[string]interface{}{
//       "count": 10,
//       "name": "test",
//   })
func (t *TempJSON) Write(path string, obj interface{}) error {
	absPath := t.getAbsolutePath(path)

	t.mu.RLock()
	entry, exists := t.cache[absPath]
	t.mu.RUnlock()

	if !exists {
		return fmt.Errorf("文件未加载到缓存: %s", path)
	}

	entry.mu.Lock()
	entry.data = obj
	entry.modified = true
	entry.mu.Unlock()

	return nil
}

// LoadAndRead 加载并读取 JSON 文件（快捷方法）
// 读取后自动卸载（除非指定 timeout > 0）
//
// path: JSON 文件路径
// needFileExists: 是否要求文件必须存在
// defaultContent: 文件不存在时使用的默认内容
// timeout: 自动卸载超时时间（秒），0 表示立即卸载
//
// 返回: JSON 数据、错误
//
// 示例:
//   // 快速读取，立即卸载
//   data, err := tj.LoadAndRead("data.json", false, map[string]interface{}{"age": 0}, 0)
//   if err != nil {
//       return err
//   }
//   dataMap := data.(map[string]interface{})
//   age := int(dataMap["age"].(float64))
func (t *TempJSON) LoadAndRead(path string, needFileExists bool, defaultContent interface{}, timeout float64) (interface{}, error) {
	// 加载到缓存
	if err := t.Load(path, needFileExists, defaultContent, timeout); err != nil {
		return nil, err
	}

	// 读取数据
	data, err := t.Read(path, true)
	if err != nil {
		return nil, err
	}

	// 如果 timeout 为 0，立即卸载
	if timeout == 0 {
		if err := t.Unload(path); err != nil {
			return data, fmt.Errorf("卸载失败: %v", err)
		}
	}

	return data, nil
}

// LoadAndWrite 加载并写入 JSON 文件（快捷方法）
// 写入后自动卸载（除非指定 timeout > 0）
//
// path: JSON 文件路径
// obj: 要写入的数据
// needFileExists: 是否要求文件必须存在
// timeout: 自动卸载超时时间（秒），0 表示立即卸载
//
// 示例:
//   // 快速写入，立即保存到磁盘
//   err := tj.LoadAndWrite("data.json", map[string]interface{}{
//       "age": 18,
//       "name": "test",
//   }, false, 0)
func (t *TempJSON) LoadAndWrite(path string, obj interface{}, needFileExists bool, timeout float64) error {
	// 加载到缓存（如果文件不存在会创建）
	if err := t.Load(path, needFileExists, nil, timeout); err != nil {
		return err
	}

	// 写入数据
	if err := t.Write(path, obj); err != nil {
		return err
	}

	// 如果 timeout 为 0，立即卸载（保存到磁盘）
	if timeout == 0 {
		if err := t.Unload(path); err != nil {
			return fmt.Errorf("卸载失败: %v", err)
		}
	}

	return nil
}

// UnloadAll 卸载所有缓存的 JSON 文件
func (t *TempJSON) UnloadAll() error {
	t.mu.Lock()
	paths := make([]string, 0, len(t.cache))
	for path := range t.cache {
		paths = append(paths, path)
	}
	t.mu.Unlock()

	var lastErr error
	for _, path := range paths {
		if err := t.Unload(path); err != nil {
			lastErr = err
		}
	}

	return lastErr
}

// GetCachedPaths 获取所有已缓存的文件路径
func (t *TempJSON) GetCachedPaths() []string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	paths := make([]string, 0, len(t.cache))
	for path := range t.cache {
		paths = append(paths, path)
	}
	return paths
}

// IsCached 检查文件是否在缓存中
func (t *TempJSON) IsCached(path string) bool {
	absPath := t.getAbsolutePath(path)
	t.mu.RLock()
	defer t.mu.RUnlock()
	_, exists := t.cache[absPath]
	return exists
}

// SaveAll 保存所有缓存的 JSON 文件到磁盘（但不卸载）
func (t *TempJSON) SaveAll() error {
	t.mu.RLock()
	entries := make(map[string]*jsonCache, len(t.cache))
	for path, entry := range t.cache {
		entries[path] = entry
	}
	t.mu.RUnlock()

	var lastErr error
	for path, entry := range entries {
		entry.mu.RLock()
		if entry.modified {
			data := entry.data
			entry.mu.RUnlock()

			if err := t.saveToFile(path, data); err != nil {
				lastErr = err
			} else {
				entry.mu.Lock()
				entry.modified = false
				entry.mu.Unlock()
			}
		} else {
			entry.mu.RUnlock()
		}
	}

	return lastErr
}

// 辅助方法

// getAbsolutePath 获取绝对路径
func (t *TempJSON) getAbsolutePath(path string) string {
	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	}
	return filepath.Clean(filepath.Join(t.defaultDir, path))
}

// saveToFile 保存数据到文件
func (t *TempJSON) saveToFile(path string, data interface{}) error {
	// 确保目录存在
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %v", err)
	}

	// 序列化为 JSON
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化 JSON 失败: %v", err)
	}

	// 写入文件
	if err := os.WriteFile(path, jsonData, 0644); err != nil {
		return fmt.Errorf("写入文件失败: %v", err)
	}

	return nil
}

// deepCopyJSON 通过 JSON 序列化/反序列化进行深拷贝
func deepCopyJSON(src interface{}) (interface{}, error) {
	// 序列化
	data, err := json.Marshal(src)
	if err != nil {
		return nil, fmt.Errorf("序列化失败: %v", err)
	}

	// 反序列化
	var dst interface{}
	if err := json.Unmarshal(data, &dst); err != nil {
		return nil, fmt.Errorf("反序列化失败: %v", err)
	}

	return dst, nil
}
