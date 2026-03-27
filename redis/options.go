package redis

import "time"

// LockOptionFunc 锁选项函数类型
type LockOptionFunc func(*LockConfig)

func (f LockOptionFunc) apply(config *LockConfig) {
	f(config)
}

// CacheOptionFunc 缓存选项函数类型
type CacheOptionFunc func(*CacheConfig)

func (f CacheOptionFunc) apply(config *CacheConfig) {
	f(config)
}

// 锁选项实现
func WithExpiration(expiration time.Duration) LockOption {
	return LockOptionFunc(func(config *LockConfig) {
		config.Expiration = expiration
	})
}

func WithRetryTimes(retryTimes int) LockOption {
	return LockOptionFunc(func(config *LockConfig) {
		config.RetryTimes = retryTimes
	})
}

func WithRetryDelay(retryDelay time.Duration) LockOption {
	return LockOptionFunc(func(config *LockConfig) {
		config.RetryDelay = retryDelay
	})
}

func WithAutoExtend(autoExtend bool) LockOption {
	return LockOptionFunc(func(config *LockConfig) {
		config.AutoExtend = autoExtend
	})
}

func WithExtendBefore(extendBefore time.Duration) LockOption {
	return LockOptionFunc(func(config *LockConfig) {
		config.ExtendBefore = extendBefore
	})
}

// 缓存选项实现
func WithDefaultTTL(ttl time.Duration) CacheOption {
	return CacheOptionFunc(func(config *CacheConfig) {
		config.DefaultTTL = ttl
	})
}

func WithNullValueTTL(ttl time.Duration) CacheOption {
	return CacheOptionFunc(func(config *CacheConfig) {
		config.NullValueTTL = ttl
	})
}

func WithKeyPrefix(prefix string) CacheOption {
	return CacheOptionFunc(func(config *CacheConfig) {
		config.KeyPrefix = prefix
	})
}

func WithLockTimeout(timeout time.Duration) CacheOption {
	return CacheOptionFunc(func(config *CacheConfig) {
		config.LockTimeout = timeout
	})
}

func WithLockRetryDelay(delay time.Duration) CacheOption {
	return CacheOptionFunc(func(config *CacheConfig) {
		config.LockRetryDelay = delay
	})
}

func WithMaxRetries(retries int) CacheOption {
	return CacheOptionFunc(func(config *CacheConfig) {
		config.MaxRetries = retries
	})
}

func WithEnableSingleFlight(enable bool) CacheOption {
	return CacheOptionFunc(func(config *CacheConfig) {
		config.EnableSingleFlight = enable
	})
}

func WithEnableMetrics(enable bool) CacheOption {
	return CacheOptionFunc(func(config *CacheConfig) {
		config.EnableMetrics = enable
	})
}

func WithEnableHotKeyDetect(enable bool) CacheOption {
	return CacheOptionFunc(func(config *CacheConfig) {
		config.EnableHotKeyDetect = enable
	})
}

func WithHotKeyThreshold(threshold int64) CacheOption {
	return CacheOptionFunc(func(config *CacheConfig) {
		config.HotKeyThreshold = threshold
	})
}

func WithEnableCompression(enable bool) CacheOption {
	return CacheOptionFunc(func(config *CacheConfig) {
		config.EnableCompression = enable
	})
}

func WithCompressionLevel(level int) CacheOption {
	return CacheOptionFunc(func(config *CacheConfig) {
		config.CompressionLevel = level
	})
}
