package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Service struct {
		Name     string `mapstructure:"name"`
		HTTPAddr string `mapstructure:"http_addr"`
		BasePath string `mapstructure:"base_path"`
	} `mapstructure:"service"`

	Backend struct {
		GRPCAddr  string `mapstructure:"grpc_addr"`
		TimeoutMS int    `mapstructure:"timeout_ms"`
	} `mapstructure:"backend"`

	Registry struct {
		Enable        bool   `mapstructure:"enable"`
		ConsulAddr    string `mapstructure:"consul_addr"`
		ServiceName   string `mapstructure:"service_name"`
		ServiceID     string `mapstructure:"service_id"`
		AdvertiseHost string `mapstructure:"advertise_host"`
		AdvertisePort int    `mapstructure:"advertise_port"`
	} `mapstructure:"registry"`

	Discovery struct {
		Enable             bool   `mapstructure:"enable"`
		BackendServiceName string `mapstructure:"backend_service_name"`
	} `mapstructure:"discovery"`

	Log struct {
		Level string `mapstructure:"level"`
	} `mapstructure:"log"`
}

func Load(path string) (*Config, error) {
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
