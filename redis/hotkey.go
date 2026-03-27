package redis

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// HotKeyDetector 热键检测器
type HotKeyDetector struct {
	mu            sync.RWMutex
	counters      map[string]int64
	threshold     int64
	hotKeys       map[string]time.Time
	reportChannel chan string
	stopChannel  chan struct{}
	isRunning    int32
}

// NewHotKeyDetector 创建新的热键检测器
func NewHotKeyDetector(threshold int64) *HotKeyDetector {
	return &HotKeyDetector{
		counters:      make(map[string]int64),
		threshold:     threshold,
		hotKeys:       make(map[string]time.Time),
		reportChannel: make(chan string, 100),
		stopChannel:  make(chan struct{}),
	}
}

// RecordAccess 记录键访问
func (h *HotKeyDetector) RecordAccess(key string) {
	h.mu.Lock()
	h.counters[key]++
	count := h.counters[key]
	
	if count >= h.threshold {
		if _, exists := h.hotKeys[key]; !exists {
			h.hotKeys[key] = time.Now()
			select {
			case h.reportChannel <- key:
			default:
			}
		}
	}
	h.mu.Unlock()
}

// Start 启动热键检测
func (h *HotKeyDetector) Start(ctx context.Context) {
	if !atomic.CompareAndSwapInt32(&h.isRunning, 0, 1) {
		return // 已经在运行
	}

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			h.Stop()
			return
		case <-h.stopChannel:
			return
		case <-ticker.C:
			h.resetCounters()
		case key := <-h.reportChannel:
			// 可以在这里上报热键到监控系统
			h.mu.RLock()
			count := h.counters[key]
			h.mu.RUnlock()
			fmt.Printf("Hot key detected: %s (access count: %d)\n", key, count)
		}
	}
}

// Stop 停止热键检测
func (h *HotKeyDetector) Stop() {
	if !atomic.CompareAndSwapInt32(&h.isRunning, 1, 0) {
		return // 已经停止
	}

	close(h.stopChannel)
}

// resetCounters 重置计数器
func (h *HotKeyDetector) resetCounters() {
	h.mu.Lock()
	defer h.mu.Unlock()

	for key, count := range h.counters {
		if count < h.threshold/2 {
			delete(h.counters, key)
		} else {
			h.counters[key] = 0
		}
	}

	// 清理过期的热键记录
	now := time.Now()
	for key, detectedAt := range h.hotKeys {
		if now.Sub(detectedAt) > time.Minute {
			delete(h.hotKeys, key)
		}
	}
}

// GetHotKeys 获取当前热键列表
func (h *HotKeyDetector) GetHotKeys() map[string]time.Time {
	h.mu.RLock()
	defer h.mu.RUnlock()

	result := make(map[string]time.Time)
	for k, v := range h.hotKeys {
		result[k] = v
	}
	return result
}

// GetCounters 获取当前计数器
func (h *HotKeyDetector) GetCounters() map[string]int64 {
	h.mu.RLock()
	defer h.mu.RUnlock()

	result := make(map[string]int64)
	for k, v := range h.counters {
		result[k] = v
	}
	return result
}

// IsRunning 检查是否正在运行
func (h *HotKeyDetector) IsRunning() bool {
	return atomic.LoadInt32(&h.isRunning) == 1
}
