package database

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/v2/event"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type MongoDbConf struct {
	ConfKey string `mapstructure:"-"`
	ConnStr string `mapstructure:"connStr"`
}

func InitMongoDb(conf MongoDbConf) (*mongo.Client, error) {
	clientOptions := options.Client().ApplyURI(conf.ConnStr)
	// 添加命令监控
	cmdMonitor := &event.CommandMonitor{
		Started: func(ctx context.Context, evt *event.CommandStartedEvent) {
			logrus.Debugf("mongodb command: %v", evt.Command)
		},
		Failed: func(ctx context.Context, evt *event.CommandFailedEvent) {
			logrus.Errorf("mongodb command failed: %v", evt.Failure)
		},
	}
	clientOptions.SetMonitor(cmdMonitor)

	client, err := mongo.Connect(clientOptions)
	if err != nil {
		return nil, fmt.Errorf("connect %s database error: %s", conf.ConfKey, err.Error())
	}

	// 检查连接状态
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err = client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("ping %s database error: %s", conf.ConfKey, err.Error())
	}

	return client, nil
}

func InitMongoDbWithConfKey(confKey string) (*mongo.Client, error) {
	var conf MongoDbConf
	if err := viper.UnmarshalKey(confKey, &conf); err != nil {
		return nil, fmt.Errorf("unmarshal %s config error: %s", confKey, err.Error())
	}

	conf.ConfKey = confKey

	return InitMongoDb(conf)
}
