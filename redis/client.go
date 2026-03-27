package redis

import (
	"context"
	"crypto/tls"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

// Client Redis 客户端包装器
type Client struct {
	cache     Cache[any]
	lockMgr   LockManager
	client    redis.UniversalClient
	config    *RedisConfig
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
		Password:      config.Password,
		Username:      config.Username,
		DB:            config.DB,
		TLSConfig:     getTLSConfig(config.EnabledTLS),
		PoolSize:      config.PoolSize,
		MinIdleConns:  config.MinIdleConns,
		MaxRetries:    config.MaxRetries,
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
