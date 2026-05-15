package registry

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"
)

// RegisterTCP 使用 TCP 健康检查把服务注册到 Consul。
// 成功时返回一个 deregister 函数；当未启用或注册失败时返回 nil。
func RegisterTCP(ctx context.Context, consul *Consul, cfg Config, fallbackServiceName string) (func(), error) {
	if !cfg.Enable {
		return nil, nil
	}
	if consul == nil {
		return nil, fmt.Errorf("consul client is nil")
	}

	svcName := cfg.ServiceName
	if svcName == "" {
		svcName = fallbackServiceName
	}
	svcID := cfg.ServiceID
	if svcID == "" {
		svcID = DefaultInstanceID(svcName)
	}

	if err := consul.Register(ctx, RegisterRequest{
		ID:      svcID,
		Name:    svcName,
		Address: cfg.AdvertiseHost,
		Port:    cfg.AdvertisePort,
		Check: &Check{
			TCP:                            net.JoinHostPort(cfg.AdvertiseHost, strconv.Itoa(cfg.AdvertisePort)),
			Interval:                       "10s",
			Timeout:                        "2s",
			DeregisterCriticalServiceAfter: "30s",
		},
	}); err != nil {
		return nil, err
	}

	return func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		_ = consul.Deregister(ctx, svcID)
		cancel()
	}, nil
}

// RegisterHTTP 使用 HTTP 健康检查把服务注册到 Consul。
// 成功时返回一个 deregister 函数；当未启用或注册失败时返回 nil。
func RegisterHTTP(ctx context.Context, consul *Consul, cfg Config, fallbackServiceName, healthURL string) (func(), error) {
	if !cfg.Enable {
		return nil, nil
	}
	if consul == nil {
		return nil, fmt.Errorf("consul client is nil")
	}

	svcName := cfg.ServiceName
	if svcName == "" {
		svcName = fallbackServiceName
	}
	svcID := cfg.ServiceID
	if svcID == "" {
		svcID = DefaultInstanceID(svcName)
	}

	if err := consul.Register(ctx, RegisterRequest{
		ID:      svcID,
		Name:    svcName,
		Address: cfg.AdvertiseHost,
		Port:    cfg.AdvertisePort,
		Check: &Check{
			HTTP:                           healthURL,
			Interval:                       "10s",
			Timeout:                        "2s",
			DeregisterCriticalServiceAfter: "30s",
		},
	}); err != nil {
		return nil, err
	}

	return func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		_ = consul.Deregister(ctx, svcID)
		cancel()
	}, nil
}
