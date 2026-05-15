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
	Next() int64
}

type Repo interface {
	Add(ctx context.Context, u UserModel) (UserModel, error)
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

func New(log *zap.Logger, repo Repo, idgen IDGen) *Server {
	return &Server{log: log, repo: repo, idgen: idgen}
}

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

var _ usergrpc.UserServiceServer = (*Server)(nil)
