package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Service struct {
		Name     string `mapstructure:"name"`
		GRPCAddr string `mapstructure:"grpc_addr"`
	} `mapstructure:"service"`

	DB struct {
		Host     string `mapstructure:"host"`
		Port     int    `mapstructure:"port"`
		User     string `mapstructure:"user"`
		Password string `mapstructure:"password"`
		Database string `mapstructure:"database"`
		SSLMode  string `mapstructure:"sslmode"`
	} `mapstructure:"db"`

	Migrations struct {
		Dir string `mapstructure:"dir"`
	} `mapstructure:"migrations"`

	ID struct {
		WorkerID int64 `mapstructure:"worker_id"`
	} `mapstructure:"id"`

	Registry struct {
		Enable        bool   `mapstructure:"enable"`
		ConsulAddr    string `mapstructure:"consul_addr"`
		ServiceName   string `mapstructure:"service_name"`
		ServiceID     string `mapstructure:"service_id"`
		AdvertiseHost string `mapstructure:"advertise_host"`
		AdvertisePort int    `mapstructure:"advertise_port"`
	} `mapstructure:"registry"`

	Log struct {
		Level string `mapstructure:"level"`
	} `mapstructure:"log"`
}

func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")
	v.SetEnvPrefix("GO_TEST_BACKEND")
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
