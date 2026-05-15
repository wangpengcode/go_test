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

func (c *Client) Close() error { return c.conn.Close() }
