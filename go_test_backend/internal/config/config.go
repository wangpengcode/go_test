package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
	"go_test/shared/pathutil"
	"go_test/shared/registry"
)

type Config struct {
	// Service 保存服务基础信息与监听地址。
	Service ServiceConfig `mapstructure:"service"`

	// DB 保存数据库连接配置。
	DB DBConfig `mapstructure:"db"`

	// Migrations 保存数据库迁移文件所在目录。
	Migrations MigrationsConfig `mapstructure:"migrations"`

	// ID 控制 ID 生成器行为（比如雪花算法 worker id）。
	ID IDConfig `mapstructure:"id"`

	// Registry 保存 Consul 注册相关配置。
	Registry registry.Config `mapstructure:"registry"`

	// ConsulKV 保存从 Consul KV 读取配置的相关参数。
	ConsulKV ConsulKVConfig `mapstructure:"consul_kv"`

	// Log 保存日志等级等配置。
	Log LogConfig `mapstructure:"log"`
}

type ConsulKVConfig struct {
	// Enable 控制是否启用 Consul KV 配置读取。
	Enable bool `mapstructure:"enable"`

	// UserBlacklistKey 是 Consul KV 里保存用户黑名单的 key。
	// 支持示例：
	// - "go_test/user/blacklist"
	// - "/go_test/user/blacklist"（前导 / 会被忽略）
	UserBlacklistKey string `mapstructure:"user_blacklist_key"`

	// CacheTTLSeconds 是缓存黑名单的秒数，避免每次 Query 都打到 Consul。
	CacheTTLSeconds int `mapstructure:"cache_ttl_seconds"`
}

// ServiceConfig 对应配置文件里的 service 段（后端服务）。
type ServiceConfig struct {
	Name     string `mapstructure:"name"`
	GRPCAddr string `mapstructure:"grpc_addr"`
}

// DBConfig 对应配置文件里的 db 段。
type DBConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
	SSLMode  string `mapstructure:"sslmode"`
}

// MigrationsConfig 控制迁移文件目录等信息。
type MigrationsConfig struct {
	Dir string `mapstructure:"dir"`
}

// IDConfig 配置 ID 生成器。
type IDConfig struct {
	WorkerID int64 `mapstructure:"worker_id"`
}

// LogConfig 控制日志输出（比如 level）。
type LogConfig struct {
	Level string `mapstructure:"level"`
}

func Load(path string) (*Config, error) {
	// 给刚接触 Go 的同学：
	// - viper 是常用的配置库。
	// - 我们通过结构体字段上的 tag（例如 `mapstructure:"..."`）把 YAML 映射到结构体里。
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

	// 为了避免“在不同目录启动导致相对路径失效”：
	// 我们把 migrations.dir 这种相对路径，按模块根目录（包含 go.mod 的目录）来解析。
	if root, ok := pathutil.ModuleRootFrom(path); ok {
		c.Migrations.Dir = pathutil.ResolveMaybeRelativeToRoot(c.Migrations.Dir, root)
	}
	return &c, nil
}
