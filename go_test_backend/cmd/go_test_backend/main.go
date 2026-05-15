package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"go.uber.org/zap"
	"go_test/go_test_backend/internal/config"
	"go_test/go_test_backend/internal/db"
	"go_test/go_test_backend/internal/db/migrate"
	"go_test/go_test_backend/internal/id"
	"go_test/go_test_backend/internal/logging"
	"go_test/go_test_backend/internal/server/grpcserver"
	"go_test/go_test_backend/internal/user"
	_ "go_test/shared/grpcjson"
	"go_test/shared/pathutil"
	"go_test/shared/registry"
	"go_test/shared/usergrpc"
	"google.golang.org/grpc"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "go_test_backend/configs/config.yaml", "config file path")
	flag.Parse()

	if p, ok := pathutil.FirstExistingFile(
		configPath,
		"configs/config.yaml",
		filepath.Join("go_test_backend", "configs", "config.yaml"),
	); ok {
		configPath = p
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		panic(err)
	}

	if root, ok := pathutil.ModuleRootFrom(configPath); ok {
		cfg.Migrations.Dir = pathutil.ResolveMaybeRelativeToRoot(cfg.Migrations.Dir, root)
	}

	log, err := logging.New(cfg.Log.Level)
	if err != nil {
		panic(err)
	}
	defer func() { _ = log.Sync() }()

	gormDB, err := db.Open(db.Config{
		Host:     cfg.DB.Host,
		Port:     cfg.DB.Port,
		User:     cfg.DB.User,
		Password: cfg.DB.Password,
		Database: cfg.DB.Database,
		SSLMode:  cfg.DB.SSLMode,
	}, log)
	if err != nil {
		log.Fatal("db open failed", zap.Error(err))
	}

	if err := migrate.ApplyDir(gormDB, cfg.Migrations.Dir, log); err != nil {
		log.Fatal("migrations failed", zap.Error(err))
	}

	idgen, err := id.NewSnowflake(cfg.ID.WorkerID)
	if err != nil {
		log.Fatal("id generator init failed", zap.Error(err))
	}

	repo := user.NewRepo(gormDB)
	svc := grpcserver.New(log, repoAdapter{repo: repo}, idgen)

	grpcSrv := grpc.NewServer(grpc.UnaryInterceptor(grpcserver.UnaryServerLogger(log)))
	usergrpc.RegisterUserServiceServer(grpcSrv, svc)

	lis, err := net.Listen("tcp", cfg.Service.GRPCAddr)
	if err != nil {
		log.Fatal("listen failed", zap.Error(err))
	}

	log.Info("backend grpc started", zap.String("addr", cfg.Service.GRPCAddr), zap.String("service", cfg.Service.Name))

	go func() {
		if err := grpcSrv.Serve(lis); err != nil {
			log.Fatal("grpc serve failed", zap.Error(err))
		}
	}()

	var deregister func()
	if cfg.Registry.Enable {
		consul := registry.NewConsul(cfg.Registry.ConsulAddr)
		svcName := cfg.Registry.ServiceName
		if svcName == "" {
			svcName = cfg.Service.Name
		}
		svcID := cfg.Registry.ServiceID
		if svcID == "" {
			svcID = registry.DefaultInstanceID(svcName)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		err := consul.Register(ctx, registry.RegisterRequest{
			ID:      svcID,
			Name:    svcName,
			Address: cfg.Registry.AdvertiseHost,
			Port:    cfg.Registry.AdvertisePort,
			Check: &registry.Check{
				TCP:                            net.JoinHostPort(cfg.Registry.AdvertiseHost, fmt.Sprintf("%d", cfg.Registry.AdvertisePort)),
				Interval:                       "10s",
				Timeout:                        "2s",
				DeregisterCriticalServiceAfter: "30s",
			},
		})
		cancel()
		if err != nil {
			log.Warn("consul register failed", zap.Error(err))
		} else {
			log.Info("consul registered", zap.String("service", svcName), zap.String("id", svcID))
			deregister = func() {
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				_ = consul.Deregister(ctx, svcID)
				cancel()
			}
		}
	}

	waitSignal()
	log.Info("shutdown")
	if deregister != nil {
		deregister()
	}
	grpcSrv.GracefulStop()
}

type repoAdapter struct {
	repo *user.Repo
}

func (a repoAdapter) Add(ctx context.Context, u grpcserver.UserModel) (grpcserver.UserModel, error) {
	out, err := a.repo.Add(ctx, user.Model{UserID: u.UserID, Name: u.Name, Status: u.Status})
	if err != nil {
		return grpcserver.UserModel{}, err
	}
	return grpcserver.UserModel{UserID: out.UserID, Name: out.Name, Status: out.Status}, nil
}

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

func waitSignal() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
}
