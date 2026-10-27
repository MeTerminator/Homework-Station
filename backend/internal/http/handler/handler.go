package handler

import (
	"sync"

	"homework-station/internal/config"
	"homework-station/internal/service"

	"gorm.io/gorm"
)

type Handler struct {
	Users           *service.UserService
	Categories      *service.CategoryService
	Assignments     *service.AssignmentService
	Submissions     *service.SubmissionService
	OverviewService *service.OverviewService
	DB              *gorm.DB
	Config          config.Config
	SessionSecret   []byte
	Settings        *service.SettingsService
	Files           *service.FileService
	Identity        *service.IdentityService
	captchaMu       sync.Mutex
	captchas        map[string]captchaEntry
}

func New(db *gorm.DB, cfg config.Config, secret []byte) *Handler {
	files := service.NewFileService(db, cfg.DataDir)
	settings := service.NewSettingsService(db)
	return &Handler{
		DB:              db,
		Config:          cfg,
		SessionSecret:   secret,
		Settings:        settings,
		Users:           service.NewUserService(db, files, settings),
		Files:           files,
		Identity:        service.NewIdentityService(),
		Categories:      service.NewCategoryService(db, files),
		Assignments:     service.NewAssignmentService(db, files),
		Submissions:     service.NewSubmissionService(db, files),
		OverviewService: service.NewOverviewService(db),
		captchas:        map[string]captchaEntry{},
	}
}
