package sdk

import (
	"fmt"
	"sync"
	"time"
)

// PacketWaiter 数据包等待器，用于等待特定数据包
type PacketWaiter struct {
	mu       sync.RWMutex
	waiters  map[uint32][]chan interface{}
	allWaiters []chan PacketEvent
}

// NewPacketWaiter 创建数据包等待器
func NewPacketWaiter() *PacketWaiter {
	return &PacketWaiter{
		waiters:    make(map[uint32][]chan interface{}),
		allWaiters: make([]chan PacketEvent, 0),
	}
}

// WaitNextPacket 等待下一个指定类型的数据包
// packetID: 数据包 ID
// timeout: 超时时间（秒）
// 返回: 数据包内容，超时返回 nil
//
// 示例:
//   packet, err := waiter.WaitNextPacket(9, 30.0)
//   if err != nil {
//       // 超时或错误处理
//   }
func (pw *PacketWaiter) WaitNextPacket(packetID uint32, timeout float64) (interface{}, error) {
	ch := make(chan interface{}, 1)

	pw.mu.Lock()
	pw.waiters[packetID] = append(pw.waiters[packetID], ch)
	pw.mu.Unlock()

	timer := time.NewTimer(time.Duration(timeout * float64(time.Second)))
	defer timer.Stop()

	select {
	case packet := <-ch:
		return packet, nil
	case <-timer.C:
		// 超时，清理等待器
		pw.mu.Lock()
		if waiters, exists := pw.waiters[packetID]; exists {
			for i, waiter := range waiters {
				if waiter == ch {
					pw.waiters[packetID] = append(waiters[:i], waiters[i+1:]...)
					break
				}
			}
		}
		pw.mu.Unlock()
		return nil, fmt.Errorf("等待数据包超时")
	}
}

// WaitNextPacketAny 等待任意数据包
// timeout: 超时时间（秒）
// 返回: PacketEvent，超时返回错误
//
// 警告: 此方法性能开销较大，不建议频繁使用
//
// 示例:
//   event, err := waiter.WaitNextPacketAny(10.0)
//   if err == nil {
//       ctx.Logf("收到数据包 ID: %d", event.ID)
//   }
func (pw *PacketWaiter) WaitNextPacketAny(timeout float64) (PacketEvent, error) {
	ch := make(chan PacketEvent, 1)

	pw.mu.Lock()
	pw.allWaiters = append(pw.allWaiters, ch)
	pw.mu.Unlock()

	timer := time.NewTimer(time.Duration(timeout * float64(time.Second)))
	defer timer.Stop()

	select {
	case packet := <-ch:
		return packet, nil
	case <-timer.C:
		// 超时，清理等待器
		pw.mu.Lock()
		for i, waiter := range pw.allWaiters {
			if waiter == ch {
				pw.allWaiters = append(pw.allWaiters[:i], pw.allWaiters[i+1:]...)
				break
			}
		}
		pw.mu.Unlock()
		return PacketEvent{}, fmt.Errorf("等待数据包超时")
	}
}

// NotifyPacket 通知数据包到达（内部方法，由主程序调用）
func (pw *PacketWaiter) NotifyPacket(packetID uint32, packet interface{}) {
	pw.mu.Lock()
	defer pw.mu.Unlock()

	// 通知特定数据包的等待器
	if waiters, exists := pw.waiters[packetID]; exists && len(waiters) > 0 {
		// 只通知第一个等待器
		select {
		case waiters[0] <- packet:
			pw.waiters[packetID] = waiters[1:]
		default:
		}
	}

	// 通知所有数据包的等待器
	event := PacketEvent{
		ID:  packetID,
		Raw: packet,
	}
	for i := len(pw.allWaiters) - 1; i >= 0; i-- {
		select {
		case pw.allWaiters[i] <- event:
			// 成功发送后移除该等待器
			pw.allWaiters = append(pw.allWaiters[:i], pw.allWaiters[i+1:]...)
		default:
		}
	}
}

// Clear 清理所有等待器（热重载时调用）
func (pw *PacketWaiter) Clear() {
	pw.mu.Lock()
	defer pw.mu.Unlock()

	pw.waiters = make(map[uint32][]chan interface{})
	pw.allWaiters = make([]chan PacketEvent, 0)
}
