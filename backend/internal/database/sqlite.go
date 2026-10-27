package database

import (
	"path/filepath"

	"homework-station/internal/config"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func openSqlite(settings config.Config, cfg *gorm.Config) (*gorm.DB, error) {

	return gorm.Open(sqlite.Open(config.Or(settings.Database.DSN, filepath.Join(settings.DataDir, "homework.db"))+"?_busy_timeout=5000&_foreign_keys=on"), cfg)
}
