package config

// DefaultSettings returns a fresh map; initialization never mutates shared state.
func DefaultSettings(bindingSecret string) map[string]string {
	return map[string]string{
		"site_title":                      "作业驿站",
		"browser_title":                   "",
		"site_logo":                       "",
		"site_icon":                       "",
		"default_extensions":              "",
		"default_max_size":                "20971520",
		"upload_cleanup_value":            "7",
		"upload_cleanup_unit":             "day",
		"avatar_max_size":                 "10485760",
		"binding_secret":                  bindingSecret,
		"qq_binding_instructions":         "私聊机器人发送 `/绑定码`，将收到的绑定码粘贴到下方。绑定码 5 分钟内有效，请勿发送给他人。",
		"registration_enabled":            "true",
		"realname_api_url":                "",
		"realname_student_no_pattern":     `^[0-9]{10}$`,
		"realname_student_no_placeholder": "10 位数字学号",
		"realname_success_field":          "success",
		"realname_success_value":          "true",
		"realname_id_field":               "data.id",
		"realname_name_field":             "data.name",
	}
}
