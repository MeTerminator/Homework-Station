package service

import (
	"net/url"
	"strconv"
	"strings"
	"unicode/utf8"

	"homework-station/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SettingsService struct {
	db                   *gorm.DB
	uploadCleanupChanged chan struct{}
}

func NewSettingsService(db *gorm.DB) *SettingsService {
	return &SettingsService{db: db, uploadCleanupChanged: make(chan struct{}, 1)}
}
func (s *SettingsService) Get(key string) string {
	var setting model.Setting
	s.db.First(&setting, clause.Eq{Column: "key", Value: key})
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
	if err := s.db.Where(clause.Like{Column: "key", Value: "realname_%"}).Find(&rows).Error; err != nil {
		return nil, Reject(500, "读取实名认证配置失败")
	}
	values := map[string]string{}
	for _, row := range rows {
		values[row.Key] = row.Value
	}
	return values, nil
}
func (s *SettingsService) Update(m map[string]string) error {
	_, valueChanged := m["upload_cleanup_value"]
	_, unitChanged := m["upload_cleanup_unit"]
	if valueChanged || unitChanged {
		values := map[string]string{
			"upload_cleanup_value": s.Get("upload_cleanup_value"),
			"upload_cleanup_unit":  s.Get("upload_cleanup_unit"),
		}
		for _, key := range []string{"upload_cleanup_value", "upload_cleanup_unit"} {
			if value, ok := m[key]; ok {
				values[key] = value
			}
		}
		if _, err := uploadRetention(values["upload_cleanup_value"], values["upload_cleanup_unit"]); err != nil {
			return Reject(400, err.Error())
		}
	}
	for k, v := range m {
		switch k {
		case "site_title":
			m[k] = strings.TrimSpace(v)
			if m[k] == "" || utf8.RuneCountInString(m[k]) > 120 {
				return Reject(400, "站点标题需为 1–120 个字符")
			}
		case "browser_title":
			m[k] = strings.TrimSpace(v)
			if utf8.RuneCountInString(m[k]) > 120 {
				return Reject(400, "网站标题最多 120 个字符")
			}
		case "site_logo", "site_icon":
			v = strings.TrimSpace(v)
			m[k] = v
			if v != "" {
				u, err := url.Parse(v)
				local := strings.HasPrefix(v, "/") && !strings.HasPrefix(v, "//") && !strings.Contains(v, "\\")
				remote := err == nil && (u.Scheme == "https" || u.Scheme == "http") && u.Host != "" && u.User == nil
				if len(v) > 2048 || err != nil || (!local && !remote) {
					return Reject(400, "图片地址需为站内绝对路径或 HTTP(S) 链接")
				}
			}
		case "default_max_size", "avatar_max_size":
			n, e := strconv.ParseInt(v, 10, 64)
			if e != nil || n <= 0 || n > 10*1024*1024*1024 {
				return Reject(400, "大小应为 1 字节至 10GB")
			}
		case "qq_binding_instructions":
			m[k] = strings.TrimSpace(v)
			if m[k] == "" || utf8.RuneCountInString(m[k]) > 2000 {
				return Reject(400, "QQ 绑定说明需为 1–2000 个字符")
			}
		case "binding_secret":
			if len(v) < 32 {
				return Reject(400, "绑定码 secret 至少 32 个字符")
			}
		case "registration_enabled":
			if v != "true" && v != "false" {
				return Reject(400, "注册开关无效")
			}
		case "realname_student_no_pattern", "realname_student_no_placeholder", "realname_api_url", "realname_success_field", "realname_success_value", "realname_id_field", "realname_name_field":
			if err := validateRealNameSetting(k, v); err != nil {
				return Reject(400, err.Error())
			}
		case "upload_cleanup_value", "upload_cleanup_unit":
			// The duration and unit were validated together above.
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
			if e := tx.Model(&model.Setting{}).Where(clause.Eq{Column: "key", Value: k}).Update("value", v).Error; e != nil {
				return e
			}
		}
		return nil
	})
	if e != nil {
		return Reject(500, "保存失败")
	}
	if valueChanged || unitChanged {
		select {
		case s.uploadCleanupChanged <- struct{}{}:
		default:
		}
	}
	return nil
}
