package database

import (
	"log"

	"homework-station/internal/config"
	"homework-station/internal/model"
	"homework-station/internal/security"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func Initialize(db *gorm.DB, bindingSecret string, output *log.Logger) error {
	var e error
	if e = db.AutoMigrate(
		&model.User{}, &model.Category{}, &model.Assignment{},
		&model.Attachment{}, &model.Blob{}, &model.Submission{},
		&model.SubmissionAttachment{}, &model.AssignmentAttachment{},
		&model.Setting{}, &model.UsedCode{},
	); e != nil {
		return e
	}
	var count int64
	if e = db.Model(&model.User{}).Count(&count).Error; e != nil {
		return e
	}
	if count == 0 {
		pw, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
		if e = db.Create(&model.User{Username: "admin", Password: string(pw), Admin: true, CategoryIDs: "[]"}).Error; e != nil {
			return e
		}
		output.Println("数据库初始化完成：管理员 admin，初始密码 123456。请登录后立即修改密码。")
	}
	defaults := config.DefaultSettings(config.Or(bindingSecret, security.RandomHex(32)))
	for k, v := range defaults {
		var s model.Setting
		if e := db.Where("key = ?", k).Attrs(model.Setting{Value: v}).FirstOrCreate(&s, model.Setting{Key: k}).Error; e != nil {
			return e
		}
	}
	return nil
}
