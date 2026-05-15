package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"go.uber.org/zap"
	"go_test/go_test_api/internal/config"
	"go_test/go_test_api/internal/grpcclient"
	"go_test/go_test_api/internal/httpserver"
	"go_test/go_test_api/internal/logging"
	_ "go_test/shared/grpcjson"
	"go_test/shared/pathutil"
	"go_test/shared/registry"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "go_test_api/configs/config.yaml", "config file path")
	flag.Parse()

	if p, ok := pathutil.FirstExistingFile(
		configPath,
		"configs/config.yaml",
		filepath.Join("go_test_api", "configs", "config.yaml"),
	); ok {
		configPath = p
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		panic(err)
	}

	log, err := logging.New(cfg.Log.Level)
	if err != nil {
		panic(err)
	}
	defer func() { _ = log.Sync() }()

	backendAddr := cfg.Backend.GRPCAddr
	var consul *registry.Consul
	if cfg.Discovery.Enable || cfg.Registry.Enable {
		consul = registry.NewConsul(cfg.Registry.ConsulAddr)
	}
	if cfg.Discovery.Enable {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		addr, err := consul.DiscoverOne(ctx, cfg.Discovery.BackendServiceName)
		cancel()
		if err != nil {
			log.Warn("consul discover backend failed, fallback to backend.grpc_addr", zap.Error(err))
		} else {
			backendAddr = addr
		}
	}
	if backendAddr == "" {
		log.Fatal("backend address is empty (set backend.grpc_addr or enable discovery)")
	}

	cli, err := grpcclient.Dial(backendAddr, 3*time.Second, log)
	if err != nil {
		log.Fatal("dial backend failed", zap.Error(err))
	}
	defer func() { _ = cli.Close() }()

	srv := httpserver.New(log, cli.User, time.Duration(cfg.Backend.TimeoutMS)*time.Millisecond)
	httpSrv := &http.Server{
		Addr:    cfg.Service.HTTPAddr,
		Handler: srv.Handler(cfg.Service.BasePath),
	}

	log.Info("api http started", zap.String("addr", cfg.Service.HTTPAddr), zap.String("service", cfg.Service.Name))

	var deregister func()
	if cfg.Registry.Enable {
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
				HTTP:                           "http://" + cfg.Registry.AdvertiseHost + ":" + strconv.Itoa(cfg.Registry.AdvertisePort) + cfg.Service.BasePath + "/health",
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

	go func() {
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("http serve failed", zap.Error(err))
		}
	}()

	waitSignal()
	log.Info("shutdown")
	if deregister != nil {
		deregister()
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = httpSrv.Shutdown(ctx)
}

func waitSignal() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
}
