package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
	"go_test/shared/registry"
)

type Config struct {
	// Service 保存 HTTP 服务配置。
	Service ServiceConfig `mapstructure:"service"`

	// Backend 保存如何连接后端 gRPC 的配置。
	Backend BackendConfig `mapstructure:"backend"`

	// Registry 保存 Consul 注册配置。
	Registry registry.Config `mapstructure:"registry"`

	// Discovery 控制是否从 Consul 发现后端地址。
	Discovery DiscoveryConfig `mapstructure:"discovery"`

	// Log 保存日志等级等配置。
	Log LogConfig `mapstructure:"log"`
}

// ServiceConfig 对应配置文件里的 service 段（API 服务）。
type ServiceConfig struct {
	Name     string `mapstructure:"name"`
	HTTPAddr string `mapstructure:"http_addr"`
	BasePath string `mapstructure:"base_path"`
}

// BackendConfig 对应配置文件里的 backend 段。
type BackendConfig struct {
	GRPCAddr  string `mapstructure:"grpc_addr"`
	TimeoutMS int    `mapstructure:"timeout_ms"`
}

// DiscoveryConfig 控制是否启用 Consul 服务发现（寻找后端）。
type DiscoveryConfig struct {
	Enable             bool   `mapstructure:"enable"`
	BackendServiceName string `mapstructure:"backend_service_name"`
}

// LogConfig 控制日志输出（比如 level）。
type LogConfig struct {
	Level string `mapstructure:"level"`
}

func Load(path string) (*Config, error) {
	// viper 会读取 YAML，并允许用环境变量覆盖配置值。
	// 例如：GO_TEST_API_SERVICE_HTTP_ADDR=":8081"。
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")
	v.SetEnvPrefix("GO_TEST_API")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var c Config
	if err := v.Unmarshal(&c); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	return &c, nil
}
