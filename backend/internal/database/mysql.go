package database

import (
	"fmt"
	"regexp"

	"homework-station/internal/config"

	mysqlconfig "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func openMysql(settings config.Config, cfg *gorm.Config) (*gorm.DB, error) {

	dsn := settings.Database.DSN
	if dsn != "" {
		mc, err := mysqlconfig.ParseDSN(dsn)
		if err != nil {
			return nil, err
		}
		name := mc.DBName
		if !regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString(name) {
			return nil, fmt.Errorf("invalid database name in DB_DSN")
		}
		mc.DBName = ""
		server, err := gorm.Open(mysql.Open(mc.FormatDSN()), cfg)
		if err != nil {
			return nil, err
		}
		err = server.Exec("CREATE DATABASE IF NOT EXISTS `" + name + "` CHARACTER SET utf8mb4").Error
		sqlDB, _ := server.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
		if err != nil {
			return nil, err
		}
		return gorm.Open(mysql.Open(dsn), cfg)
	}
	name := config.Or(settings.Database.Name, "homework")
	if !regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString(name) {
		return nil, fmt.Errorf("invalid DB_NAME")
	}
	base := fmt.Sprintf("%s:%s@tcp(%s:%s)/", config.Or(settings.Database.User, "root"), settings.Database.Password, config.Or(settings.Database.Host, "localhost"), config.Or(settings.Database.Port, "3306"))
	server, e := gorm.Open(mysql.Open(base+"?charset=utf8mb4&parseTime=True&loc=UTC"), cfg)
	if e != nil {
		return nil, e
	}
	e = server.Exec("CREATE DATABASE IF NOT EXISTS `" + name + "` CHARACTER SET utf8mb4").Error
	s, _ := server.DB()
	if s != nil {
		s.Close()
	}
	if e != nil {
		return nil, e
	}
	return gorm.Open(mysql.Open(base+name+"?charset=utf8mb4&parseTime=True&loc=UTC"), cfg)
}
