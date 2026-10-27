package service

import (
	"strconv"
	"strings"

	"homework-station/internal/model"

	"gorm.io/gorm"
)

type SettingsService struct{ db *gorm.DB }

func NewSettingsService(db *gorm.DB) *SettingsService { return &SettingsService{db: db} }
func (s *SettingsService) Get(key string) string {
	var setting model.Setting
	s.db.First(&setting, "key = ?", key)
	return setting.Value
}
func (s *SettingsService) All() map[string]string {
	rows := []model.Setting{}
	s.db.Find(&rows)
	values := map[string]string{}
	for _, row := range rows {
		values[row.Key] = row.Value
	}
	return values
}
func (s *SettingsService) IdentityConfig() (map[string]string, error) {
	var rows []model.Setting
	if err := s.db.Where("key LIKE ?", "realname_%").Find(&rows).Error; err != nil {
		return nil, Reject(500, "读取实名认证配置失败")
	}
	values := map[string]string{}
	for _, row := range rows {
		values[row.Key] = row.Value
	}
	return values, nil
}
func (s *SettingsService) Update(m map[string]string) error {
	for k, v := range m {
		switch k {
		case "site_title":
			if strings.TrimSpace(v) == "" || len(v) > 120 {
				return Reject(400, "站点标题需为 1–120 字节")
			}
		case "default_max_size", "avatar_max_size":
			n, e := strconv.ParseInt(v, 10, 64)
			if e != nil || n <= 0 || n > 10*1024*1024*1024 {
				return Reject(400, "大小应为 1 字节至 10GB")
			}
		case "binding_secret":
			if len(v) < 32 {
				return Reject(400, "绑定码 secret 至少 32 个字符")
			}
		case "registration_enabled":
			if v != "true" && v != "false" {
				return Reject(400, "注册开关无效")
			}
		case "realname_api_url", "realname_success_field", "realname_success_value", "realname_id_field", "realname_name_field":
			if err := validateRealNameSetting(k, v); err != nil {
				return Reject(400, err.Error())
			}
		case "default_extensions":
			if len(v) > 1000 {
				return Reject(400, "后缀设置过长")
			}
		default:
			return Reject(400, "未知设置项")
		}
	}
	e := s.db.Transaction(func(tx *gorm.DB) error {
		for k, v := range m {
			if e := tx.Model(&model.Setting{}).Where("key = ?", k).Update("value", v).Error; e != nil {
				return e
			}
		}
		return nil
	})
	if e != nil {
		return Reject(500, "保存失败")
	}
	return nil
}
