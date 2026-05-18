package app

import (
	"context"
	"fmt"
	"net"
	"time"

	"go.uber.org/zap"
	"go_test/go_test_backend/internal/config"
	"go_test/go_test_backend/internal/db"
	"go_test/go_test_backend/internal/db/migrate"
	"go_test/go_test_backend/internal/id"
	"go_test/go_test_backend/internal/logging"
	"go_test/go_test_backend/internal/server/grpcserver"
	"go_test/go_test_backend/internal/user"
	"go_test/shared/registry"
	"go_test/shared/usergrpc"
	"google.golang.org/grpc"
)

// Options defines inputs for starting the backend service.
type Options struct {
	// ConfigPath is the path to YAML config file.
	ConfigPath string
}

// Run starts the backend gRPC service and blocks until ctx is done.
//
// 给刚接触 Go 的同学：
// - Go 的函数可以返回多个值（这里返回 error）。
// - 尽量把依赖（logger、数据库、repo 等）通过参数传进来，少用全局变量，代码更清晰、也更好测试。
func Run(ctx context.Context, opts Options) error {
	// 1) 读取配置文件（YAML -> struct）。
	cfg, err := config.Load(opts.ConfigPath)
	if err != nil {
		return err
	}

	// 1.1) 尽早创建日志组件，后续步骤都能用统一方式打日志。
	log, err := logging.New(cfg.Log.Level)
	if err != nil {
		return err
	}
	defer func() { _ = log.Sync() }()

	// 2) 连接数据库。
	gormDB, err := db.Open(db.Config{
		Host:     cfg.DB.Host,
		Port:     cfg.DB.Port,
		User:     cfg.DB.User,
		Password: cfg.DB.Password,
		Database: cfg.DB.Database,
		SSLMode:  cfg.DB.SSLMode,
	}, log)
	if err != nil {
		return fmt.Errorf("db open: %w", err)
	}

	// 3) 执行数据库迁移（把 migrations 目录里的 .sql 按版本顺序执行）。
	if err := migrate.ApplyDir(gormDB, cfg.Migrations.Dir, log); err != nil {
		return fmt.Errorf("migrations: %w", err)
	}

	// 4) 创建业务依赖（ID 生成器、Repo、gRPC 服务实现等）。
	idgen, err := id.NewSnowflake(cfg.ID.WorkerID)
	if err != nil {
		return fmt.Errorf("id generator: %w", err)
	}

	// repoAdapter 实现了 grpcserver.Repo 接口（见 adapter.go）。
	repo := user.NewRepo(gormDB)

	var blacklist grpcserver.UserBlacklistChecker
	if cfg.ConsulKV.Enable && cfg.Registry.ConsulAddr != "" && cfg.ConsulKV.UserBlacklistKey != "" {
		consul := registry.NewConsul(cfg.Registry.ConsulAddr)
		ttl := time.Duration(cfg.ConsulKV.CacheTTLSeconds) * time.Second
		blacklist = grpcserver.NewConsulUserBlacklist(consul, cfg.ConsulKV.UserBlacklistKey, ttl)
	}

	svc := grpcserver.New(log, repoAdapter{repo: repo}, idgen, blacklist)

	// 5) 启动 gRPC 服务端。
	grpcSrv := grpc.NewServer(grpc.UnaryInterceptor(grpcserver.UnaryServerLogger(log)))
	usergrpc.RegisterUserServiceServer(grpcSrv, svc)

	lis, err := net.Listen("tcp", cfg.Service.GRPCAddr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", cfg.Service.GRPCAddr, err)
	}
	log.Info("backend grpc started", zap.String("addr", cfg.Service.GRPCAddr), zap.String("service", cfg.Service.Name))

	// 6) （可选）注册到 Consul，供其它服务发现。
	var deregister func()
	if cfg.Registry.Enable {
		consul := registry.NewConsul(cfg.Registry.ConsulAddr)
		rctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		fn, err := registry.RegisterTCP(rctx, consul, cfg.Registry, cfg.Service.Name)
		cancel()
		if err != nil {
			log.Warn("consul register failed", zap.Error(err))
		} else if fn != nil {
			log.Info("consul registered", zap.String("service", cfg.Registry.ServiceName), zap.String("id", cfg.Registry.ServiceID))
			deregister = fn
		}
	}

	// 在后台 goroutine 里跑 gRPC 服务器，然后阻塞等待退出信号（ctx.Done 或 Serve 返回错误）。
	errCh := make(chan error, 1)
	go func() { errCh <- grpcSrv.Serve(lis) }()

	select {
	case <-ctx.Done():
		log.Info("shutdown (ctx done)")
	case err := <-errCh:
		if err != nil {
			log.Error("grpc serve failed", zap.Error(err))
		}
	}

	if deregister != nil {
		deregister()
	}
	grpcSrv.GracefulStop()
	return nil
}
