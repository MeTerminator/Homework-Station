package service

import (
	"homework-station/internal/model"
)

func (svc *UserService) Unbind(id uint, kind string) error {
	changes := map[string]interface{}{}
	switch kind {
	case "real":
		changes["student_no"] = nil
		changes["real_name"] = ""
	case "qq":
		changes["qq"] = nil
	default:
		return Reject(400, "绑定类型不正确")
	}
	r := svc.db.Model(&model.User{}).Where("id = ?", id).Updates(changes)
	if r.Error != nil || r.RowsAffected == 0 {
		return Reject(404, "用户不存在")
	}
	return nil
}
