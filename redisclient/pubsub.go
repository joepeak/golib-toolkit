package redisclient

import (
	"context"
	"time"

	"github.com/joepeak/golib-toolkit/convert"
	"github.com/sirupsen/logrus"
)

// Publish 发布消息
func Publish(ctx context.Context, channel string, message interface{}) error {
	_, err := DefaultClient.Publish(ctx, channel, convert.ObjToJson(message)).Result()
	return err
}

// Subscribe 订阅消息（支持 Redis 断线重连）
func Subscribe(ctx context.Context, channel string, handler func(data string) error) error {

	for {
		pubsub := DefaultClient.Subscribe(ctx, channel)
		ch := pubsub.Channel()

		logrus.Info("Redis 订阅成功, channel: ", channel)

		// 监听消息
		for {
			select {
			case <-ctx.Done(): // 监听外部取消信号，安全退出
				_ = pubsub.Close() // 确保关闭 Redis 订阅连接
				logrus.Warn("Redis 订阅退出, channel: ", channel)
				return nil

			case msg, ok := <-ch: // 监听 Redis 消息
				if !ok {
					logrus.Warn("Redis 订阅通道关闭, 5秒后重连, channel: ", channel)
					_ = pubsub.Close()          // 关闭当前连接，防止遗留连接
					time.Sleep(5 * time.Second) // 休眠后重新订阅
					break                       // 退出当前 for 循环，进入重连
				}

				// 处理收到的消息
				if err := handler(msg.Payload); err != nil {
					logrus.Error("消息处理失败, channel: ", channel, ", error: ", err)
				} else {
					logrus.Debug("收到 Redis 消息, channel: ", channel, ", message: ", msg.Payload)
				}
			}
		}
	}
}
