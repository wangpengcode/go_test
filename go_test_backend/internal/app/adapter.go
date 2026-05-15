package app

import (
	"context"

	"go_test/go_test_backend/internal/server/grpcserver"
	"go_test/go_test_backend/internal/user"
)

// repoAdapter is an adapter that connects the "database repo" layer to the "gRPC server" layer.
//
// 给刚接触 Go 的同学：
// - “接口（interface）”描述的是一组方法（见 grpcserver.Repo）。
// - Go 里没有 implements 关键字：只要一个类型把接口要求的方法都实现了，它就“自动实现”该接口。
type repoAdapter struct {
	repo *user.Repo
}

// Add 实现了 grpcserver.Repo.Add。
func (a repoAdapter) Add(ctx context.Context, u grpcserver.UserModel) (grpcserver.UserModel, error) {
	out, err := a.repo.Add(ctx, user.Model{UserID: u.UserID, Name: u.Name, Status: u.Status})
	if err != nil {
		return grpcserver.UserModel{}, err
	}
	return grpcserver.UserModel{UserID: out.UserID, Name: out.Name, Status: out.Status}, nil
}

// Query 实现了 grpcserver.Repo.Query。
func (a repoAdapter) Query(ctx context.Context, userID string) (grpcserver.UserModel, bool, error) {
	out, ok, err := a.repo.Query(ctx, userID)
	if err != nil {
		return grpcserver.UserModel{}, false, err
	}
	if !ok {
		return grpcserver.UserModel{}, false, nil
	}
	return grpcserver.UserModel{UserID: out.UserID, Name: out.Name, Status: out.Status}, true, nil
}
