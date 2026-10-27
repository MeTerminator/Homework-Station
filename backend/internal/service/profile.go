package service

import (
	"encoding/json"

	"golang.org/x/crypto/bcrypt"

	"homework-station/internal/model"
	"homework-station/internal/security"
)

type ProfileInput struct {
	Username    string `json:"username"`
	CategoryIDs []uint `json:"category_ids"`
}
type PasswordInput struct {
	Old      string `json:"old"`
	Password string `json:"password"`
	Confirm  string `json:"confirm"`
}

func (svc *UserService) UpdateProfile(u model.User, v ProfileInput) (model.User, error) {
	changes := map[string]interface{}{}
	if v.Username != "" {
		if !security.ValidUsername(v.Username) {
			return model.User{}, Reject(400, "用户名格式不正确")
		}
		changes["username"] = v.Username
	}
	if v.CategoryIDs != nil {
		for _, id := range v.CategoryIDs {
			var n int64
			svc.db.Model(&model.Category{}).Where("id = ?", id).Count(&n)
			if n == 0 {
				return model.User{}, Reject(400, "分类不存在")
			}
		}
		b, _ := json.Marshal(v.CategoryIDs)
		changes["category_ids"] = string(b)
	}
	if svc.db.Model(&u).Updates(changes).Error != nil {
		return model.User{}, Reject(409, "用户名已存在")
	}
	svc.db.First(&u, u.ID)
	return u, nil
}

func (svc *UserService) ChangePassword(u model.User, v PasswordInput) (model.User, error) {
	if bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(v.Old)) != nil {
		return model.User{}, Reject(400, "旧密码不正确")
	}
	if !security.ValidPassword(v.Password) || v.Password != v.Confirm {
		return model.User{}, Reject(400, "新密码需为 6–72 字节且两次一致")
	}
	pw, _ := bcrypt.GenerateFromPassword([]byte(v.Password), bcrypt.DefaultCost)
	u.AuthVersion++
	if svc.db.Model(&u).Updates(map[string]interface{}{"password": string(pw), "auth_version": u.AuthVersion}).Error != nil {
		return model.User{}, Reject(500, "保存失败")
	}

	return u, nil
}
