package redisclient

import (
	"context"
	"crypto/tls"
	"strings"
	"time"

	_ "github.com/joepeak/golib-conf"

	redis "github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/go-redsync/redsync/v4"
	redsyncredis "github.com/go-redsync/redsync/v4/redis"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
)

var (
	DefaultClient redis.UniversalClient

	redisSync *redsync.Redsync
	redisPool redsyncredis.Pool
)

// LockOptions 定义分布式锁的配置选项
type LockOptions struct {
	Expiration int32         // 锁的过期时间（秒）
	MaxRetries int           // 最大重试次数
	RetryDelay time.Duration // 重试间隔
}

// DefaultLockOptions 提供锁配置的默认值
var DefaultLockOptions = LockOptions{
	Expiration: 3,                      // 默认过期时间 3 秒
	MaxRetries: 5,                      // 默认最大重试 5 次
	RetryDelay: 200 * time.Millisecond, // 默认重试间隔 200 毫秒
}

func init() {

	addr := viper.GetString("redis.addr")
	username := viper.GetString("redis.username")
	password := viper.GetString("redis.password")
	db := viper.GetInt("redis.db")
	enabledTls := viper.GetBool("redis.enabledTls")

	if viper.GetBool("redis.enabledCluster") {
		addrs := viper.GetStringSlice("redis.cluster.addrs")
		opts := &redis.ClusterOptions{
			Addrs:    addrs,
			Password: password,
		}

		if username != "" {
			opts.Username = username
		}

		if enabledTls {
			opts.TLSConfig = &tls.Config{
				InsecureSkipVerify: true,
			}
		}

		DefaultClient = redis.NewClusterClient(opts)

		if err := DefaultClient.Ping(context.TODO()).Err(); err != nil {
			logrus.Fatal("Connect redis cluster failure, addr: " + strings.Join(addrs, ",") + ", error: " + err.Error())
		}

		logrus.Info("Connect redis cluster success, addrs: ", strings.Join(addrs, ","))
	} else {
		opts := &redis.Options{
			Addr:     addr,
			Password: password,
			DB:       db,
		}

		if username != "" {
			opts.Username = username
		}

		if enabledTls {
			opts.TLSConfig = &tls.Config{
				InsecureSkipVerify: true,
			}
		}

		DefaultClient = redis.NewClient(opts)

		if err := DefaultClient.Ping(context.Background()).Err(); err != nil {
			logrus.Fatal("Connect redis failure, addr: " + addr + ", error: " + err.Error())
		}

		logrus.Info("Connect redis success, addr: ", addr)
	}

	// 初始化 redsync
	redisPool = goredis.NewPool(DefaultClient)
	redisSync = redsync.New(redisPool)
}
