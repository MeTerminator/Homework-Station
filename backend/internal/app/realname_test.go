package app

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	appconfig "homework-station/internal/config"
	"homework-station/internal/model"
)

func TestConfigurableRealName(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("DATA_DIR", t.TempDir())
	t.Setenv("DB_DRIVER", "sqlite")
	t.Setenv("DB_DSN", "")
	t.Setenv("JWT_SECRET", "test-session-secret-32-characters-long")
	application, err := New(appconfig.Load())
	if err != nil {
		t.Fatal(err)
	}
	db := application.DB
	defer func() {
		application.Close()
	}()
	router := application.Router
	admin := &testClient{t: t, r: router}
	student := &testClient{t: t, r: router}
	admin.request("POST", "/api/login", gin.H{"username": "admin", "password": "123456"}, 200)
	u := student.register("identity-user")
	config := map[ // Failed settings requests must not partially update other settings.
	string]string{"realname_api_url": "https://identity.example/verify/{id}?name={name}&token=abc", "realname_success_field": "meta.code", "realname_success_value": "0", "realname_id_field": "payload.people[0].number", "realname_name_field": "$.payload.people.0.fullName"}
	student.request("PUT", "/api/admin/settings", config, 403)
	admin.request("PUT", "/api/admin/settings", config, 200)
	saved := decode[map[string]string](t, admin.request("GET", "/api/admin/settings", nil, 200))
	for k, v := range config {
		if saved[k] != v {
			t.Fatalf("setting %s was not persisted", k)
		}
	}
	for _, bad := range []map[string]string{{"realname_api_url": "file:///tmp/{id}"}, {"realname_api_url": "https://example.com/verify?name={name}"}, {"realname_api_url": "https://{id}.example.com/"}, {"realname_success_field": "data..ok"}, {"realname_id_field": ""}, {"realname_success_value": "ok"}, {"realname_success_value": "{}"}} {
		admin.request("PUT", "/api/admin/settings", bad, 400)
	}
	oldTitle := application.Handler.Settings.Get("site_title")
	admin.request("PUT", "/api/admin/settings", gin.H{"site_title": "must not save", "realname_success_value": "invalid"}, 400)
	if application.Handler.Settings.Get("site_title") != oldTitle {
		t.Fatal("invalid configuration was partially saved")
	}
	response := `{"meta":{"code":0},"payload":{"people":[{"number":2026000001,"fullName":"张 三&+#"}]}}`
	status := 200
	oldClient := application.Handler.Identity.Client
	application.Handler.Identity.Client = &http.Client{Transport: roundTrip(func(req *http.Request) (*http.Response, error) {
		if req.URL.Path != "/verify/2026000001" || req.URL.Query().Get("name") != "张 三&+#" || req.URL.Query().Get("token") != "abc" {
			t.Fatalf("incorrectly escaped template: %s", req.URL)
		}
		return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader(response))}, nil
	})}
	defer func() {
		application.Handler.Identity.Client = oldClient
	}()
	input := gin.H{"student_no": "2026000001", "name": "张 三&+#"}
	student.request("POST", "/api/me/bind/real", input, 200)
	var bound model.User
	db.First(&bound, u.ID)
	if bound.StudentNo == nil || *bound.StudentNo != "2026000001" || bound.RealName != "张 三&+#" {
		t.Fatal("custom field mapping did not bind correct identity")
	}
	admin.request("POST", fmt.Sprintf("/api/admin/users/%d/unbind/real", u.ID), nil, 200)
	for _, tc := range []struct {
		body   string
		status int
		want   int
	}{{`{"meta":{"code":1}}`, 200, 400}, {`{"meta":{"code":0},"payload":{"people":[]}}`, 200, 502}, {`{"payload":{}}`, 200, 502}, {`{"meta":{"code":0},"payload":{"people":[{"number":"2026000002","fullName":"张 三&+#"}]}}`, 200, 400}, {`{"meta":{"code":0},"payload":{"people":[{"number":"2026000001","fullName":"其他姓名"}]}}`, 200, 400}, {`<html>unavailable</html>`, 200, 502}, {`{}`, 500, 502}} {
		response = tc.body
		status = tc.status
		student.request("POST", "/api/me/bind/real", input, tc.want)
	}
	db.First(&bound, u.ID)
	if bound.StudentNo != nil {
		t.Fatal("failed identity checks must never bind")
	}
	application.Close()
	application, err = New(appconfig.Load())
	if err != nil {
		t.Fatal(err)
	}
	db = application.DB
	for k, v := range config {
		if application.Handler.Settings.Get(k) != v {
			t.Fatalf("custom configuration reset on restart: %s", k)
		}
	}
}
