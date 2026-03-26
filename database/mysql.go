package database

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"gorm.io/plugin/dbresolver"
)

type MysqlConf struct {
	ConfKey     string `mapstructure:"-"`
	ConnStr     string `mapstructure:"connStr"`
	MaxIdleConn int    `mapstructure:"maxIdleConn"`
	MaxOpenConn int    `mapstructure:"maxOpenConn"`
	Slave       struct {
		MaxIdleConn int      `mapstructure:"maxIdleConn"`
		MaxOpenConn int      `mapstructure:"maxOpenConn"`
		ConnStrs    []string `mapstructure:"connStrs"`
	} `mapstructure:"slave"`
}

func InitMysqlDb(conf MysqlConf) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       conf.ConnStr, // DSN data source name
		DefaultStringSize:         200,          // string 类型字段的默认长度
		SkipInitializeWithVersion: false,        // 根据版本自动配置
	}), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		// 添加Logger配置
		Logger: logger.New(
			logrus.StandardLogger(), // 使用 logrus 作为输出器（也可以用 log.New）
			logger.Config{
				SlowThreshold:             time.Second,  // 慢SQL阈值
				LogLevel:                  logger.Error, // 日志级别
				IgnoreRecordNotFoundError: true,         // 忽略ErrRecordNotFound错误
				Colorful:                  false,        // 禁用彩色打印
			},
		),
	})

	if err != nil {
		return nil, fmt.Errorf("open %s database error: %s", conf.ConfKey, err.Error())
	}

	if sqlDb, err := db.DB(); err != nil {
		return nil, fmt.Errorf("connect %s database error: %s", conf.ConfKey, err.Error())
	} else {
		sqlDb.SetMaxIdleConns(conf.MaxIdleConn)
		sqlDb.SetMaxOpenConns(conf.MaxOpenConn)
	}

	// 配置从数据库
	if conf.Slave.ConnStrs != nil {
		slaveConnStrs := conf.Slave.ConnStrs

		var slaveDbs []gorm.Dialector

		for _, slaveDbConnStr := range slaveConnStrs {
			slaveDbs = append(slaveDbs, mysql.Open(slaveDbConnStr))
		}

		if len(slaveDbs) > 0 {
			err = db.Use(
				dbresolver.Register(dbresolver.Config{
					Sources:           []gorm.Dialector{mysql.Open(conf.ConnStr)},
					Replicas:          slaveDbs,
					Policy:            dbresolver.RandomPolicy{},
					TraceResolverMode: true,
				}).
					SetMaxIdleConns(conf.Slave.MaxIdleConn).
					SetMaxOpenConns(conf.Slave.MaxOpenConn),
			)

			if err != nil {
				return nil, fmt.Errorf("connect %s database with slaves error: %s", conf.ConfKey, err.Error())
			}
		}
	}

	logrus.Info("Connect ", conf.ConfKey, " database success")
	return db, nil
}

func InitMysqlDbWithConfKey(confKey string) (*gorm.DB, error) {
	var conf MysqlConf
	if err := viper.UnmarshalKey(confKey, &conf); err != nil {
		return nil, fmt.Errorf("unmarshal %s config error: %s", confKey, err.Error())
	}

	conf.ConfKey = confKey

	return InitMysqlDb(conf)
}
