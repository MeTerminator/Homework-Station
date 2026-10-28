package app

import (
	"os"
	"testing"

	"github.com/gin-gonic/gin"

	"homework-station/internal/config"
	"homework-station/internal/service"
)

// MYSQL_TEST_DSN must point to a disposable database dedicated to this test.
func TestMySQLSettings(t *testing.T) {
	dsn := os.Getenv("MYSQL_TEST_DSN")
	if dsn == "" {
		t.Skip("MYSQL_TEST_DSN is not configured")
	}
	gin.SetMode(gin.TestMode)
	cfg := config.Config{
		DataDir:       t.TempDir(),
		SessionSecret: "test-session-secret-32-characters-long",
		Database:      config.Database{Driver: "mysql", DSN: dsn},
	}
	application, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { application.Close() })
	settings := service.NewSettingsService(application.DB)
	defaults := config.DefaultSettings("")
	if got := settings.Get("site_title"); got != defaults["site_title"] {
		t.Fatalf("default site title: got %q, want %q", got, defaults["site_title"])
	}
	values := map[string]string{
		"site_title":                      "MySQL 配置测试",
		"realname_student_no_pattern":     `^[A-Z]{2}[0-9]{8}$`,
		"realname_student_no_placeholder": "两位字母和八位数字",
	}
	if err := settings.Update(values); err != nil {
		t.Fatal(err)
	}
	identity, err := settings.IdentityConfig()
	if err != nil {
		t.Fatal(err)
	}
	if identity["realname_student_no_pattern"] != values["realname_student_no_pattern"] {
		t.Fatal("identity settings were not read back")
	}
	admin := &testClient{t: t, r: application.Router}
	admin.request("POST", "/api/login", gin.H{"username": "admin", "password": "123456"}, 200)
	admin.request("PUT", "/api/admin/settings", gin.H{"browser_title": "MySQL 浏览器标题"}, 200)
	public := decode[map[string]any](t, admin.request("GET", "/api/public-settings", nil, 200))
	if public["site_title"] != values["site_title"] || public["browser_title"] != "MySQL 浏览器标题" {
		t.Fatal("public settings were not read back")
	}
	if err := application.Close(); err != nil {
		t.Fatal(err)
	}
	restarted, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer restarted.Close()
	for key, want := range values {
		if got := restarted.Handler.Settings.Get(key); got != want {
			t.Fatalf("restart overwrote %s: got %q, want %q", key, got, want)
		}
	}
}
