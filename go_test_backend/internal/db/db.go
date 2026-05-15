package db

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
	SSLMode  string
}

// Open 打开一个 GORM 的数据库连接。
//
// 给刚接触 Go 的同学：
//   - `(*gorm.DB, error)` 表示函数返回两个值：
//     1）gorm.DB 的指针；2）error（错误）。
//   - Go 里“错误是普通值”，常见写法是：
//     result, err := ...
//     if err != nil { return ..., err }
func Open(cfg Config, log *zap.Logger) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=Asia/Shanghai",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database, cfg.SSLMode,
	)

	gormLogger := logger.New(
		&zapGormWriter{log: log},
		logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	return gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: gormLogger})
}

type zapGormWriter struct {
	log *zap.Logger
}

// Printf 是 GORM logger.New(...) 所需要的 writer 接口方法。
// GORM 会在每一条 SQL 日志时调用它。
func (w *zapGormWriter) Printf(format string, args ...any) {
	// GORM 的 logger 会把 format+args 组合成一行日志文本。
	w.log.Info(fmt.Sprintf(format, args...))
}
