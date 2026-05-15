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

func (w *zapGormWriter) Printf(format string, args ...any) {
	// gorm logger already formats message + args into a single line.
	w.log.Info(fmt.Sprintf(format, args...))
}
