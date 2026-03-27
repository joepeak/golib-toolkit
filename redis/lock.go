package redis

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisDistributedLock Redis 分布式锁实现
type RedisDistributedLock struct {
	client    redis.UniversalClient
	key       string
	value     string
	config    *LockConfig
	createdAt time.Time

	// 自动续期
	autoRefreshCtx    context.Context
	autoRefreshCancel context.CancelFunc
	autoRefreshWG    sync.WaitGroup
	isAutoRefreshing bool
	mu               sync.RWMutex
}

// NewRedisDistributedLock 创建新的分布式锁
func NewRedisDistributedLock(client redis.UniversalClient, key string, opts ...LockOption) *RedisDistributedLock {
	config := DefaultLockConfig()
	for _, opt := range opts {
		opt.apply(config)
	}

	// 生成唯一值
	value := generateUniqueValue()

	return &RedisDistributedLock{
		client: client,
		key:    key,
		value:  value,
		config: config,
	}
}

// Lock 获取锁（阻塞直到成功或超时）
func (l *RedisDistributedLock) Lock(ctx context.Context) error {
	for i := 0; i < l.config.RetryTimes; i++ {
		success, err := l.TryLock(ctx)
		if err != nil {
			return fmt.Errorf("lock failed: %w", err)
		}
		if success {
			return nil
		}

		// 等待重试
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(l.config.RetryDelay):
			continue
		}
	}

	return ErrLockFailed
}

// TryLock 尝试获取锁（非阻塞）
func (l *RedisDistributedLock) TryLock(ctx context.Context) (bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// 检查是否已经持有锁
	if l.isLocked() {
		return true, nil
	}

	// 尝试获取锁
	success, err := l.client.SetNX(ctx, l.key, l.value, l.config.Expiration).Result()
	if err != nil {
		return false, fmt.Errorf("set nx failed: %w", err)
	}

	if success {
		l.createdAt = time.Now()
		return true, nil
	}

	return false, nil
}

// Unlock 释放锁
func (l *RedisDistributedLock) Unlock(ctx context.Context) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if !l.isLocked() {
		return ErrLockNotHeld
	}

	// 停止自动续期
	l.stopAutoRefresh()

	// 使用 Lua 脚本确保只能释放自己持有的锁
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`

	result, err := l.client.Eval(ctx, script, []string{l.key}, l.value).Result()
	if err != nil {
		return fmt.Errorf("unlock failed: %w", err)
	}

	if result.(int64) == 0 {
		return ErrLockNotHeld
	}

	// 重置锁状态
	l.createdAt = time.Time{}
	return nil
}

// ForceUnlock 强制释放锁（谨慎使用）
func (l *RedisDistributedLock) ForceUnlock(ctx context.Context) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// 停止自动续期
	l.stopAutoRefresh()

	// 直接删除键
	err := l.client.Del(ctx, l.key).Err()
	if err != nil {
		return fmt.Errorf("force unlock failed: %w", err)
	}

	// 重置锁状态
	l.createdAt = time.Time{}
	return nil
}

// IsLocked 检查是否持有锁
func (l *RedisDistributedLock) IsLocked() bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.isLocked()
}

// isLocked 内部检查是否持有锁
func (l *RedisDistributedLock) isLocked() bool {
	return !l.createdAt.IsZero()
}

// GetTTL 获取锁的剩余时间
func (l *RedisDistributedLock) GetTTL(ctx context.Context) (time.Duration, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if !l.isLocked() {
		return 0, ErrLockNotHeld
	}

	ttl, err := l.client.TTL(ctx, l.key).Result()
	if err != nil {
		return 0, fmt.Errorf("get ttl failed: %w", err)
	}

	return ttl, nil
}

// Refresh 刷新锁的过期时间
func (l *RedisDistributedLock) Refresh(ctx context.Context) error {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if !l.isLocked() {
		return ErrLockNotHeld
	}

	// 使用 Lua 脚本确保只能刷新自己持有的锁
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("expire", KEYS[1], ARGV[2])
		else
			return 0
		end
	`

	result, err := l.client.Eval(ctx, script, []string{l.key}, l.value, int(l.config.Expiration.Seconds())).Result()
	if err != nil {
		return fmt.Errorf("refresh failed: %w", err)
	}

	if result.(int64) == 0 {
		return ErrLockNotHeld
	}

	return nil
}

// StartAutoRefresh 启动自动续期
func (l *RedisDistributedLock) StartAutoRefresh(ctx context.Context) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if !l.isLocked() {
		return ErrLockNotHeld
	}

	if l.isAutoRefreshing {
		return nil // 已经在自动续期
	}

	if !l.config.AutoExtend {
		return nil // 未启用自动续期
	}

	l.autoRefreshCtx, l.autoRefreshCancel = context.WithCancel(ctx)
	l.isAutoRefreshing = true
	l.autoRefreshWG.Add(1)

	go l.autoRefreshLoop()

	return nil
}

// StopAutoRefresh 停止自动续期
func (l *RedisDistributedLock) StopAutoRefresh() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.stopAutoRefresh()
}

// stopAutoRefresh 内部停止自动续期
func (l *RedisDistributedLock) stopAutoRefresh() {
	if !l.isAutoRefreshing {
		return
	}

	l.autoRefreshCancel()
	l.autoRefreshWG.Wait()
	l.isAutoRefreshing = false
}

// autoRefreshLoop 自动续期循环
func (l *RedisDistributedLock) autoRefreshLoop() {
	defer l.autoRefreshWG.Done()

	ticker := time.NewTicker(l.config.Expiration - l.config.ExtendBefore)
	defer ticker.Stop()

	for {
		select {
		case <-l.autoRefreshCtx.Done():
			return
		case <-ticker.C:
			if err := l.Refresh(l.autoRefreshCtx); err != nil {
				// 续期失败，记录日志但不退出循环
				// 实际应用中可以添加日志记录
				continue
			}
		}
	}
}

// GetKey 获取锁的键
func (l *RedisDistributedLock) GetKey() string {
	return l.key
}

// GetValue 获取锁的值
func (l *RedisDistributedLock) GetValue() string {
	return l.value
}

// GetCreatedAt 获取锁的创建时间
func (l *RedisDistributedLock) GetCreatedAt() time.Time {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.createdAt
}

// Close 关闭锁资源
func (l *RedisDistributedLock) Close() error {
	l.StopAutoRefresh()
	return nil
}

// generateUniqueValue 生成唯一值
func generateUniqueValue() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// RedisLockManager Redis 锁管理器
type RedisLockManager struct {
	client  redis.UniversalClient
	metrics *Metrics
}

// NewRedisLockManager 创建锁管理器
func NewRedisLockManager(client redis.UniversalClient) *RedisLockManager {
	return &RedisLockManager{
		client:  client,
		metrics: NewMetrics(),
	}
}

// NewLock 创建新的锁
func (lm *RedisLockManager) NewLock(key string, opts ...LockOption) DistributedLock {
	return NewRedisDistributedLock(lm.client, key, opts...)
}

// NewLockWithConfig 使用指定配置创建锁
func (lm *RedisLockManager) NewLockWithConfig(key string, config *LockConfig) DistributedLock {
	return NewRedisDistributedLock(lm.client, key, LockOptionFunc(func(c *LockConfig) {
		*c = *config
	}))
}

// LockMany 批量获取锁
func (lm *RedisLockManager) LockMany(ctx context.Context, keys []string, opts ...LockOption) (map[string]DistributedLock, error) {
	locks := make(map[string]DistributedLock)
	var firstErr error

	for _, key := range keys {
		lock := lm.NewLock(key, opts...)
		if err := lock.Lock(ctx); err != nil {
			if firstErr == nil {
				firstErr = err
			}
			// 释放已获取的锁
			lm.UnlockMany(ctx, locks)
			return nil, firstErr
		} else {
			locks[key] = lock
		}
	}

	return locks, nil
}

// UnlockMany 批量释放锁
func (lm *RedisLockManager) UnlockMany(ctx context.Context, locks map[string]DistributedLock) error {
	var firstErr error

	for _, lock := range locks {
		if err := lock.Unlock(ctx); err != nil {
			if firstErr == nil {
				firstErr = err
			}
		}
	}

	return firstErr
}

// IsLocked 检查指定键是否被锁定
func (lm *RedisLockManager) IsLocked(ctx context.Context, key string) (bool, error) {
	exists, err := lm.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("check lock failed: %w", err)
	}
	return exists > 0, nil
}

// GetLockInfo 获取锁信息
func (lm *RedisLockManager) GetLockInfo(ctx context.Context, key string) (*LockInfo, error) {
	pipe := lm.client.Pipeline()
	valueCmd := pipe.Get(ctx, key)
	ttlCmd := pipe.TTL(ctx, key)

	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("get lock info failed: %w", err)
	}

	value, _ := valueCmd.Result()
	ttl, _ := ttlCmd.Result()

	if value == "" {
		return &LockInfo{
			Key:       key,
			IsExpired: true,
		}, nil
	}

	return &LockInfo{
		Key:       key,
		Value:     value,
		CreatedAt: time.Now(), // 无法准确获取创建时间
		ExpiresAt: time.Now().Add(ttl),
		TTL:       ttl,
		IsExpired: ttl <= 0,
	}, nil
}

// CleanupExpiredLocks 清理过期锁
func (lm *RedisLockManager) CleanupExpiredLocks(ctx context.Context) error {
	// 这个方法需要根据具体的锁键模式来实现
	// 这里提供一个基础实现，实际使用时可能需要自定义
	pattern := "lock:*"
	
	var cursor uint64
	var keys []string

	for {
		var err error
		var scanKeys []string
		scanKeys, cursor, err = lm.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return fmt.Errorf("scan locks failed: %w", err)
		}

		keys = append(keys, scanKeys...)

		if cursor == 0 {
			break
		}
	}

	if len(keys) > 0 {
		// 检查每个键的 TTL，删除已过期的
		pipe := lm.client.Pipeline()
		for _, key := range keys {
			pipe.TTL(ctx, key)
		}

		cmds, err := pipe.Exec(ctx)
		if err != nil {
			return fmt.Errorf("check ttl failed: %w", err)
		}

		expiredKeys := make([]string, 0)
		for i, cmd := range cmds {
			ttl := cmd.(*redis.DurationCmd).Val()
			if ttl <= 0 {
				expiredKeys = append(expiredKeys, keys[i])
			}
		}

		if len(expiredKeys) > 0 {
			_, err := lm.client.Del(ctx, expiredKeys...).Result()
			if err != nil {
				return fmt.Errorf("delete expired locks failed: %w", err)
			}
		}
	}

	return nil
}

// Close 关闭锁管理器
func (lm *RedisLockManager) Close() error {
	// Redis 客户端由外部管理，这里不需要关闭
	return nil
}

// GetMetrics 获取锁管理器的指标
func (lm *RedisLockManager) GetMetrics() *Metrics {
	return lm.metrics
}
