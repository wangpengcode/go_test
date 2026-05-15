package usergrpc

import (
	"context"

	"go_test/shared/user"
	"google.golang.org/grpc"
)

const (
	ServiceName = "user.UserService"
)

// UserServiceServer 是服务端接口：后端（go_test_backend）必须实现它。
type UserServiceServer interface {
	Add(context.Context, *user.User) (*user.User, error)
	Query(context.Context, *user.QueryRequest) (*user.User, error)
}

// RegisterUserServiceServer 把一个 UserServiceServer 的实现注册到 gRPC server 里。
func RegisterUserServiceServer(s grpc.ServiceRegistrar, srv UserServiceServer) {
	s.RegisterService(&grpc.ServiceDesc{
		ServiceName: ServiceName,
		HandlerType: (*UserServiceServer)(nil),
		Methods: []grpc.MethodDesc{
			{
				MethodName: "Add",
				Handler:    addHandler(srv),
			},
			{
				MethodName: "Query",
				Handler:    queryHandler(srv),
			},
		},
		Streams:  []grpc.StreamDesc{},
		Metadata: "user_service.json",
	}, srv)
}

// addHandler 把接口方法 Add “适配”为 gRPC 方法处理器。
func addHandler(srv UserServiceServer) grpc.MethodHandler {
	return func(_ any, ctx context.Context, dec func(any) error, _ grpc.UnaryServerInterceptor) (any, error) {
		var in user.User
		if err := dec(&in); err != nil {
			return nil, err
		}
		return srv.Add(ctx, &in)
	}
}

// queryHandler 把接口方法 Query “适配”为 gRPC 方法处理器。
func queryHandler(srv UserServiceServer) grpc.MethodHandler {
	return func(_ any, ctx context.Context, dec func(any) error, _ grpc.UnaryServerInterceptor) (any, error) {
		var in user.QueryRequest
		if err := dec(&in); err != nil {
			return nil, err
		}
		return srv.Query(ctx, &in)
	}
}

// UserServiceClient 是客户端接口：go_test_api 用它来调用后端。
type UserServiceClient interface {
	Add(ctx context.Context, in *user.User, opts ...grpc.CallOption) (*user.User, error)
	Query(ctx context.Context, in *user.QueryRequest, opts ...grpc.CallOption) (*user.User, error)
}

type userServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewUserServiceClient(cc grpc.ClientConnInterface) UserServiceClient {
	return &userServiceClient{cc: cc}
}

// Add 调用后端的 Add 接口。
func (c *userServiceClient) Add(ctx context.Context, in *user.User, opts ...grpc.CallOption) (*user.User, error) {
	out := new(user.User)
	err := c.cc.Invoke(ctx, "/"+ServiceName+"/Add", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Query 调用后端的 Query 接口。
func (c *userServiceClient) Query(ctx context.Context, in *user.QueryRequest, opts ...grpc.CallOption) (*user.User, error) {
	out := new(user.User)
	err := c.cc.Invoke(ctx, "/"+ServiceName+"/Query", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}
