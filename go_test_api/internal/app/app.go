package app

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"go.uber.org/zap"
	"go_test/go_test_api/internal/config"
	"go_test/go_test_api/internal/grpcclient"
	"go_test/go_test_api/internal/httpserver"
	"go_test/go_test_api/internal/logging"
	"go_test/shared/registry"
)

// Options 定义启动 API 服务所需的输入参数。
type Options struct {
	// ConfigPath 是 YAML 配置文件路径。
	ConfigPath string
}

// Run 启动 HTTP API 服务，并阻塞直到 ctx 结束（比如 Ctrl+C）。
func Run(ctx context.Context, opts Options) error {
	// 1) 读取配置文件（YAML -> struct）。
	cfg, err := config.Load(opts.ConfigPath)
	if err != nil {
		return err
	}

	// 1.1) 尽早创建 logger，后续流程都能用它打日志。
	log, err := logging.New(cfg.Log.Level)
	if err != nil {
		return err
	}
	defer func() { _ = log.Sync() }()

	// 2) 决定后端地址：要么直连（backend.grpc_addr），要么从 Consul 发现。
	backendAddr := cfg.Backend.GRPCAddr
	var consul *registry.Consul
	if cfg.Discovery.Enable || cfg.Registry.Enable {
		consul = registry.NewConsul(cfg.Registry.ConsulAddr)
	}
	if cfg.Discovery.Enable {
		dctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		addr, err := consul.DiscoverOne(dctx, cfg.Discovery.BackendServiceName)
		cancel()
		if err != nil {
			log.Warn("consul discover backend failed, fallback to backend.grpc_addr", zap.Error(err))
		} else {
			backendAddr = addr
		}
	}
	if backendAddr == "" {
		return fmt.Errorf("backend address is empty (set backend.grpc_addr or enable discovery)")
	}

	// 3) 创建 gRPC 客户端，连接 backend。
	cli, err := grpcclient.Dial(backendAddr, 3*time.Second, log)
	if err != nil {
		return fmt.Errorf("dial backend: %w", err)
	}
	defer func() { _ = cli.Close() }()

	// 4) 构建 HTTP 路由与处理器（内部会调用 gRPC）。
	srv := httpserver.New(log, cli.User, time.Duration(cfg.Backend.TimeoutMS)*time.Millisecond)
	httpSrv := &http.Server{
		Addr:    cfg.Service.HTTPAddr,
		Handler: srv.Handler(cfg.Service.BasePath),
	}
	log.Info("api http started", zap.String("addr", cfg.Service.HTTPAddr), zap.String("service", cfg.Service.Name))

	// 5) （可选）把 API 服务注册到 Consul。
	var deregister func()
	if cfg.Registry.Enable {
		healthURL := "http://" + cfg.Registry.AdvertiseHost + ":" + strconv.Itoa(cfg.Registry.AdvertisePort) + cfg.Service.BasePath + "/health"
		rctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		fn, err := registry.RegisterHTTP(rctx, consul, cfg.Registry, cfg.Service.Name, healthURL)
		cancel()
		if err != nil {
			log.Warn("consul register failed", zap.Error(err))
		} else if fn != nil {
			deregister = fn
		}
	}

	// 在后台 goroutine 里跑 HTTP 服务器，然后阻塞等待退出信号（ctx.Done 或 ListenAndServe 返回错误）。
	errCh := make(chan error, 1)
	go func() {
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case <-ctx.Done():
		log.Info("shutdown (ctx done)")
	case err := <-errCh:
		if err != nil {
			log.Error("http serve failed", zap.Error(err))
		}
	}

	if deregister != nil {
		deregister()
	}
	sctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = httpSrv.Shutdown(sctx)
	return nil
}
