package rpc

import (
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// ConnectServer 建立 gRPC 连接，支持自动重连
func ConnectServer(serverName string) *grpc.ClientConn {
	grpcAddr := viper.GetString("rpc." + serverName + ".addr")
	enabledTls := viper.GetBool("rpc." + serverName + ".enabledTls")

	var opts []grpc.DialOption

	// 配置 TLS 连接
	if enabledTls {
		pemFile := viper.GetString("rpc." + serverName + ".pemFile")
		creds, err := credentials.NewClientTLSFromFile(pemFile, "")
		if err != nil {
			logrus.Fatalf("Failed to load TLS credentials: %v", err)
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	// 设置 Keepalive 以保持长连接
	opts = append(opts, grpc.WithKeepaliveParams(keepalive.ClientParameters{
		Time:                10 * time.Second, // 每 10 秒发送 keepalive ping
		Timeout:             3 * time.Second,  // 服务器未响应的超时时间
		PermitWithoutStream: true,             // 允许在无活动 RPC 时发送 keepalive
	}))

	// gRPC 连接管理
	// 正确的 gRPC 连接配置
	opts = append(opts, grpc.WithConnectParams(grpc.ConnectParams{
		MinConnectTimeout: 5 * time.Second, // 最小连接超时
		Backoff: backoff.Config{
			BaseDelay:  1.0 * time.Second, // 基础延迟
			Multiplier: 1.6,               // 乘数
			Jitter:     0.2,               // 抖动
			MaxDelay:   120 * time.Second, // 最大延迟
		},
	}))

	// 创建 gRPC 连接
	conn, err := grpc.NewClient(grpcAddr, opts...)
	if err != nil {
		logrus.Fatalf("Failed to connect to %s server at %s: %v", serverName, grpcAddr, err)
	}

	logrus.Infof("Connected to %s server at %s", serverName, grpcAddr)

	// 启动连接状态监控（自动重连）
	go monitorConnection(serverName, grpcAddr, conn)

	return conn
}

// 监控 gRPC 连接状态，若断开则尝试重新连接
func monitorConnection(serverName, grpcAddr string, conn *grpc.ClientConn) {
	for {
		state := conn.GetState()
		logrus.Debugf("gRPC connection to %s(%s): %s", serverName, grpcAddr, state.String())

		// 若连接已断开，则等待重连
		if state == connectivity.TransientFailure || state == connectivity.Shutdown {
			logrus.Warnf("gRPC connection lost: %s(%s). Attempting to reconnect...", serverName, grpcAddr)
			conn.ResetConnectBackoff() // 触发 gRPC 内部的重连机制
		}

		time.Sleep(5 * time.Second) // 每 5 秒检查一次状态
	}
}
