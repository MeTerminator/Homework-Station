package database

import (
	"fmt"
	"regexp"
	"strings"

	"homework-station/internal/config"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func openPostgres(settings config.Config, cfg *gorm.Config) (*gorm.DB, error) {

	dsn := settings.Database.DSN
	if dsn != "" {
		pc, err := pgx.ParseConfig(dsn)
		if err != nil {
			return nil, err
		}
		name := pc.Database
		if !regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString(name) {
			return nil, fmt.Errorf("invalid database name in DB_DSN")
		}
		pc.Database = "postgres"
		conn := stdlib.OpenDB(*pc)
		server, err := gorm.Open(postgres.New(postgres.Config{Conn: conn}), cfg)
		if err != nil {
			conn.Close()
			return nil, err
		}
		var n int64
		err = server.Raw("SELECT count(*) FROM pg_database WHERE datname = ?", name).Scan(&n).Error
		if err == nil && n == 0 {
			err = server.Exec(`CREATE DATABASE "` + name + `"`).Error
		}
		conn.Close()
		if err != nil {
			return nil, err
		}
		return gorm.Open(postgres.Open(dsn), cfg)
	}
	name := config.Or(settings.Database.Name, "homework")
	if !regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString(name) {
		return nil, fmt.Errorf("invalid DB_NAME")
	}
	// libpq DSN values are quoted to preserve passwords containing spaces.
	quote := func(s string) string {
		return "'" + strings.ReplaceAll(strings.ReplaceAll(s, `\`, `\\`), `'`, `\'`) + "'"
	}
	base := fmt.Sprintf("host=%s port=%s user=%s password=%s sslmode=%s", quote(config.Or(settings.Database.Host, "localhost")), quote(config.Or(settings.Database.Port, "5432")), quote(config.Or(settings.Database.User, "postgres")), quote(settings.Database.Password), quote(config.Or(settings.Database.SSLMode, "disable")))
	server, e := gorm.Open(postgres.Open(base+" dbname=postgres"), cfg)
	if e != nil {
		return nil, e
	}
	var n int64
	e = server.Raw("SELECT count(*) FROM pg_database WHERE datname = ?", name).Scan(&n).Error
	if e == nil && n == 0 {
		e = server.Exec(`CREATE DATABASE "` + name + `"`).Error
	}
	s, _ := server.DB()
	if s != nil {
		s.Close()
	}
	if e != nil {
		return nil, e
	}
	return gorm.Open(postgres.Open(base+" dbname="+quote(name)), cfg)
}
