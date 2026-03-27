package redis

import (
	"context"
	"encoding/json"
	"errors"
	"time"
)

// GetOrLoad 带自动加载的获取
func (c *RedisCache[T]) GetOrLoad(ctx context.Context, key string, loader LoaderFunc[T]) (*T, error) {
	// 先尝试从缓存获取
	value, err := c.Get(ctx, key)
	if err == nil {
		return value, nil
	}

	if !errors.Is(err, ErrCacheMiss) {
		return nil, err
	}

	// 使用提供的 loader
	if loader == nil {
		return nil, ErrLoadFailed
	}

	// 防缓存击穿
	if c.config.EnableSingleFlight {
		return c.getWithSingleFlight(ctx, key, loader)
	}

	return c.loadAndCache(ctx, key, loader)
}

// GetManyOrLoad 批量自动加载
func (c *RedisCache[T]) GetManyOrLoad(ctx context.Context, keys []string, loader BatchLoaderFunc[T]) (map[string]*T, error) {
	// 先批量获取缓存
	result, err := c.GetMany(ctx, keys)
	if err != nil {
		return nil, err
	}

	// 找出缺失的 key
	missingKeys := make([]string, 0)
	for _, key := range keys {
		if _, exists := result[key]; !exists {
			missingKeys = append(missingKeys, key)
		}
	}

	if len(missingKeys) == 0 {
		return result, nil
	}

	// 使用提供的 loader
	if loader == nil {
		return nil, ErrLoadFailed
	}

	// 加载缺失的数据
	loadedData, err := loader(ctx, missingKeys)
	if err != nil {
		return nil, err
	}

	// 批量缓存
	if err := c.SetMany(ctx, loadedData, c.config.DefaultTTL); err != nil {
		// 记录错误但不影响返回
		if c.config.EnableMetrics {
			c.metrics.RecordError()
		}
	}

	// 合并结果
	for k, v := range loadedData {
		result[k] = v
	}

	return result, nil
}

// Refresh 刷新缓存
func (c *RedisCache[T]) Refresh(ctx context.Context, key string, loader LoaderFunc[T], ttl time.Duration) error {
	if loader == nil {
		return ErrLoadFailed
	}

	// 异步刷新，避免阻塞
	go func() {
		backgroundCtx := context.Background()
		value, err := loader(backgroundCtx, key)
		if err == nil {
			c.Set(backgroundCtx, key, value, ttl)
		}
	}()

	return nil
}

// WarmUp 预热缓存
func (c *RedisCache[T]) WarmUp(ctx context.Context, keys []string, loader BatchLoaderFunc[T]) error {
	if loader == nil {
		return ErrLoadFailed
	}

	// 批量加载并缓存
	loadedData, err := loader(ctx, keys)
	if err != nil {
		return err
	}

	return c.SetMany(ctx, loadedData, c.config.DefaultTTL)
}

// AsyncSet 异步设置缓存
func (c *RedisCache[T]) AsyncSet(ctx context.Context, key string, value *T, ttl time.Duration) error {
	go func() {
		c.Set(ctx, key, value, ttl)
	}()
	return nil
}

// AsyncDelete 异步删除缓存
func (c *RedisCache[T]) AsyncDelete(ctx context.Context, key string) error {
	go func() {
		c.Delete(ctx, key)
	}()
	return nil
}

// getWithSingleFlight 使用 singleflight 防止缓存击穿
func (c *RedisCache[T]) getWithSingleFlight(ctx context.Context, key string, loader LoaderFunc[T]) (*T, error) {
	result, err := c.sf.Do(key, func() (interface{}, error) {
		return c.loadAndCache(ctx, key, loader)
	})

	if err != nil {
		return nil, err
	}

	return result.(*T), nil
}

// loadAndCache 加载数据并缓存
func (c *RedisCache[T]) loadAndCache(ctx context.Context, key string, loader LoaderFunc[T]) (*T, error) {
	// 再次检查缓存（双重检查）
	value, err := c.Get(ctx, key)
	if err == nil {
		return value, nil
	}

	if !errors.Is(err, ErrCacheMiss) {
		return nil, err
	}

	// 调用加载函数
	value, err = loader(ctx, key)
	if err != nil {
		// 缓存空值防止缓存穿透
		c.cacheNull(ctx, key)
		if c.config.EnableMetrics {
			c.metrics.RecordLoadError()
		}
		return nil, ErrLoadFailed
	}

	// 缓存数据
	if err := c.Set(ctx, key, value, c.config.DefaultTTL); err != nil {
		// 记录错误但不影响返回
		if c.config.EnableMetrics {
			c.metrics.RecordError()
		}
	}

	if c.config.EnableMetrics {
		c.metrics.RecordLoad()
	}

	return value, nil
}

// cacheNull 缓存空值，防止缓存穿透
func (c *RedisCache[T]) cacheNull(ctx context.Context, key string) {
	cacheItem := CacheItem[T]{
		IsNull:   true,
		NullAt:   time.Now(),
		CreatedAt: time.Now(),
	}

	data, _ := json.Marshal(cacheItem)
	c.client.Set(ctx, c.buildKey(key), data, c.config.NullValueTTL)
}
