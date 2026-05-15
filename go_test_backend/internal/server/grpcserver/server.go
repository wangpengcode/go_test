package grpcserver

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"go_test/shared/user"
	"go_test/shared/usergrpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type IDGen interface {
	// Next 返回一个新的 int64 ID（例如雪花算法）。
	Next() int64
}

type Repo interface {
	// Add 保存用户并返回保存后的模型。
	Add(ctx context.Context, u UserModel) (UserModel, error)
	// Query 按 userID 查询用户；found 表示是否找到。
	Query(ctx context.Context, userID string) (UserModel, bool, error)
}

type UserModel struct {
	UserID string
	Name   string
	Status string
}

type Server struct {
	log   *zap.Logger
	repo  Repo
	idgen IDGen
}

// New constructs a Server.
//
// 给刚接触 Go 的同学：
// - `Repo` 和 `IDGen` 是接口：这样你可以在测试或不同存储方式下替换实现。
// - Go 更强调“组合而不是继承”：通过参数传入依赖，而不是搞一堆父类/子类。
func New(log *zap.Logger, repo Repo, idgen IDGen) *Server {
	return &Server{log: log, repo: repo, idgen: idgen}
}

// Add 实现 gRPC 接口 UserService/Add。
func (s *Server) Add(ctx context.Context, in *user.User) (*user.User, error) {
	if in == nil {
		return nil, status.Error(codes.InvalidArgument, "empty user")
	}
	if in.Name == "" || in.Status == "" {
		return nil, status.Error(codes.InvalidArgument, "name/status required")
	}
	u := UserModel{UserID: in.UserID, Name: in.Name, Status: in.Status}
	if u.UserID == "" {
		u.UserID = fmt.Sprintf("%d", s.idgen.Next())
	}
	s.log.Info("add user", zap.String("user_id", u.UserID))
	out, err := s.repo.Add(ctx, u)
	if err != nil {
		s.log.Error("db add user failed", zap.Error(err))
		return nil, status.Error(codes.Internal, "db error")
	}
	return &user.User{UserID: out.UserID, Name: out.Name, Status: out.Status}, nil
}

// Query 实现 gRPC 接口 UserService/Query。
func (s *Server) Query(ctx context.Context, in *user.QueryRequest) (*user.User, error) {
	if in == nil || in.UserID == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id required")
	}
	s.log.Info("query user", zap.String("user_id", in.UserID))
	out, ok, err := s.repo.Query(ctx, in.UserID)
	if err != nil {
		s.log.Error("db query user failed", zap.Error(err))
		return nil, status.Error(codes.Internal, "db error")
	}
	if !ok {
		return nil, status.Error(codes.NotFound, "user not found")
	}
	return &user.User{UserID: out.UserID, Name: out.Name, Status: out.Status}, nil
}

// 编译期检查：
// 如果 *Server 没有实现 usergrpc.UserServiceServer 需要的所有方法，这行代码会在编译时报错。
// 这是 Go 项目里很常见的写法，用来防止重构后“接口实现悄悄断掉”。
var _ usergrpc.UserServiceServer = (*Server)(nil)
