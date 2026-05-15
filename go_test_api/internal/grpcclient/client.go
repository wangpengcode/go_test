package grpcclient

import (
	"context"
	"time"

	"go.uber.org/zap"
	"go_test/shared/grpcjson"
	"go_test/shared/usergrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn *grpc.ClientConn
	User usergrpc.UserServiceClient
	log  *zap.Logger
}

// Dial 连接后端 gRPC，并创建强类型的客户端（UserServiceClient）。
//
// 给刚接触 Go 的同学：
// - grpc.DialContext(...) 会返回一个 *grpc.ClientConn（长连接，建议复用）。
// - 我们把它保存到 Client 里，方便后续 Close() 关闭连接。
func Dial(addr string, timeout time.Duration, log *zap.Logger) (*Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	conn, err := grpc.DialContext(
		ctx,
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.CallContentSubtype(grpcjson.Name)),
		grpc.WithUnaryInterceptor(unaryClientLogger(log)),
	)
	if err != nil {
		return nil, err
	}

	return &Client{
		conn: conn,
		User: usergrpc.NewUserServiceClient(conn),
		log:  log,
	}, nil
}

// Close 关闭底层 gRPC 连接。
func (c *Client) Close() error { return c.conn.Close() }
