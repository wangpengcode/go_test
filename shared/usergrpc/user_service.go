package usergrpc

import (
	"context"

	"go_test/shared/user"
	"google.golang.org/grpc"
)

const (
	ServiceName = "user.UserService"
)

type UserServiceServer interface {
	Add(context.Context, *user.User) (*user.User, error)
	Query(context.Context, *user.QueryRequest) (*user.User, error)
}

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

func addHandler(srv UserServiceServer) grpc.MethodHandler {
	return func(_ any, ctx context.Context, dec func(any) error, _ grpc.UnaryServerInterceptor) (any, error) {
		var in user.User
		if err := dec(&in); err != nil {
			return nil, err
		}
		return srv.Add(ctx, &in)
	}
}

func queryHandler(srv UserServiceServer) grpc.MethodHandler {
	return func(_ any, ctx context.Context, dec func(any) error, _ grpc.UnaryServerInterceptor) (any, error) {
		var in user.QueryRequest
		if err := dec(&in); err != nil {
			return nil, err
		}
		return srv.Query(ctx, &in)
	}
}

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

func (c *userServiceClient) Add(ctx context.Context, in *user.User, opts ...grpc.CallOption) (*user.User, error) {
	out := new(user.User)
	err := c.cc.Invoke(ctx, "/"+ServiceName+"/Add", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userServiceClient) Query(ctx context.Context, in *user.QueryRequest, opts ...grpc.CallOption) (*user.User, error) {
	out := new(user.User)
	err := c.cc.Invoke(ctx, "/"+ServiceName+"/Query", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}
