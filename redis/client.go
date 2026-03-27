package redis

import (
	"context"
	"crypto/tls"
	"fmt"
	"reflect"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

// Client Redis 客户端包装器
type Client struct {
	cache   Cache[any]
	lockMgr LockManager
	client  redis.UniversalClient
	config  *RedisConfig
}

// NewClient 创建新的 Redis 客户端
func NewClient(config *RedisConfig) (*Client, error) {
	client, err := createRedisClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create redis client: %w", err)
	}

	return &Client{
		client:  client,
		config:  config,
		cache:   NewRedisCache[any](client),
		lockMgr: NewRedisLockManager(client),
	}, nil
}

// NewClientFromViper 从 Viper 配置创建客户端
func NewClientFromViper(key string) (*Client, error) {
	config := &RedisConfig{}
	if err := viper.UnmarshalKey(key, config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal redis config: %w", err)
	}

	return NewClient(config)
}

// Cache 获取缓存客户端
func (c *Client) Cache() Cache[any] {
	return c.cache
}

// LockManager 获取锁管理器
func (c *Client) LockManager() LockManager {
	return c.lockMgr
}

// SimpleLock 创建简单锁（只需要键名）
func (c *Client) SimpleLock(key string) DistributedLock {
	return c.lockMgr.NewLock(key,
		WithExpiration(30*time.Second),
		WithRetryTimes(3),
	)
}

// SimpleLockWithTimeout 创建带超时的简单锁
func (c *Client) SimpleLockWithTimeout(key string, timeout time.Duration) DistributedLock {
	return c.lockMgr.NewLock(key,
		WithExpiration(timeout),
		WithRetryTimes(3),
	)
}

// AutoLock 创建自动续期的锁
func (c *Client) AutoLock(key string) DistributedLock {
	return c.lockMgr.NewLock(key,
		WithExpiration(30*time.Second),
		WithRetryTimes(3),
		WithAutoExtend(true),
	)
}

// SimpleGet 简单获取缓存（无context，类型安全）
func (c *Client) SimpleGet(key string, valueType any) (any, error) {
	value, err := c.cache.Get(context.Background(), key)
	if err != nil {
		return nil, err
	}
	
	// 如果用户传入了类型示例，尝试类型转换
	if valueType != nil {
		// 使用反射进行类型转换
		if reflect.TypeOf(valueType).Kind() == reflect.Ptr {
			// 创建新的类型实例
			newValue := reflect.New(reflect.TypeOf(valueType).Elem()).Interface()
			if err := mapstructure.Decode(value, newValue); err == nil {
				return newValue, nil
			}
		}
	}
	
	return value, nil
}

// SimpleSet 简单设置缓存（无context）
func (c *Client) SimpleSet(key string, value any, ttl time.Duration) error {
	return c.cache.Set(context.Background(), key, &value, ttl)
}

// SimpleDelete 简单删除缓存（无context）
func (c *Client) SimpleDelete(key string) error {
	return c.cache.Delete(context.Background(), key)
}

// SimpleExists 简单检查缓存是否存在（无context）
func (c *Client) SimpleExists(key string) (bool, error) {
	return c.cache.Exists(context.Background(), key)
}

// SimpleGetOrLoad 简单的带自动加载获取（无context，类型安全）
func (c *Client) SimpleGetOrLoad(key string, valueType any, loader func(string) (any, error)) (any, error) {
	// 先尝试获取缓存
	value, err := c.cache.Get(context.Background(), key)
	if err == nil {
		// 如果缓存命中，尝试类型转换
		if valueType != nil {
			if reflect.TypeOf(valueType).Kind() == reflect.Ptr {
				newValue := reflect.New(reflect.TypeOf(valueType).Elem()).Interface()
				if err := mapstructure.Decode(value, newValue); err == nil {
					return newValue, nil
				}
			}
		}
		return value, nil
	}

	// 缓存未命中，调用加载函数
	loadedValue, err := loader(key)
	if err != nil {
		return nil, err
	}

	// 缓存加载的值
	if err := c.cache.Set(context.Background(), key, &loadedValue, 10*time.Minute); err != nil {
		// 记录错误但不影响返回
	}

	// 尝试类型转换
	if valueType != nil {
		if reflect.TypeOf(valueType).Kind() == reflect.Ptr {
			newValue := reflect.New(reflect.TypeOf(valueType).Elem()).Interface()
			if err := mapstructure.Decode(loadedValue, newValue); err == nil {
				return newValue, nil
			}
		}
	}

	return loadedValue, nil
}

// SimpleLock 简单获取锁（无context）
func (c *Client) SimpleLockNoCtx(key string) error {
	lock := c.SimpleLock(key)
	return lock.Lock(context.Background())
}

// SimpleUnlock 简单释放锁（无context）
func (c *Client) SimpleUnlockNoCtx(key string) error {
	lock := c.SimpleLock(key)
	return lock.Unlock(context.Background())
}

// RawClient 获取原始 Redis 客户端
func (c *Client) RawClient() redis.UniversalClient {
	return c.client
}

// Config 获取配置
func (c *Client) Config() *RedisConfig {
	return c.config
}

// Ping 检查连接
func (c *Client) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

// Close 关闭客户端
func (c *Client) Close() error {
	// 关闭缓存
	if err := c.cache.Close(); err != nil {
		return fmt.Errorf("failed to close cache: %w", err)
	}

	// 关闭锁管理器
	if err := c.lockMgr.Close(); err != nil {
		return fmt.Errorf("failed to close lock manager: %w", err)
	}

	// Redis 客户端由外部管理，这里不关闭
	return nil
}

// createRedisClient 创建 Redis 客户端
func createRedisClient(config *RedisConfig) (redis.UniversalClient, error) {
	if config.EnabledCluster {
		// 集群模式
		if len(config.ClusterAddrs) == 0 {
			return nil, fmt.Errorf("cluster enabled but no addresses provided")
		}

		clusterOpts := &redis.ClusterOptions{
			Addrs:     config.ClusterAddrs,
			Password:  config.Password,
			TLSConfig: getTLSConfig(config.EnabledTLS),
			PoolSize:  config.PoolSize,
		}

		return redis.NewClusterClient(clusterOpts), nil
	}

	// 单机模式
	if config.Addr == "" {
		return nil, fmt.Errorf("redis address is required")
	}

	opts := &redis.Options{
		Addr:         config.Addr,
		Password:     config.Password,
		Username:     config.Username,
		DB:           config.DB,
		TLSConfig:    getTLSConfig(config.EnabledTLS),
		PoolSize:     config.PoolSize,
		MinIdleConns: config.MinIdleConns,
		MaxRetries:   config.MaxRetries,
	}

	return redis.NewClient(opts), nil
}

// getTLSConfig 获取 TLS 配置
func getTLSConfig(enabled bool) *tls.Config {
	if !enabled {
		return nil
	}

	return &tls.Config{
		InsecureSkipVerify: true,
	}
}
