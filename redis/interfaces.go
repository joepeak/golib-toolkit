package redis

import (
	"context"
	"time"
)

// BasicCache 基础缓存接口 - 提供简单的缓存操作
type BasicCache[T any] interface {
	// 基础操作
	Get(ctx context.Context, key string) (*T, error)
	Set(ctx context.Context, key string, value *T, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)

	// 批量操作
	GetMany(ctx context.Context, keys []string) (map[string]*T, error)
	SetMany(ctx context.Context, items map[string]*T, ttl time.Duration) error
	DeleteMany(ctx context.Context, keys []string) error

	// 模式匹配操作
	Invalidate(ctx context.Context, pattern string) error
	Scan(ctx context.Context, pattern string) ([]string, error)

	// 配置和监控
	Config() *CacheConfig
	Metrics() *Metrics
}

// AdvancedCache 高级缓存接口 - 包含自动加载功能
type AdvancedCache[T any] interface {
	BasicCache[T]

	// 可选加载函数的方法
	GetOrLoad(ctx context.Context, key string, loader LoaderFunc[T]) (*T, error)
	GetManyOrLoad(ctx context.Context, keys []string, loader BatchLoaderFunc[T]) (map[string]*T, error)

	// 刷新和预热
	Refresh(ctx context.Context, key string, loader LoaderFunc[T], ttl time.Duration) error
	WarmUp(ctx context.Context, keys []string, loader BatchLoaderFunc[T]) error

	// 异步操作
	AsyncSet(ctx context.Context, key string, value *T, ttl time.Duration) error
	AsyncDelete(ctx context.Context, key string) error
}

// Cache 缓存客户端接口 - 组合所有功能
type Cache[T any] interface {
	BasicCache[T]
	AdvancedCache[T]

	// 生命周期管理
	Close() error
	Ping(ctx context.Context) error

	// 高级功能
	GetWithTTL(ctx context.Context, key string) (*T, time.Duration, error)
	SetIfNotExists(ctx context.Context, key string, value *T, ttl time.Duration) (bool, error)
	Increment(ctx context.Context, key string, delta int64) (int64, error)
	Decrement(ctx context.Context, key string, delta int64) (int64, error)
}

// DistributedLock 分布式锁接口
type DistributedLock interface {
	// 基础锁操作
	Lock(ctx context.Context) error
	TryLock(ctx context.Context) (bool, error)
	Unlock(ctx context.Context) error
	ForceUnlock(ctx context.Context) error

	// 锁状态
	IsLocked() bool
	GetTTL(ctx context.Context) (time.Duration, error)
	Refresh(ctx context.Context) error

	// 锁信息
	GetKey() string
	GetValue() string
	GetCreatedAt() time.Time

	// 自动续期
	StartAutoRefresh(ctx context.Context) error
	StopAutoRefresh()

	// 生命周期
	Close() error
}

// LockManager 分布式锁管理器接口
type LockManager interface {
	// 创建锁
	NewLock(key string, opts ...LockOption) DistributedLock
	NewLockWithConfig(key string, config *LockConfig) DistributedLock

	// 批量操作
	LockMany(ctx context.Context, keys []string, opts ...LockOption) (map[string]DistributedLock, error)
	UnlockMany(ctx context.Context, locks map[string]DistributedLock) error

	// 锁状态查询
	IsLocked(ctx context.Context, key string) (bool, error)
	GetLockInfo(ctx context.Context, key string) (*LockInfo, error)

	// 清理过期锁
	CleanupExpiredLocks(ctx context.Context) error

	// 生命周期
	Close() error
}

// LockInfo 锁信息
type LockInfo struct {
	Key         string        `json:"key"`
	Value       string        `json:"value"`
	Owner       string        `json:"owner"`
	CreatedAt   time.Time     `json:"created_at"`
	ExpiresAt   time.Time     `json:"expires_at"`
	TTL         time.Duration `json:"ttl"`
	IsExpired   bool          `json:"is_expired"`
}

// Option 配置选项类型
type Option interface {
	apply(config interface{})
}

// LockOption 锁选项类型
type LockOption interface {
	apply(config *LockConfig)
}

// CacheOption 缓存选项类型
type CacheOption interface {
	apply(config *CacheConfig)
}
