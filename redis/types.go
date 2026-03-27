package redis

import (
	"context"
	"errors"
	"time"
)

var (
	ErrCacheMiss      = errors.New("cache: key not found")
	ErrLockFailed     = errors.New("cache: failed to acquire lock")
	ErrLockNotHeld    = errors.New("lock: lock not held by this instance")
	ErrLockExpired    = errors.New("lock: lock has expired")
	ErrLoadFailed     = errors.New("cache: loader function failed")
	ErrInvalidKey     = errors.New("cache: invalid key")
	ErrInvalidValue   = errors.New("cache: invalid value")
	ErrInvalidTTL     = errors.New("cache: invalid ttl")
)

// CacheItem 缓存项结构
type CacheItem[T any] struct {
	Value     *T        `json:"value,omitempty"`
	IsNull     bool      `json:"is_null,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	ExpiresAt  time.Time `json:"expires_at,omitempty"`
	NullAt     time.Time `json:"null_at,omitempty"`
}

// IsExpired 检查缓存项是否过期
func (ci *CacheItem[T]) IsExpired() bool {
	return !ci.ExpiresAt.IsZero() && time.Now().After(ci.ExpiresAt)
}

// LoaderFunc 单个数据加载函数
type LoaderFunc[T any] func(ctx context.Context, key string) (*T, error)

// BatchLoaderFunc 批量数据加载函数
type BatchLoaderFunc[T any] func(ctx context.Context, keys []string) (map[string]*T, error)

// CacheConfig 缓存配置
type CacheConfig struct {
	// 基础配置
	DefaultTTL   time.Duration `json:"default_ttl"`    // 默认过期时间
	NullValueTTL  time.Duration `json:"null_value_ttl"`  // 空值缓存时间，防止缓存穿透
	KeyPrefix     string        `json:"key_prefix"`     // 键前缀

	// 锁配置
	LockTimeout    time.Duration `json:"lock_timeout"`    // 分布式锁超时时间
	LockRetryDelay time.Duration `json:"lock_retry_delay"` // 锁重试延迟

	// 性能配置
	MaxRetries          int  `json:"max_retries"`           // 最大重试次数
	EnableSingleFlight  bool `json:"enable_singleflight"`  // 是否启用singleflight
	EnableMetrics       bool `json:"enable_metrics"`        // 是否启用指标统计
	EnableHotKeyDetect  bool `json:"enable_hotkey_detect"` // 是否启用热键检测
	HotKeyThreshold    int64 `json:"hotkey_threshold"`    // 热键阈值（每秒访问次数）

	// 序列化配置
	EnableCompression bool `json:"enable_compression"` // 是否启用压缩
	CompressionLevel int  `json:"compression_level"`  // 压缩级别
}

// DefaultConfig 返回默认配置
func DefaultConfig() *CacheConfig {
	return &CacheConfig{
		DefaultTTL:         5 * time.Minute,
		NullValueTTL:       1 * time.Minute,
		KeyPrefix:          "cache:",
		LockTimeout:        5 * time.Second,
		LockRetryDelay:     100 * time.Millisecond,
		MaxRetries:         3,
		EnableSingleFlight:  true,
		EnableMetrics:      true,
		EnableHotKeyDetect: false,
		HotKeyThreshold:    1000,
		EnableCompression:  false,
		CompressionLevel:   6,
	}
}

// LockConfig 分布式锁配置
type LockConfig struct {
	Expiration    time.Duration `json:"expiration"`     // 锁过期时间
	RetryTimes    int           `json:"retry_times"`     // 重试次数
	RetryDelay    time.Duration `json:"retry_delay"`     // 重试延迟
	AutoExtend    bool          `json:"auto_extend"`     // 自动续期
	ExtendBefore  time.Duration `json:"extend_before"`   // 过期前多久续期
}

// DefaultLockConfig 返回默认锁配置
func DefaultLockConfig() *LockConfig {
	return &LockConfig{
		Expiration:   30 * time.Second,
		RetryTimes:   3,
		RetryDelay:   100 * time.Millisecond,
		AutoExtend:   false,
		ExtendBefore: 5 * time.Second,
	}
}

// RedisConfig Redis 连接配置
type RedisConfig struct {
	// 基础连接配置
	Addr         string `mapstructure:"addr"`
	Password     string `mapstructure:"password"`
	DB           int    `mapstructure:"db"`
	Username     string `mapstructure:"username"`

	// 集群配置
	EnabledCluster bool     `mapstructure:"enabledCluster"`
	ClusterAddrs  []string `mapstructure:"cluster.addrs"`

	// TLS 配置
	EnabledTLS bool `mapstructure:"enabledTls"`

	// 连接池配置
	PoolSize     int `mapstructure:"poolSize"`
	MinIdleConns int `mapstructure:"minIdleConns"`
	MaxRetries   int `mapstructure:"maxRetries"`
}

// DefaultRedisConfig 返回默认 Redis 配置
func DefaultRedisConfig() *RedisConfig {
	return &RedisConfig{
		Addr:          "localhost:6379",
		Password:      "",
		DB:            0,
		Username:      "",
		EnabledCluster: false,
		ClusterAddrs:  []string{},
		EnabledTLS:    false,
		PoolSize:      10,
		MinIdleConns:  5,
		MaxRetries:    3,
	}
}
