package redisclient

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// 获取分布式锁
// 获取分布式锁
func Lock(lockKey string) (*redsync.Mutex, error) {
	opts := LockOptions{
		Expiration: viper.GetInt32("redis.lock.expiration"),
		MaxRetries: viper.GetInt("redis.lock.retryTimes"),
		RetryDelay: viper.GetDuration("redis.lock.retryInterval") * time.Millisecond,
	}

	// 默认值
	if opts.Expiration <= 0 {
		opts.Expiration = DefaultLockOptions.Expiration
	}
	if opts.MaxRetries <= 0 {
		opts.MaxRetries = DefaultLockOptions.MaxRetries
	}
	if opts.RetryDelay <= 0 {
		opts.RetryDelay = DefaultLockOptions.RetryDelay
	}

	return LockWithOptions(lockKey, opts)
}

// LockWithOptions 使用指定的选项获取分布式锁
// 返回锁定的 mutex，失败时返回错误
// 使用完后需调用 Unlock() 释放锁
// 获取锁
func LockWithOptions(lockKey string, opts LockOptions) (*redsync.Mutex, error) {
	if opts.Expiration <= 0 || opts.MaxRetries < 1 || opts.RetryDelay < 0 {
		return nil, errors.New("invalid lock options")
	}

	mutex := redisSync.NewMutex(
		lockKey,
		redsync.WithExpiry(time.Duration(opts.Expiration)*time.Second),
		redsync.WithTries(opts.MaxRetries),
		redsync.WithRetryDelay(opts.RetryDelay),
	)

	if err := mutex.Lock(); err != nil {
		return nil, fmt.Errorf("failed to acquire lock: %w", err)
	}

	return mutex, nil
}

// 解锁
func UnlockSafe(mutex *redsync.Mutex) error {
	if _, err := mutex.Unlock(); err != nil {
		if errors.Is(err, redsync.ErrLockAlreadyExpired) {
			logrus.Debug("Lock already expired, ignore error")
			return nil
		}
		return err
	}
	return nil
}
