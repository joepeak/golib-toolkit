package redis

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisCache Redis 缓存实现
type RedisCache[T any] struct {
	client  redis.UniversalClient
	config  *CacheConfig
	metrics *Metrics
	sf      *SingleFlight
	hotKeys *HotKeyDetector

	// 可选的高级功能
	loader     LoaderFunc[T]
	batchLoader BatchLoaderFunc[T]

	mu sync.RWMutex
}

// NewRedisCache 创建新的 Redis 缓存实例
func NewRedisCache[T any](client redis.UniversalClient, opts ...CacheOption) *RedisCache[T] {
	config := DefaultConfig()
	for _, opt := range opts {
		opt.apply(config)
	}

	cache := &RedisCache[T]{
		client:  client,
		config:  config,
		metrics: NewMetrics(),
		sf:      NewSingleFlight(),
		hotKeys: NewHotKeyDetector(config.HotKeyThreshold),
	}

	// 启动热键检测
	if config.EnableHotKeyDetect {
		go cache.hotKeys.Start(context.Background())
	}

	return cache
}

// Config 获取缓存配置
func (c *RedisCache[T]) Config() *CacheConfig {
	return c.config
}

// Metrics 获取缓存指标
func (c *RedisCache[T]) Metrics() *Metrics {
	return c.metrics
}

// Close 关闭缓存
func (c *RedisCache[T]) Close() error {
	c.hotKeys.Stop()
	return nil
}

// Ping 检查 Redis 连接
func (c *RedisCache[T]) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

// buildKey 构建缓存键
func (c *RedisCache[T]) buildKey(key string) string {
	if c.config.KeyPrefix == "" {
		return key
	}
	return c.config.KeyPrefix + key
}

// Get 获取缓存
func (c *RedisCache[T]) Get(ctx context.Context, key string) (*T, error) {
	start := time.Now()
	defer func() {
		if c.config.EnableMetrics {
			c.metrics.RecordLatency("get", time.Since(start))
		}
	}()

	cacheKey := c.buildKey(key)
	data, err := c.client.Get(ctx, cacheKey).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			if c.config.EnableMetrics {
				c.metrics.RecordMiss()
			}
			return nil, ErrCacheMiss
		}
		if c.config.EnableMetrics {
			c.metrics.RecordError()
		}
		return nil, err
	}

	var cacheItem CacheItem[T]
	if err := json.Unmarshal(data, &cacheItem); err != nil {
		if c.config.EnableMetrics {
			c.metrics.RecordError()
		}
		return nil, err
	}

	if cacheItem.IsNull {
		if c.config.EnableMetrics {
			c.metrics.RecordMiss()
		}
		return nil, ErrCacheMiss
	}

	if cacheItem.IsExpired() {
		// 异步删除过期键
		go c.client.Del(context.Background(), cacheKey)
		if c.config.EnableMetrics {
			c.metrics.RecordMiss()
		}
		return nil, ErrCacheMiss
	}

	if c.config.EnableMetrics {
		c.metrics.RecordHit()
	}

	// 热键检测
	if c.config.EnableHotKeyDetect {
		c.hotKeys.RecordAccess(key)
	}

	return cacheItem.Value, nil
}

// Set 设置缓存
func (c *RedisCache[T]) Set(ctx context.Context, key string, value *T, ttl time.Duration) error {
	start := time.Now()
	defer func() {
		if c.config.EnableMetrics {
			c.metrics.RecordLatency("set", time.Since(start))
		}
	}()

	if ttl == 0 {
		ttl = c.config.DefaultTTL
	}

	cacheKey := c.buildKey(key)
	var expiresAt time.Time
	if ttl > 0 {
		expiresAt = time.Now().Add(ttl)
	}

	cacheItem := CacheItem[T]{
		Value:     value,
		IsNull:    false,
		CreatedAt: time.Now(),
		ExpiresAt: expiresAt,
	}

	data, err := json.Marshal(cacheItem)
	if err != nil {
		if c.config.EnableMetrics {
			c.metrics.RecordError()
		}
		return err
	}

	err = c.client.Set(ctx, cacheKey, data, ttl).Err()
	if err != nil {
		if c.config.EnableMetrics {
			c.metrics.RecordError()
		}
		return err
	}

	if c.config.EnableMetrics {
		c.metrics.RecordSet()
	}

	return nil
}

// Delete 删除缓存
func (c *RedisCache[T]) Delete(ctx context.Context, key string) error {
	start := time.Now()
	defer func() {
		if c.config.EnableMetrics {
			c.metrics.RecordLatency("delete", time.Since(start))
		}
	}()

	cacheKey := c.buildKey(key)
	err := c.client.Del(ctx, cacheKey).Err()
	if err != nil {
		if c.config.EnableMetrics {
			c.metrics.RecordError()
		}
		return err
	}

	if c.config.EnableMetrics {
		c.metrics.RecordDelete()
	}

	return nil
}

// Exists 检查缓存是否存在
func (c *RedisCache[T]) Exists(ctx context.Context, key string) (bool, error) {
	cacheKey := c.buildKey(key)
	exists, err := c.client.Exists(ctx, cacheKey).Result()
	if err != nil {
		if c.config.EnableMetrics {
			c.metrics.RecordError()
		}
		return false, err
	}
	return exists > 0, nil
}

// GetMany 批量获取缓存
func (c *RedisCache[T]) GetMany(ctx context.Context, keys []string) (map[string]*T, error) {
	if len(keys) == 0 {
		return make(map[string]*T), nil
	}

	start := time.Now()
	defer func() {
		if c.config.EnableMetrics {
			c.metrics.RecordLatency("get_many", time.Since(start))
		}
	}()

	cacheKeys := make([]string, len(keys))
	for i, key := range keys {
		cacheKeys[i] = c.buildKey(key)
	}

	pipe := c.client.Pipeline()
	cmds := make(map[string]*redis.StringCmd)
	for i, cacheKey := range cacheKeys {
		cmds[keys[i]] = pipe.Get(ctx, cacheKey)
	}

	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		if c.config.EnableMetrics {
			c.metrics.RecordError()
		}
		return nil, err
	}

	result := make(map[string]*T)
	for key, cmd := range cmds {
		data, err := cmd.Bytes()
		if err != nil {
			if errors.Is(err, redis.Nil) {
				continue
			}
			if c.config.EnableMetrics {
				c.metrics.RecordError()
			}
			return nil, err
		}

		var cacheItem CacheItem[T]
		if err := json.Unmarshal(data, &cacheItem); err != nil {
			if c.config.EnableMetrics {
				c.metrics.RecordError()
			}
			return nil, err
		}

		if !cacheItem.IsNull && !cacheItem.IsExpired() {
			result[key] = cacheItem.Value

			// 热键检测
			if c.config.EnableHotKeyDetect {
				c.hotKeys.RecordAccess(key)
			}
		}
	}

	if c.config.EnableMetrics {
		hitCount := len(result)
		missCount := len(keys) - len(result)
		for i := 0; i < hitCount; i++ {
			c.metrics.RecordHit()
		}
		for i := 0; i < missCount; i++ {
			c.metrics.RecordMiss()
		}
	}

	return result, nil
}

// SetMany 批量设置缓存
func (c *RedisCache[T]) SetMany(ctx context.Context, items map[string]*T, ttl time.Duration) error {
	if len(items) == 0 {
		return nil
	}

	start := time.Now()
	defer func() {
		if c.config.EnableMetrics {
			c.metrics.RecordLatency("set_many", time.Since(start))
		}
	}()

	if ttl == 0 {
		ttl = c.config.DefaultTTL
	}

	pipe := c.client.Pipeline()
	var expiresAt time.Time
	if ttl > 0 {
		expiresAt = time.Now().Add(ttl)
	}

	for key, value := range items {
		cacheKey := c.buildKey(key)
		cacheItem := CacheItem[T]{
			Value:     value,
			IsNull:    false,
			CreatedAt: time.Now(),
			ExpiresAt: expiresAt,
		}

		data, err := json.Marshal(cacheItem)
		if err != nil {
			if c.config.EnableMetrics {
				c.metrics.RecordError()
			}
			continue
		}

		pipe.Set(ctx, cacheKey, data, ttl)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		if c.config.EnableMetrics {
			c.metrics.RecordError()
		}
		return err
	}

	if c.config.EnableMetrics {
		for i := 0; i < len(items); i++ {
			c.metrics.RecordSet()
		}
	}

	return nil
}

// DeleteMany 批量删除缓存
func (c *RedisCache[T]) DeleteMany(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return nil
	}

	start := time.Now()
	defer func() {
		if c.config.EnableMetrics {
			c.metrics.RecordLatency("delete_many", time.Since(start))
		}
	}()

	cacheKeys := make([]string, len(keys))
	for i, key := range keys {
		cacheKeys[i] = c.buildKey(key)
	}

	err := c.client.Del(ctx, cacheKeys...).Err()
	if err != nil {
		if c.config.EnableMetrics {
			c.metrics.RecordError()
		}
		return err
	}

	if c.config.EnableMetrics {
		for i := 0; i < len(keys); i++ {
			c.metrics.RecordDelete()
		}
	}

	return nil
}

// Invalidate 失效缓存（支持模式匹配）
func (c *RedisCache[T]) Invalidate(ctx context.Context, pattern string) error {
	start := time.Now()
	defer func() {
		if c.config.EnableMetrics {
			c.metrics.RecordLatency("invalidate", time.Since(start))
		}
	}()

	cachePattern := c.buildKey(pattern)
	var cursor uint64
	var keys []string

	for {
		var err error
		var scanKeys []string
		scanKeys, cursor, err = c.client.Scan(ctx, cursor, cachePattern, 100).Result()
		if err != nil {
			if c.config.EnableMetrics {
				c.metrics.RecordError()
			}
			return err
		}

		keys = append(keys, scanKeys...)

		if cursor == 0 {
			break
		}
	}

	if len(keys) > 0 {
		err := c.client.Del(ctx, keys...).Err()
		if err != nil {
			if c.config.EnableMetrics {
				c.metrics.RecordError()
			}
			return err
		}

		if c.config.EnableMetrics {
			for i := 0; i < len(keys); i++ {
				c.metrics.RecordDelete()
			}
		}
	}

	return nil
}

// Scan 扫描缓存键
func (c *RedisCache[T]) Scan(ctx context.Context, pattern string) ([]string, error) {
	cachePattern := c.buildKey(pattern)
	var cursor uint64
	var keys []string
	prefix := c.config.KeyPrefix

	for {
		var err error
		var scanKeys []string
		scanKeys, cursor, err = c.client.Scan(ctx, cursor, cachePattern, 100).Result()
		if err != nil {
			if c.config.EnableMetrics {
				c.metrics.RecordError()
			}
			return nil, err
		}

		keys = append(keys, scanKeys...)

		if cursor == 0 {
			break
		}
	}

	// 移除前缀
	if prefix != "" {
		for i, key := range keys {
			keys[i] = strings.TrimPrefix(key, prefix)
		}
	}

	return keys, nil
}

// GetWithTTL 获取缓存并返回剩余时间
func (c *RedisCache[T]) GetWithTTL(ctx context.Context, key string) (*T, time.Duration, error) {
	cacheKey := c.buildKey(key)
	pipe := c.client.Pipeline()
	dataCmd := pipe.Get(ctx, cacheKey)
	ttlCmd := pipe.TTL(ctx, cacheKey)

	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, 0, err
	}

	data, err := dataCmd.Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, 0, ErrCacheMiss
		}
		return nil, 0, err
	}

	ttl, _ := ttlCmd.Result()
	if ttl <= 0 {
		return nil, 0, ErrCacheMiss
	}

	var cacheItem CacheItem[T]
	if err := json.Unmarshal(data, &cacheItem); err != nil {
		return nil, 0, err
	}

	if cacheItem.IsNull {
		return nil, 0, ErrCacheMiss
	}

	return cacheItem.Value, ttl, nil
}

// SetIfNotExists 如果键不存在则设置
func (c *RedisCache[T]) SetIfNotExists(ctx context.Context, key string, value *T, ttl time.Duration) (bool, error) {
	if ttl == 0 {
		ttl = c.config.DefaultTTL
	}

	cacheKey := c.buildKey(key)
	var expiresAt time.Time
	if ttl > 0 {
		expiresAt = time.Now().Add(ttl)
	}

	cacheItem := CacheItem[T]{
		Value:     value,
		IsNull:    false,
		CreatedAt: time.Now(),
		ExpiresAt: expiresAt,
	}

	data, err := json.Marshal(cacheItem)
	if err != nil {
		return false, err
	}

	result, err := c.client.SetNX(ctx, cacheKey, data, ttl).Result()
	if err != nil {
		return false, err
	}

	if result && c.config.EnableMetrics {
		c.metrics.RecordSet()
	}

	return result, nil
}

// Increment 递增计数器
func (c *RedisCache[T]) Increment(ctx context.Context, key string, delta int64) (int64, error) {
	cacheKey := c.buildKey(key)
	result, err := c.client.IncrBy(ctx, cacheKey, delta).Result()
	if err != nil {
		if c.config.EnableMetrics {
			c.metrics.RecordError()
		}
		return 0, err
	}

	if c.config.EnableMetrics {
		c.metrics.RecordSet()
	}

	return result, nil
}

// Decrement 递减计数器
func (c *RedisCache[T]) Decrement(ctx context.Context, key string, delta int64) (int64, error) {
	cacheKey := c.buildKey(key)
	result, err := c.client.DecrBy(ctx, cacheKey, delta).Result()
	if err != nil {
		if c.config.EnableMetrics {
			c.metrics.RecordError()
		}
		return 0, err
	}

	if c.config.EnableMetrics {
		c.metrics.RecordSet()
	}

	return result, nil
}
