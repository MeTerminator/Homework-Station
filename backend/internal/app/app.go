package app

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"

	"homework-station/internal/config"
	"homework-station/internal/database"
	"homework-station/internal/http/handler"
	"homework-station/internal/http/router"
	"homework-station/internal/security"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type App struct {
	Config     config.Config
	DB         *gorm.DB
	Router     *gin.Engine
	Handler    *handler.Handler
	logFile    *os.File
	closeOnce  sync.Once
	closeError error
}

func New(cfg config.Config) (*App, error) {
	for _, path := range []string{cfg.DataDir, filepath.Join(cfg.DataDir, "blobs"), filepath.Join(cfg.DataDir, "chunks")} {
		if err := os.MkdirAll(path, 0700); err != nil {
			return nil, err
		}
	}
	logFile, err := os.OpenFile(filepath.Join(cfg.DataDir, "app.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return nil, err
	}
	output := log.New(io.MultiWriter(os.Stdout, logFile), "", log.LstdFlags)
	secret := []byte(cfg.SessionSecret)
	if len(secret) == 0 {
		path := filepath.Join(cfg.DataDir, ".session-secret")
		secret, err = os.ReadFile(path)
		if os.IsNotExist(err) {
			secret = []byte(security.RandomHex(32))
			err = os.WriteFile(path, secret, 0600)
		}
		if err != nil {
			logFile.Close()
			return nil, err
		}
	}
	if len(secret) < 32 {
		logFile.Close()
		return nil, fmt.Errorf("JWT_SECRET 至少需要 32 个字符")
	}
	db, err := database.Open(cfg, output)
	if err != nil {
		logFile.Close()
		return nil, err
	}
	application := &App{Config: cfg, DB: db, logFile: logFile}
	sqlDB, err := db.DB()
	if err != nil {
		application.Close()
		return nil, err
	}
	sqlDB.SetMaxOpenConns(1)
	if err = database.Initialize(db, cfg.BindingSecret, output); err != nil {
		application.Close()
		return nil, err
	}
	application.Handler = handler.New(db, cfg, secret)
	application.Router, err = router.New(application.Handler, output, cfg)
	if err != nil {
		application.Close()
		return nil, err
	}
	return application, nil
}
func (a *App) Close() error {
	a.closeOnce.Do(func() {
		sqlDB, err := a.DB.DB()
		if err == nil {
			err = sqlDB.Close()
		}
		if a.logFile != nil {
			if logErr := a.logFile.Close(); err == nil {
				err = logErr
			}
		}
		a.closeError = err
	})
	return a.closeError
}
