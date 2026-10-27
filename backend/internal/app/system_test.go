package app

import (
	"archive/zip"
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"homework-station/internal/config"
	"homework-station/internal/model"
	"homework-station/internal/service"
)

func TestSystem(t *testing.T) {
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
	dataDir := application.Config.DataDir
	_ = dataDir
	defer func() {
		application.Close()
	}()
	r := application.Router
	admin := &testClient{t: t, r: r}
	student := &testClient{t: t, r: r}
	other := &testClient{t: t, r: r}
	guest := &testClient{t: t, r: r}
	guest.request("GET", "/api/me", nil, 401)
	admin.request("POST", "/api/login", gin.H{"username": "admin", "password": "wrong"}, 401)
	adminUser := decode[model.User](t, admin.request("POST", "/api/login", gin.H{"username": "admin", "password": "123456"}, 200))
	if !adminUser.Admin {
		t.Fatal("admin missing")
	}
	u := student.register("student")
	otherUser := other.register("other")
	student.request("GET", "/api/admin/overview", nil, 403)
	cat := decode[model.Category](t, admin.request("POST", "/api/admin/categories", gin.H{"name": "前端"}, 200))
	child := decode[model.Category](t, admin.request("POST", "/api/admin/categories", gin.H{"name": "React", "parent_id": cat.ID}, 200))
	admin.request("PUT", fmt.Sprintf("/api/admin/categories/%d", cat.ID), gin.H{"name": "循环", "parent_id": child.ID}, 400)
	admin.request("DELETE", fmt.Sprintf("/api/admin/categories/%d", cat.ID), nil, 409)
	assignment := model.Assignment{Name: "组件设计", CategoryID: &child.ID, Status: "published", Description: "# 作业", AllowAttachments: true, Extensions: "txt,zip", MaxSize: 3 * 1024 * 1024, Multiple: false, Withdraw: true}
	assignment = decode[model.Assignment](t, admin.request("POST", "/api/admin/assignments", assignment, 200))
	path := fmt.Sprintf("/api/assignments/%d", assignment.ID)
	student.request("PATCH", "/api/me", gin.H{"category_ids": []uint{cat.ID, child.ID}}, 200)
	info := decode[struct {
		User model.User `json:"user"`
	}](t, admin.request("GET", fmt.Sprintf("/api/admin/users/%d", u.ID), nil, 200))
	if info.User.CategoryIDs != fmt.Sprintf("[%d,%d]", cat.ID, child.ID) {
		t.Fatal("category preferences not saved")
	}
	content := bytes.Repeat([]byte("abc123"), 220000)
	f := student.init("answer.txt", content, assignment.ID)
	student.request("POST", fmt.Sprintf("/api/uploads/%d/complete", f.ID), nil, 409)
	other.request("GET", fmt.Sprintf("/api/uploads/%d", f.ID), nil, 404)
	other.request("GET", fmt.Sprintf("/api/attachments/%d/download", f.ID), nil, 404)
	student.request("PUT", fmt.Sprintf("/api/uploads/%d/chunks/0", f.ID), content[:100], 400)
	student.request("PUT", fmt.Sprintf("/api/uploads/%d/chunks/0", f.ID), content[:service.ChunkBytes], 200)
	resume := decode[model.Attachment](t, student.request("POST", "/api/uploads", gin.H{"name": "answer.txt", "size": len(content), "sha256": sha(content), "assignment_id": assignment.ID}, 200))
	if resume.ID != f.ID || len(resume.Received) != 1 {
		t.Fatal("resume failed")
	}
	archive := admin.request("GET", fmt.Sprintf("/api/attachments/%d/download", f.ID), nil, 200)
	zr, e := zip.NewReader(bytes.NewReader(archive.Body.Bytes()), int64(archive.Body.Len()))
	if e != nil || len(zr.File) != 2 {
		t.Fatal("partial ZIP missing manifest/chunk", e)
	}
	student.request("PUT", fmt.Sprintf("/api/uploads/%d/chunks/1", f.ID), content[service.ChunkBytes:], 200)
	f = decode[model.Attachment](t, student.request("POST", fmt.Sprintf("/api/uploads/%d/complete", f.ID), nil, 200))
	download := student.request("GET", fmt.Sprintf("/api/attachments/%d/download", f.ID), nil, 200)
	if !bytes.Equal(download.Body.Bytes(), content) {
		t.Fatal("download mismatch")
	}
	second := other.send(other.init("copy.txt", content, assignment.ID), content)
	var blobs int64
	db.Model(&model.Blob{}).Count(&blobs)
	if blobs != 1 {
		t.Fatal("dedup failed", blobs)
	}
	other.request("POST", path+"/submissions", gin.H{"attachment_ids": []uint{f.ID}}, 400)
	student.request("POST", path+"/submissions", gin.H{}, 400)
	sub := decode[model.Submission](t, student.request("POST", path+"/submissions", gin.H{"content": "完成", "attachment_ids": []uint{f.ID}}, 201))
	student.request("POST", path+"/submissions", gin.H{"content": "重复"}, 409)
	otherSubs := decode[[]model.Submission](t, other.request("GET", path+"/submissions", nil, 200))
	if len(otherSubs) != 0 {
		t.Fatal("submission data leaked")
	}
	other.request("POST", fmt.Sprintf("/api/submissions/%d/withdraw", sub.ID), nil, 404)
	student.request("POST", fmt.Sprintf("/api/submissions/%d/withdraw", sub.ID), nil, 200)
	sub2 := decode[model.Submission](t, student.request("POST", path+"/submissions", gin.H{"content": "重交", "attachment_ids": []uint{f.ID}}, 201))
	all := decode[[]model.Submission](t, admin.request("GET", path+"/submissions", nil, 200))
	if len(all) != 2 {
		t.Fatal("history lost")
	}
	bad := student.init("bad.txt", []byte("expected"), assignment.ID)
	student.request("PUT", fmt.Sprintf("/api/uploads/%d/chunks/0", bad.ID), []byte("badbytes"), 200)
	student.request("POST", fmt.Sprintf("/api/uploads/%d/complete", bad.ID), nil, 422)
	student.request("POST", "/api/uploads", gin.H{"name": "bad.exe", "size": 1, "sha256": sha([]byte{1}), "assignment_id": assignment.ID}, 400)
	deadline := time.Now().Add(-time.Hour)
	assignment.Deadline = &deadline
	admin.request("PUT", fmt.Sprintf("/api/admin/assignments/%d", assignment.ID), assignment, 200)
	other.request("POST", path+"/submissions", gin.H{"content": "迟交"}, 409)
	assignment.Deadline = nil
	assignment.Status = "hidden"
	admin.request("PUT", fmt.Sprintf("/api/admin/assignments/%d", assignment.ID), assignment, 200)
	student.request("GET", path, nil, 404)
	assignment.Status = "published"
	admin.request("PUT", fmt.Sprintf("/api/admin/assignments/%d", assignment.ID), assignment, 200)
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	img.Set(1, 1, color.RGBA{R: 255, A: 255})
	var pngData bytes.Buffer
	png.Encode(&pngData, img)
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	part, _ := mw.CreateFormFile("avatar", "avatar.png")
	part.Write(pngData.Bytes())
	mw.Close()
	req := httptest.NewRequest("POST", "/api/me/avatar", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.AddCookie(student.cookie)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatal(rec.Code, rec.Body.String())
	}
	avatarUser := decode[model.User](t, rec)
	if avatarUser.AvatarID == nil {
		t.Fatal("missing avatar")
	}
	av := student.request("GET", fmt.Sprintf("/api/attachments/%d/download?inline=true", *avatarUser.AvatarID), nil, 200)
	if !strings.Contains(string(av.Body.Bytes()[:12]), "WEBP") {
		t.Fatal("not WebP")
	}
	admin.request("PUT", "/api/admin/settings", gin.H{"realname_api_url": "https://identity.example/verify?id={id}"}, 200)
	oldClient := application.Handler.Identity.Client
	application.Handler.Identity.Client = &http.Client{Transport: roundTrip(func(req *http.Request) (*http.Response, error) {
		if req.URL.Query().Get("id") != "2026000001" {
			t.Fatal("wrong student query")
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"success":true,"data":{"id":"2026000001","name":"张三"}}`))}, nil
	})}
	defer func() {
		application.Handler.Identity.Client = oldClient
	}()
	student.request("POST", "/api/me/bind/real", gin.H{"student_no": "123", "name": "张三"}, 400)
	student.request("POST", "/api/me/bind/real", gin.H{"student_no": "2026000001", "name": "李四"}, 400)
	student.request("POST", "/api/me/bind/real", gin.H{"student_no": "2026000001", "name": "张三"}, 200)
	other.request("POST", "/api/me/bind/real", gin.H{"student_no": "2026000001", "name": "张三"}, 409)
	claims := jwt.MapClaims{"qq": "123456789", "iat": time.Now().Unix(), "exp": time.Now().Add(5 * time.Minute).Unix(), "ttl": 300, "aud": "homework-qq-bind"}
	code, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(application.Handler.Settings.Get("binding_secret")))
	student.request("POST", "/api/me/bind/qq", gin.H{"code": "invalid"}, 400)
	student.request("POST", "/api/me/bind/qq", gin.H{"code": code}, 200)
	admin.request("POST", fmt.Sprintf("/api/admin/users/%d/unbind/qq", u.ID), nil, 200)
	student.request("POST", "/api/me/bind/qq", gin.H{"code": code}, 409)
	claims["iat"] = time.Now().Add(-10 * time.Minute).Unix()
	claims["exp"] = time.Now().Add(-5 * time.Minute).Unix()
	expired, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(application.Handler.Settings.Get("binding_secret")))
	student.request("POST", "/api/me/bind/qq", gin.H{"code": expired}, 400)
	admin.request("DELETE", fmt.Sprintf("/api/admin/attachments/%d", f.ID), nil, 409)
	admin.request("DELETE", fmt.Sprintf("/api/admin/submissions/%d?files=true", sub2.ID), nil, 200)
	var still model.Attachment
	if db.First(&still, f.ID).Error != nil {
		t.Fatal("file still referenced by historical submission removed")
	}
	admin.request("DELETE", fmt.Sprintf("/api/admin/submissions/%d?files=true", sub.ID), nil, 200)
	if _, e := os.Stat(filepath.Join(dataDir, "blobs", f.SHA256)); e != nil {
		t.Fatal("shared blob deleted")
	}
	other.request("GET", fmt.Sprintf("/api/attachments/%d/download", second.ID), nil, 200)
	admin.request("DELETE", fmt.Sprintf("/api/admin/attachments/%d", second.ID), nil, 200)
	if _, e := os.Stat(filepath.Join(dataDir, "blobs", f.SHA256)); !os.IsNotExist(e) {
		t.Fatal("unreferenced blob retained")
	}
	admin.request("DELETE", fmt.Sprintf("/api/admin/assignments/%d", assignment.ID), nil, 200)
	admin.request("DELETE", fmt.Sprintf("/api/admin/categories/%d", child.ID), nil, 200)
	overview := decode[map[string]interface{}](t, admin.request("GET", "/api/admin/overview", nil, 200))
	if overview["users"].(float64) != 3 || len(overview["days"].([]interface{})) != 30 {
		t.Fatal("overview incorrect")
	}
	admin.request("PATCH", fmt.Sprintf("/api/admin/users/%d", adminUser.ID), gin.H{"username": "admin", "admin": false}, 409)
	admin.request("PATCH", fmt.Sprintf("/api/admin/users/%d", otherUser.ID), gin.H{"username": "other", "admin": true}, 200)
	other.request("GET", "/api/me", nil, 401)
	oldCookie := student.cookie
	student.request("POST", "/api/me/password", gin.H{"old": "bad", "password": "newpass123", "confirm": "newpass123"}, 400)
	student.request("POST", "/api/me/password", gin.H{"old": "password123", "password": "newpass123", "confirm": "newpass123"}, 200)
	stale := &testClient{t: t, r: r, cookie: oldCookie}
	stale.request("GET", "/api/me", nil, 401)
	admin.request("PUT", "/api/admin/settings", gin.H{"site_title": "新标题", "binding_secret": "persistent-binding-secret-32-characters"}, 200)
	secret := application.Handler.Settings.Get("binding_secret")
	application.Close()
	application, err = New(config.Load())
	if err != nil {
		t.Fatal(err)
	}
	db = application.DB
	if application.Handler.Settings.Get("site_title") != "新标题" || application.Handler.Settings.Get("binding_secret") != secret {
		t.Fatal("settings reset on restart")
	}
	var users int64
	db.Model(&model.User{}).Count(&users)
	if users != 3 {
		t.Fatal("database reinitialized")
	}
	req = httptest.NewRequest("POST", "/api/logout", nil)
	req.Header.Set("Origin", "https://evil.example")
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != 403 {
		t.Fatal("cross-origin write accepted")
	}
}
