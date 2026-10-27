package config

// DefaultSettings returns a fresh map; initialization never mutates shared state.
func DefaultSettings(bindingSecret string) map[string]string {
	return map[string]string{
		"site_title":             "作业驿站",
		"default_extensions":     "",
		"default_max_size":       "20971520",
		"avatar_max_size":        "10485760",
		"binding_secret":         bindingSecret,
		"registration_enabled":   "true",
		"realname_api_url":       "",
		"realname_success_field": "success",
		"realname_success_value": "true",
		"realname_id_field":      "data.id",
		"realname_name_field":    "data.name",
	}
}
