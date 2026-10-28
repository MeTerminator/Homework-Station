package app

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"homework-station/internal/config"
	"homework-station/internal/model"
)

func TestConfigurableStudentNumber(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("DATA_DIR", t.TempDir())
	t.Setenv("DB_DRIVER", "sqlite")
	t.Setenv("DB_DSN", "")
	t.Setenv("JWT_SECRET", "test-session-secret-32-characters-long")
	application, err := New(config.Load())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { application.Close() }()
	admin := &testClient{t: t, r: application.Router}
	student := &testClient{t: t, r: application.Router}
	admin.request("POST", "/api/login", gin.H{"username": "admin", "password": "123456"}, 200)
	user := student.register("custom-student-number")
	defaults := decode[map[string]string](t, admin.request("GET", "/api/admin/settings", nil, 200))
	if defaults["realname_student_no_pattern"] != `^[0-9]{10}$` || defaults["realname_student_no_placeholder"] != "10 位数字学号" {
		t.Fatal("incorrect defaults")
	}
	for _, invalid := range []gin.H{
		{"realname_student_no_pattern": "["},
		{"realname_student_no_pattern": ""},
		{"realname_student_no_pattern": strings.Repeat("a", 1025)},
		{"realname_student_no_placeholder": " "},
		{"realname_student_no_placeholder": strings.Repeat("字", 201)},
	} {
		invalid["site_title"] = "must not save"
		admin.request("PUT", "/api/admin/settings", invalid, 400)
		if application.Handler.Settings.Get("site_title") != defaults["site_title"] {
			t.Fatal("invalid settings partially saved")
		}
	}
	// Omitted anchors must still match the whole student number.
	settings := gin.H{
		"realname_student_no_pattern":     `[A-Z]{2}[0-9]{12}`,
		"realname_student_no_placeholder": "两位大写字母和十二位数字，例如 AB202600000001",
		"realname_api_url":                "https://identity.example/verify?id={id}&name={name}",
	}
	student.request("PUT", "/api/admin/settings", settings, 403)
	admin.request("PUT", "/api/admin/settings", settings, 200)
	public := decode[map[string]any](t, student.request("GET", "/api/public-settings", nil, 200))
	if public["realname_student_no_placeholder"] != settings["realname_student_no_placeholder"] {
		t.Fatal("placeholder not published")
	}
	if _, exists := public["realname_api_url"]; exists {
		t.Fatal("private API configuration leaked")
	}
	calls := 0
	application.Handler.Identity.Client = &http.Client{Transport: roundTrip(func(req *http.Request) (*http.Response, error) {
		calls++
		body, _ := json.Marshal(gin.H{"success": true, "data": gin.H{"id": req.URL.Query().Get("id"), "name": req.URL.Query().Get("name")}})
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(string(body)))}, nil
	})}
	for _, invalid := range []string{"2026000001", "ab202600000001", "XAB202600000001", "AB202600000001X", "AB202600000001\n", strings.Repeat("a", 129)} {
		result := student.request("POST", "/api/me/bind/real", gin.H{"student_no": invalid, "name": "测试姓名"}, 400)
		if len(invalid) <= 128 && !strings.Contains(result.Body.String(), settings["realname_student_no_placeholder"].(string)) {
			t.Fatal("missing format hint")
		}
	}
	if calls != 0 {
		t.Fatal("invalid student number called identity API")
	}
	bound := decode[model.User](t, student.request("POST", "/api/me/bind/real", gin.H{"student_no": "AB202600000001", "name": "测试姓名"}, 200))
	if bound.StudentNo == nil || *bound.StudentNo != "AB202600000001" || calls != 1 {
		t.Fatal("custom student number did not bind")
	}
	admin.request("POST", fmt.Sprintf("/api/admin/users/%d/unbind/real", user.ID), nil, 200)
	application.Close()
	application, err = New(config.Load())
	if err != nil {
		t.Fatal(err)
	}
	for key, value := range settings {
		if application.Handler.Settings.Get(key) != value {
			t.Fatalf("%s not persisted across restart", key)
		}
	}
}
