package service

import (
	"gorm.io/gorm"

	"homework-station/internal/model"
	"homework-station/internal/security"
)

type AdminUserInput struct {
	Username string `json:"username"`
	Admin    bool   `json:"admin"`
	Password string `json:"password"`
}
type UserService struct {
	db       *gorm.DB
	files    *FileService
	settings *SettingsService
}

func NewUserService(db *gorm.DB, files *FileService, settings *SettingsService) *UserService {
	return &UserService{db: db, files: files, settings: settings}
}
func (svc *UserService) UpdateAdmin(id uint, v AdminUserInput) (model.User, error) {
	svc.files.MutationMu.Lock()
	defer svc.files.MutationMu.Unlock()
	var u model.User
	if svc.db.First(&u, id).Error != nil {
		return model.User{}, Reject(404, "用户不存在")
	}
	if !security.ValidUsername(v.Username) {
		return model.User{}, Reject(400, "用户名格式不正确")
	}
	if u.Admin && !v.Admin {
		var n int64
		svc.db.Model(&model.User{}).Where("admin = ?", true).Count(&n)
		if n <= 1 {
			return model.User{}, Reject(409, "必须保留至少一个管理员")
		}
	}
	changes := map[string]interface{}{"username": v.Username, "admin": v.Admin}
	if u.Admin != v.Admin {
		changes["auth_version"] = u.AuthVersion + 1
	}
	if v.Password != "" {
		if !security.ValidPassword(v.Password) {
			return model.User{}, Reject(400, "密码需为 6–72 字节")
		}
		pw := security.HashPassword(v.Password)
		changes["password"] = pw
		changes["auth_version"] = u.AuthVersion + 1
	}
	if svc.db.Model(&u).Updates(changes).Error != nil {
		return model.User{}, Reject(409, "保存失败，用户名可能已存在")
	}
	svc.db.First(&u, u.ID)

	return u, nil
}
