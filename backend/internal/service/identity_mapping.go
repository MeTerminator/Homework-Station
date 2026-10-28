package service

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

func jsonField(value any, path string) (any, bool) {
	if path == "$" {
		return value, true
	}
	path = strings.TrimPrefix(path, "$.")
	path = strings.ReplaceAll(strings.ReplaceAll(path, "[", "."), "]", "")
	for _, key := range strings.Split(path, ".") {
		switch node := value.(type) {
		case map[string]any:
			var ok bool
			value, ok = node[key]
			if !ok {
				return nil, false
			}
		case []any:
			i, err := strconv.Atoi(key)
			if err != nil || i < 0 || i >= len(node) {
				return nil, false
			}
			value = node[i]
		default:
			return nil, false
		}
	}
	return value, true
}

func validJSONPath(path string) bool {
	if path == "$" {
		return true
	}
	if path == "" || len(path) > 512 || strings.ContainsAny(path, " \r\n\t") {
		return false
	}
	path = strings.TrimPrefix(path, "$.")
	path = strings.ReplaceAll(strings.ReplaceAll(path, "[", "."), "]", "")
	for _, key := range strings.Split(path, ".") {
		if key == "" {
			return false
		}
	}
	return true
}

func realNameURL(template, id, name string) string {
	parts := strings.SplitN(template, "?", 2)
	parts[0] = strings.NewReplacer("{id}", url.PathEscape(id), "{name}", url.PathEscape(name)).Replace(parts[0])
	if len(parts) == 2 {
		parts[1] = strings.NewReplacer("{id}", url.QueryEscape(id), "{name}", url.QueryEscape(name)).Replace(parts[1])
	}
	return strings.Join(parts, "?")
}

func validateRealNameSetting(key, value string) error {
	switch key {
	case "realname_student_no_pattern":
		if value == "" || len(value) > 1024 {
			return fmt.Errorf("学号正则表达式必填且最多 1024 字节")
		}
		if _, err := regexp.Compile(`\A(?:` + value + `)\z`); err != nil {
			return fmt.Errorf("学号正则表达式无效，请使用 Go（RE2）正则语法：%v", err)
		}
	case "realname_student_no_placeholder":
		if strings.TrimSpace(value) == "" || utf8.RuneCountInString(value) > 200 {
			return fmt.Errorf("学号占位提示需为 1–200 个字符")
		}
	case "realname_api_url":
		if value == "" {
			return nil
		} // Empty URL explicitly disables real-name binding.
		raw, rawErr := url.Parse(value)
		u, err := url.Parse(realNameURL(value, "2026000001", "测试姓名"))
		if rawErr != nil || strings.ContainsAny(raw.Host, "{}") || err != nil || len(value) > 4096 || !strings.Contains(value, "{id}") || (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" || u.User != nil || u.Fragment != "" || strings.ContainsAny(u.Host, "{}") {
			return fmt.Errorf("认证 API 需为包含 {id} 的 HTTP(S) URL，可选 {name}，不能含账号或片段")
		}
	case "realname_success_field", "realname_id_field", "realname_name_field":
		if !validJSONPath(value) {
			return fmt.Errorf("JSON 字段路径不能为空，请使用 data.id 或 data.0.name 等路径")
		}
	case "realname_success_value":
		var v any
		if json.Unmarshal([]byte(value), &v) != nil {
			return fmt.Errorf(`成功值必须是 JSON 值，例如 true、0 或 "ok"`)
		}
		switch v.(type) {
		case map[string]any, []any:
			return fmt.Errorf("成功值需为布尔值、数字、字符串或 null")
		}
	}
	return nil
}

func jsonMatches(actual any, expected string) bool {
	var value any
	decoder := json.NewDecoder(strings.NewReader(expected))
	decoder.UseNumber()
	if decoder.Decode(&value) != nil {
		return false
	}
	a, errA := json.Marshal(actual)
	b, errB := json.Marshal(value)
	return errA == nil && errB == nil && string(a) == string(b)
}

func identityText(value any) (string, bool) {
	switch v := value.(type) {
	case string:
		return v, true
	case json.Number:
		return string(v), true
	default:
		return "", false
	}
}
