package database

import (
	"fmt"
	"log"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"homework-station/internal/config"
)

func Open(settings config.Config, output *log.Logger) (*gorm.DB, error) {
	options := &gorm.Config{Logger: logger.New(output, logger.Config{SlowThreshold: time.Second, LogLevel: logger.Warn, IgnoreRecordNotFoundError: true})}
	switch strings.ToLower(settings.Database.Driver) {
	case "sqlite":
		return openSqlite(settings, options)
	case "mysql":
		return openMysql(settings, options)
	case "postgres", "postgresql", "pgsql":
		return openPostgres(settings, options)
	default:
		return nil, fmt.Errorf("DB_DRIVER 只支持 sqlite/mysql/postgres")
	}
}
