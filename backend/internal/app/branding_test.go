package app

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"mime/multipart"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"homework-station/internal/config"
)

func TestBrandingSettings(t *testing.T) {
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
	guest := &testClient{t: t, r: application.Router}
	guest.request("PUT", "/api/admin/settings", gin.H{"site_title": "Example School"}, 401)
	guest.request("POST", "/api/admin/settings/images", nil, 401)
	student := &testClient{t: t, r: application.Router}
	student.register("branding-student")
	student.request("POST", "/api/admin/settings/images", nil, 403)
	admin.request("POST", "/api/login", gin.H{"username": "admin", "password": "123456"}, 200)

	upload := func(data []byte, want int) *httptest.ResponseRecorder {
		t.Helper()
		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		part, err := writer.CreateFormFile("image", "logo.png")
		if err != nil {
			t.Fatal(err)
		}
		part.Write(data)
		writer.Close()
		req := httptest.NewRequest("POST", "/api/admin/settings/images", &body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.AddCookie(admin.cookie)
		res := httptest.NewRecorder()
		application.Router.ServeHTTP(res, req)
		if res.Code != want {
			t.Fatalf("upload = %d want %d: %s", res.Code, want, res.Body.String())
		}
		return res
	}
	upload([]byte("<svg onload='alert(1)'></svg>"), 400)
	upload(make([]byte, 2*1024*1024+1), 400)
	img := image.NewNRGBA(image.Rect(0, 0, 32, 32))
	img.Set(0, 0, color.NRGBA{R: 255, A: 128})
	var input bytes.Buffer
	if err := png.Encode(&input, img); err != nil {
		t.Fatal(err)
	}
	url := decode[map[string]string](t, upload(input.Bytes(), 201))["url"]
	asset := guest.request("GET", url, nil, 200)
	decoded, err := png.Decode(bytes.NewReader(asset.Body.Bytes()))
	if err != nil || decoded.At(0, 0) != img.At(0, 0) {
		t.Fatal("uploaded image did not preserve transparency")
	}
	guest.request("GET", "/api/branding/unknown.png", nil, 404)
	values := map[string]string{
		"site_title":    "Homework Hub 学习站",
		"browser_title": "Example School — Homework",
		"site_logo":     url,
		"site_icon":     url,
	}
	admin.request("PUT", "/api/admin/settings", values, 200)
	public := decode[map[string]any](t, guest.request("GET", "/api/public-settings", nil, 200))
	for k, v := range values {
		if public[k] != v {
			t.Fatalf("public setting %s = %v want %s", k, public[k], v)
		}
	}
	if _, ok := public["binding_secret"]; ok {
		t.Fatal("public settings exposed binding secret")
	}
	for _, bad := range []map[string]string{
		{"site_logo": "javascript:alert(1)"},
		{"site_icon": "//external.example/icon.png"},
		{"site_icon": "/\\external.example/icon.png"},
		{"site_title": "  "},
		{"site_title": strings.Repeat("中", 121)},
		{"browser_title": strings.Repeat("a", 121)},
	} {
		admin.request("PUT", "/api/admin/settings", bad, 400)
	}
	admin.request("PUT", "/api/admin/settings", gin.H{"site_title": strings.Repeat("中", 120)}, 200)
	admin.request("PUT", "/api/admin/settings", values, 200)
	application.Close()
	application, err = New(config.Load())
	if err != nil {
		t.Fatal(err)
	}
	for k, v := range values {
		if application.Handler.Settings.Get(k) != v {
			t.Fatalf("setting %s lost after restart", k)
		}
	}
	guest.r = application.Router
	guest.request("GET", url, nil, 200)
	admin.r = application.Router
	admin.request("PUT", "/api/admin/settings", gin.H{"browser_title": "", "site_logo": "", "site_icon": ""}, 200)
	public = decode[map[string]any](t, guest.request("GET", "/api/public-settings", nil, 200))
	if public["site_logo"] != "" || public["site_icon"] != "" || public["browser_title"] != "" || public["site_title"] != values["site_title"] {
		t.Fatal("reset did not clear branding overrides")
	}
}
