package redis

import (
	"sync"
	"sync/atomic"
	"time"
)

// Metrics 缓存指标统计
type Metrics struct {
	mu sync.RWMutex

	// 基础统计
	hits   int64 // 缓存命中数
	misses int64 // 缓存未命中数
	sets   int64 // 设置操作数
	deletes int64 // 删除操作数
	errors int64 // 错误数

	// 加载统计
	loads        int64 // 加载操作数
	loadErrors   int64 // 加载错误数
	loadFailures int64 // 加载失败数

	// 延迟统计
	getLatencies    []time.Duration // Get 操作延迟
	setLatencies    []time.Duration // Set 操作延迟
	deleteLatencies []time.Duration // Delete 操作延迟

	// 锁统计
	lockAcquires   int64 // 锁获取次数
	lockConflicts  int64 // 锁冲突次数
	lockTimeouts   int64 // 锁超时次数
	lockErrors     int64 // 锁错误次数

	// 热键统计
	hotKeys map[string]int64 // 热键访问计数

	// 配置
	maxLatencyRecords int // 最大延迟记录数
}

// NewMetrics 创建新的指标实例
func NewMetrics() *Metrics {
	return &Metrics{
		getLatencies:     make([]time.Duration, 0),
		setLatencies:     make([]time.Duration, 0),
		deleteLatencies:  make([]time.Duration, 0),
		hotKeys:         make(map[string]int64),
		maxLatencyRecords: 1000,
	}
}

// RecordHit 记录缓存命中
func (m *Metrics) RecordHit() {
	atomic.AddInt64(&m.hits, 1)
}

// RecordMiss 记录缓存未命中
func (m *Metrics) RecordMiss() {
	atomic.AddInt64(&m.misses, 1)
}

// RecordSet 记录设置操作
func (m *Metrics) RecordSet() {
	atomic.AddInt64(&m.sets, 1)
}

// RecordDelete 记录删除操作
func (m *Metrics) RecordDelete() {
	atomic.AddInt64(&m.deletes, 1)
}

// RecordError 记录错误
func (m *Metrics) RecordError() {
	atomic.AddInt64(&m.errors, 1)
}

// RecordLoad 记录加载操作
func (m *Metrics) RecordLoad() {
	atomic.AddInt64(&m.loads, 1)
}

// RecordLoadError 记录加载错误
func (m *Metrics) RecordLoadError() {
	atomic.AddInt64(&m.loadErrors, 1)
}

// RecordLoadFailure 记录加载失败
func (m *Metrics) RecordLoadFailure() {
	atomic.AddInt64(&m.loadFailures, 1)
}

// RecordLatency 记录操作延迟
func (m *Metrics) RecordLatency(operation string, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	switch operation {
	case "get":
		m.addLatency(&m.getLatencies, duration)
	case "set":
		m.addLatency(&m.setLatencies, duration)
	case "delete":
		m.addLatency(&m.deleteLatencies, duration)
	}
}

// addLatency 添加延迟记录
func (m *Metrics) addLatency(latencies *[]time.Duration, duration time.Duration) {
	*latencies = append(*latencies, duration)
	if len(*latencies) > m.maxLatencyRecords {
		*latencies = (*latencies)[1:]
	}
}

// RecordLockAcquire 记录锁获取
func (m *Metrics) RecordLockAcquire() {
	atomic.AddInt64(&m.lockAcquires, 1)
}

// RecordLockConflict 记录锁冲突
func (m *Metrics) RecordLockConflict() {
	atomic.AddInt64(&m.lockConflicts, 1)
}

// RecordLockTimeout 记录锁超时
func (m *Metrics) RecordLockTimeout() {
	atomic.AddInt64(&m.lockTimeouts, 1)
}

// RecordLockError 记录锁错误
func (m *Metrics) RecordLockError() {
	atomic.AddInt64(&m.lockErrors, 1)
}

// RecordHotKey 记录热键访问
func (m *Metrics) RecordHotKey(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.hotKeys[key]++
}

// GetStats 获取统计信息
func (m *Metrics) GetStats() *Stats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	hits := atomic.LoadInt64(&m.hits)
	misses := atomic.LoadInt64(&m.misses)
	sets := atomic.LoadInt64(&m.sets)
	deletes := atomic.LoadInt64(&m.deletes)
	errors := atomic.LoadInt64(&m.errors)
	loads := atomic.LoadInt64(&m.loads)
	loadErrors := atomic.LoadInt64(&m.loadErrors)
	loadFailures := atomic.LoadInt64(&m.loadFailures)
	lockAcquires := atomic.LoadInt64(&m.lockAcquires)
	lockConflicts := atomic.LoadInt64(&m.lockConflicts)
	lockTimeouts := atomic.LoadInt64(&m.lockTimeouts)
	lockErrors := atomic.LoadInt64(&m.lockErrors)

	totalOperations := hits + misses
	hitRate := float64(0)
	if totalOperations > 0 {
		hitRate = float64(hits) / float64(totalOperations) * 100
	}

	return &Stats{
		Hits:            hits,
		Misses:          misses,
		Sets:            sets,
		Deletes:         deletes,
		Errors:          errors,
		Loads:           loads,
		LoadErrors:      loadErrors,
		LoadFailures:    loadFailures,
		LockAcquires:    lockAcquires,
		LockConflicts:   lockConflicts,
		LockTimeouts:    lockTimeouts,
		LockErrors:      lockErrors,
		HitRate:         hitRate,
		TotalOperations:  totalOperations,
		GetLatency:      m.calculatePercentile(m.getLatencies),
		SetLatency:      m.calculatePercentile(m.setLatencies),
		DeleteLatency:   m.calculatePercentile(m.deleteLatencies),
		HotKeys:         m.copyHotKeys(),
	}
}

// Stats 统计信息结构
type Stats struct {
	Hits           int64            `json:"hits"`
	Misses         int64            `json:"misses"`
	Sets           int64            `json:"sets"`
	Deletes        int64            `json:"deletes"`
	Errors         int64            `json:"errors"`
	Loads          int64            `json:"loads"`
	LoadErrors     int64            `json:"load_errors"`
	LoadFailures   int64            `json:"load_failures"`
	LockAcquires   int64            `json:"lock_acquires"`
	LockConflicts  int64            `json:"lock_conflicts"`
	LockTimeouts   int64            `json:"lock_timeouts"`
	LockErrors     int64            `json:"lock_errors"`
	HitRate        float64          `json:"hit_rate"`
	TotalOperations int64            `json:"total_operations"`
	GetLatency     *LatencyStats    `json:"get_latency"`
	SetLatency     *LatencyStats    `json:"set_latency"`
	DeleteLatency  *LatencyStats    `json:"delete_latency"`
	HotKeys        map[string]int64 `json:"hot_keys"`
}

// LatencyStats 延迟统计
type LatencyStats struct {
	Count     int           `json:"count"`
	Min       time.Duration `json:"min"`
	Max       time.Duration `json:"max"`
	Avg       time.Duration `json:"avg"`
	P50       time.Duration `json:"p50"`
	P95       time.Duration `json:"p95"`
	P99       time.Duration `json:"p99"`
}

// calculatePercentile 计算延迟百分位数
func (m *Metrics) calculatePercentile(latencies []time.Duration) *LatencyStats {
	if len(latencies) == 0 {
		return &LatencyStats{}
	}

	// 复制并排序
	sorted := make([]time.Duration, len(latencies))
	copy(sorted, latencies)

	// 简单排序（实际应用中可以使用更高效的算法）
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	count := len(sorted)
	min := sorted[0]
	max := sorted[count-1]

	// 计算平均值
	var sum time.Duration
	for _, lat := range sorted {
		sum += lat
	}
	avg := sum / time.Duration(count)

	// 计算百分位数
	p50 := sorted[count*50/100]
	p95 := sorted[count*95/100]
	p99 := sorted[count*99/100]

	return &LatencyStats{
		Count: count,
		Min:   min,
		Max:   max,
		Avg:   avg,
		P50:   p50,
		P95:   p95,
		P99:   p99,
	}
}

// copyHotKeys 复制热键统计
func (m *Metrics) copyHotKeys() map[string]int64 {
	result := make(map[string]int64)
	for k, v := range m.hotKeys {
		result[k] = v
	}
	return result
}

// Reset 重置所有统计
func (m *Metrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	atomic.StoreInt64(&m.hits, 0)
	atomic.StoreInt64(&m.misses, 0)
	atomic.StoreInt64(&m.sets, 0)
	atomic.StoreInt64(&m.deletes, 0)
	atomic.StoreInt64(&m.errors, 0)
	atomic.StoreInt64(&m.loads, 0)
	atomic.StoreInt64(&m.loadErrors, 0)
	atomic.StoreInt64(&m.loadFailures, 0)
	atomic.StoreInt64(&m.lockAcquires, 0)
	atomic.StoreInt64(&m.lockConflicts, 0)
	atomic.StoreInt64(&m.lockTimeouts, 0)
	atomic.StoreInt64(&m.lockErrors, 0)

	m.getLatencies = m.getLatencies[:0]
	m.setLatencies = m.setLatencies[:0]
	m.deleteLatencies = m.deleteLatencies[:0]
	m.hotKeys = make(map[string]int64)
}
