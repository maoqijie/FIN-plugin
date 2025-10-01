package sdk

import (
	"strconv"
	"strings"
	"sync"
	"time"
)

// Utils 提供实用工具方法，类似 ToolDelta 的 utils 模块
type Utils struct{}

// NewUtils 创建 Utils 实例
func NewUtils() *Utils {
	return &Utils{}
}

// SimpleFormat 执行简单的字符串格式化替换
// kw: 替换字典，键为占位符（不含 {}），值为替换内容
// sub: 要格式化的字符串，使用 {key} 作为占位符
// 返回: 格式化后的字符串
//
// 示例:
//   kw := map[string]string{"name": "玩家1", "score": "100"}
//   result := utils.SimpleFormat(kw, "玩家 {name} 的分数是 {score}")
//   // 返回: "玩家 玩家1 的分数是 100"
func (u *Utils) SimpleFormat(kw map[string]string, sub string) string {
	result := sub
	for key, value := range kw {
		placeholder := "{" + key + "}"
		result = strings.ReplaceAll(result, placeholder, value)
	}
	return result
}

// TryInt 尝试将输入转换为整数
// input: 任意类型的输入
// 返回: 转换成功返回整数和 true，失败返回 0 和 false
//
// 支持的类型:
//   - int, int8, int16, int32, int64
//   - uint, uint8, uint16, uint32, uint64
//   - float32, float64
//   - string (可解析的数字字符串)
//   - bool (true=1, false=0)
func (u *Utils) TryInt(input interface{}) (int, bool) {
	switch v := input.(type) {
	case int:
		return v, true
	case int8:
		return int(v), true
	case int16:
		return int(v), true
	case int32:
		return int(v), true
	case int64:
		return int(v), true
	case uint:
		return int(v), true
	case uint8:
		return int(v), true
	case uint16:
		return int(v), true
	case uint32:
		return int(v), true
	case uint64:
		return int(v), true
	case float32:
		return int(v), true
	case float64:
		return int(v), true
	case string:
		if i, err := strconv.Atoi(v); err == nil {
			return i, true
		}
		return 0, false
	case bool:
		if v {
			return 1, true
		}
		return 0, true
	default:
		return 0, false
	}
}

// FillListIndex 用默认值填充列表，使其长度与参考列表一致
// list: 要填充的切片
// reference: 参考切片（用于确定目标长度）
// defaultValue: 用于填充的默认值
// 返回: 填充后的切片
//
// 示例:
//   list := []string{"a", "b"}
//   reference := []string{"1", "2", "3", "4"}
//   result := utils.FillListIndex(list, reference, "default")
//   // 返回: ["a", "b", "default", "default"]
func (u *Utils) FillListIndex(list []interface{}, reference []interface{}, defaultValue interface{}) []interface{} {
	result := make([]interface{}, len(list))
	copy(result, list)

	for len(result) < len(reference) {
		result = append(result, defaultValue)
	}

	return result
}

// FillStringList 用默认值填充字符串列表（类型安全的版本）
func (u *Utils) FillStringList(list []string, referenceLen int, defaultValue string) []string {
	result := make([]string, len(list))
	copy(result, list)

	for len(result) < referenceLen {
		result = append(result, defaultValue)
	}

	return result
}

// FillIntList 用默认值填充整数列表（类型安全的版本）
func (u *Utils) FillIntList(list []int, referenceLen int, defaultValue int) []int {
	result := make([]int, len(list))
	copy(result, list)

	for len(result) < referenceLen {
		result = append(result, defaultValue)
	}

	return result
}

// ToPlayerSelector 将玩家名转换为目标选择器
// playerName: 玩家名称
// 返回: 目标选择器字符串
//
// 示例:
//   selector := utils.ToPlayerSelector("玩家1")
//   // 返回: "@a[name=\"玩家1\"]"
func (u *Utils) ToPlayerSelector(playerName string) string {
	if strings.HasPrefix(playerName, "@") {
		// 已经是选择器，直接返回
		return playerName
	}
	return `@a[name="` + playerName + `"]`
}

// ResultCallback 表示一对回调锁（getter 和 setter）
type ResultCallback struct {
	mu      sync.Mutex
	value   interface{}
	ready   bool
	cond    *sync.Cond
	timeout time.Duration
}

// CreateResultCallback 创建一对回调锁（getter 和 setter 方法）
// timeout: 超时时间（秒），0 表示永不超时
// 返回: getter 函数和 setter 函数
//
// 用途: 在异步操作中等待结果
//
// 示例:
//   getter, setter := utils.CreateResultCallback(5.0)
//
//   // 在协程中等待结果
//   go func() {
//       result, ok := getter()
//       if ok {
//           fmt.Println("收到结果:", result)
//       } else {
//           fmt.Println("超时或未设置")
//       }
//   }()
//
//   // 在另一个地方设置结果
//   time.Sleep(2 * time.Second)
//   setter("操作完成")
func (u *Utils) CreateResultCallback(timeout float64) (func() (interface{}, bool), func(interface{})) {
	cb := &ResultCallback{
		ready:   false,
		timeout: time.Duration(timeout * float64(time.Second)),
	}
	cb.cond = sync.NewCond(&cb.mu)

	// getter 函数：等待并获取结果
	getter := func() (interface{}, bool) {
		cb.mu.Lock()
		defer cb.mu.Unlock()

		if cb.timeout > 0 {
			// 带超时的等待
			done := make(chan struct{})
			go func() {
				cb.mu.Lock()
				for !cb.ready {
					cb.cond.Wait()
				}
				cb.mu.Unlock()
				close(done)
			}()

			select {
			case <-done:
				return cb.value, true
			case <-time.After(cb.timeout):
				return nil, false
			}
		} else {
			// 无超时等待
			for !cb.ready {
				cb.cond.Wait()
			}
			return cb.value, true
		}
	}

	// setter 函数：设置结果并唤醒等待者
	setter := func(value interface{}) {
		cb.mu.Lock()
		defer cb.mu.Unlock()

		if cb.ready {
			// 已经设置过，忽略后续设置
			return
		}

		cb.value = value
		cb.ready = true
		cb.cond.Broadcast()
	}

	return getter, setter
}

// RunAsync 在新的 goroutine 中运行函数（简化的线程工具）
// fn: 要执行的函数
//
// 示例:
//   utils.RunAsync(func() {
//       time.Sleep(3 * time.Second)
//       fmt.Println("异步任务完成")
//   })
func (u *Utils) RunAsync(fn func()) {
	go fn()
}

// RunAsyncWithResult 在新的 goroutine 中运行函数并通过 channel 返回结果
// fn: 要执行的函数，返回任意类型的结果
// 返回: 用于接收结果的 channel
//
// 示例:
//   resultChan := utils.RunAsyncWithResult(func() interface{} {
//       time.Sleep(2 * time.Second)
//       return "任务完成"
//   })
//   result := <-resultChan
//   fmt.Println(result)
func (u *Utils) RunAsyncWithResult(fn func() interface{}) <-chan interface{} {
	resultChan := make(chan interface{}, 1)
	go func() {
		result := fn()
		resultChan <- result
		close(resultChan)
	}()
	return resultChan
}

// Gather 并行运行多个函数并收集结果
// fns: 要并行执行的函数列表
// 返回: 所有函数的返回值列表（顺序与输入一致）
//
// 示例:
//   results := utils.Gather(
//       func() interface{} { return "结果1" },
//       func() interface{} { return "结果2" },
//       func() interface{} { return "结果3" },
//   )
//   // results = []interface{}{"结果1", "结果2", "结果3"}
func (u *Utils) Gather(fns ...func() interface{}) []interface{} {
	results := make([]interface{}, len(fns))
	var wg sync.WaitGroup

	for i, fn := range fns {
		wg.Add(1)
		go func(index int, f func() interface{}) {
			defer wg.Done()
			results[index] = f()
		}(i, fn)
	}

	wg.Wait()
	return results
}

// Timer 定时器结构，用于创建定时任务
type Timer struct {
	interval time.Duration
	fn       func()
	stopChan chan struct{}
	running  bool
	mu       sync.Mutex
}

// NewTimer 创建新的定时器
// interval: 执行间隔（秒）
// fn: 要定时执行的函数
// 返回: Timer 实例
//
// 示例:
//   timer := utils.NewTimer(5.0, func() {
//       fmt.Println("每 5 秒执行一次")
//   })
//   timer.Start()
//   // 稍后停止
//   timer.Stop()
func (u *Utils) NewTimer(interval float64, fn func()) *Timer {
	return &Timer{
		interval: time.Duration(interval * float64(time.Second)),
		fn:       fn,
		stopChan: make(chan struct{}),
		running:  false,
	}
}

// Start 启动定时器
func (t *Timer) Start() {
	t.mu.Lock()
	if t.running {
		t.mu.Unlock()
		return
	}
	t.running = true
	t.mu.Unlock()

	go func() {
		ticker := time.NewTicker(t.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				t.fn()
			case <-t.stopChan:
				return
			}
		}
	}()
}

// Stop 停止定时器
func (t *Timer) Stop() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.running {
		return
	}

	close(t.stopChan)
	t.running = false
}

// IsRunning 检查定时器是否正在运行
func (t *Timer) IsRunning() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.running
}

// Sleep 睡眠指定秒数（便捷方法）
func (u *Utils) Sleep(seconds float64) {
	time.Sleep(time.Duration(seconds * float64(time.Second)))
}

// Contains 检查切片是否包含指定元素（字符串版本）
func (u *Utils) Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// ContainsInt 检查切片是否包含指定元素（整数版本）
func (u *Utils) ContainsInt(slice []int, item int) bool {
	for _, n := range slice {
		if n == item {
			return true
		}
	}
	return false
}

// Max 返回两个整数中的较大值
func (u *Utils) Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Min 返回两个整数中的较小值
func (u *Utils) Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Clamp 将值限制在指定范围内
func (u *Utils) Clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
