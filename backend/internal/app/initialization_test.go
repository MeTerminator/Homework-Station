package app

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"homework-station/internal/config"
	"homework-station/internal/model"
)

func TestInitializationContainsNoSampleContent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("DATA_DIR", t.TempDir())
	t.Setenv("DB_DRIVER", "sqlite")
	t.Setenv("DB_DSN", "")
	t.Setenv("JWT_SECRET", "test-session-secret-32-characters-long")
	application, err := New(config.Load())
	if err != nil {
		t.Fatal(err)
	}
	db := application.DB
	defer func() {
		application.Close()
	}()
	var users []model. // A second startup preserves the administrator and leaves content tables empty.
				User
	if err := db.Find(&users).Error; err != nil {
		t.Fatal(err)
	}
	if len(users) != 1 || users[0].Username != "admin" || !users[0].Admin || bcrypt.CompareHashAndPassword([]byte(users[0].Password), []byte("123456")) != nil {
		t.Fatal("initialization must create only the default administrator")
	}
	for _, model := range []any{&model.Category{}, &model.Assignment{}, &model.Submission{}, &model.Attachment{}, &model.Blob{}, &model.UsedCode{}, &model.SubmissionAttachment{}, &model.AssignmentAttachment{}} {
		var count int64
		if err := db.Model(model).Count(&count).Error; err != nil {
			t.Fatal(err)
		}
		if count != 0 {
			t.Fatalf("initialization created sample content in %T", model)
		}
	}
	if application.Handler.Settings.Get("realname_api_url") != "" {
		t.Fatal("initialization must not configure an external identity service")
	}
	client := &testClient{t: t, r: application.Router}
	public := decode[map[string]any](t, client.request("GET", "/api/public-settings", nil, 200))
	if public["realname_enabled"] != false {
		t.Fatal("real-name verification must be disabled before configuration")
	}
	if _, ok := public["realname_api_url"]; ok {
		t.Fatal("public settings expose the private identity service URL")
	}
	client.request("POST", "/api/login", gin.H{"username": "admin", "password": "123456"}, 200)
	oldClient := application.Handler.Identity.Client
	application.Handler.Identity.Client = &http.Client{Transport: roundTrip(func(*http.Request) (*http.Response, error) {
		t.Fatal("an unconfigured identity check must not make a network request")
		return nil, nil
	})}
	defer func() {
		application.Handler.Identity.Client = oldClient
	}()
	client.request("POST", "/api/me/bind/real", gin.H{"student_no": "2026000001", "name": "测试用户"}, 503)
	client.request("PUT", "/api/admin/settings", gin.H{"realname_api_url": "https://identity.example/verify?id={id}&name={name}"}, 200)
	public = decode[map[string]any](t, client.request("GET", "/api/public-settings", nil, 200))
	if public["realname_enabled"] != true {
		t.Fatal("configured service was not enabled")
	}
	client.request("PUT", "/api/admin/settings", gin.H{"realname_api_url": ""}, 200)
	client.request("POST", "/api/me/bind/real", gin.H{"student_no": "2026000001", "name": "测试用户"}, 503)
	application.Close()
	application, err = New(config.Load())
	if err != nil {
		t.Fatal(err)
	}
	db = application.DB
	var admin model.User
	db.First(&admin, users[0].ID)
	if admin.Password != users[0].Password {
		t.Fatal("startup reset the existing administrator password")
	}
	var count int64
	db.Model(&model.Assignment{}).Count(&count)
	if count != 0 || application.Handler.Settings.Get("realname_api_url") != "" {
		t.Fatal("startup repopulated content or an identity endpoint")
	}
}
