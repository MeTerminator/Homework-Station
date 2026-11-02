package app

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"homework-station/internal/config"
	"homework-station/internal/model"
)

func TestSubmissionManagementAndPreviews(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("DATA_DIR", t.TempDir())
	t.Setenv("DB_DRIVER", "sqlite")
	t.Setenv("DB_DSN", "")
	t.Setenv("JWT_SECRET", "test-session-secret-32-characters-long")
	app, err := New(config.Load())
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()
	admin := &testClient{t: t, r: app.Router}
	owner := &testClient{t: t, r: app.Router}
	other := &testClient{t: t, r: app.Router}
	guest := &testClient{t: t, r: app.Router}
	admin.request("POST", "/api/login", gin.H{"username": "admin", "password": "123456"}, 200)
	user := owner.register("preview-owner")
	other.register("preview-other")
	guest.request("GET", "/api/admin/submissions", nil, 401)
	owner.request("GET", "/api/admin/submissions", nil, 403)
	empty := decode[struct {
		Items []model.Submission `json:"items"`
		Total int64              `json:"total"`
	}](t, admin.request("GET", "/api/admin/submissions", nil, 200))
	if empty.Items == nil || empty.Total != 0 {
		t.Fatal("empty list must contain [] and zero total")
	}
	assignments := []model.Assignment{{Name: "First", Status: "published"}, {Name: "Second", Status: "hidden"}}
	for i := range assignments {
		if err := app.DB.Create(&assignments[i]).Error; err != nil {
			t.Fatal(err)
		}
	}
	now := time.Now().UTC().Truncate(time.Second)
	// CreatedAt differs from ID order; equal timestamps must still paginate stably.
	for i := 0; i < 12; i++ {
		created := now.Add(-time.Duration(i) * time.Hour)
		if i == 1 {
			created = now
		}
		s := model.Submission{AssignmentID: assignments[i%2].ID, UserID: user.ID, Content: fmt.Sprint(i), CreatedAt: created, Withdrawn: i == 1}
		if err := app.DB.Create(&s).Error; err != nil {
			t.Fatal(err)
		}
	}
	type page struct {
		Items []model.Submission `json:"items"`
		Total int64              `json:"total"`
	}
	first := decode[page](t, admin.request("GET", "/api/admin/submissions?page=1&size=10", nil, 200))
	second := decode[page](t, admin.request("GET", "/api/admin/submissions?page=2&size=10", nil, 200))
	if first.Total != 12 || len(first.Items) != 10 || len(second.Items) != 2 {
		t.Fatal("pagination mismatch")
	}
	if first.Items[0].Content != "1" || !first.Items[0].Withdrawn || first.Items[1].Content != "0" {
		t.Fatal("timestamp/id ordering mismatch")
	}
	for _, s := range append(first.Items, second.Items...) {
		if s.Username != user.Username || s.AssignmentName == "" {
			t.Fatal("missing user/assignment label")
		}
	}
	cases := []struct {
		name, content, mime string
		inline              bool
	}{
		{"answer.md", "# Render me\n\n**hello**", "application/octet-stream", false},
		{"answer.py", "print('hello')", "application/octet-stream", false},
		{"answer.pdf", "%PDF-1.4\n%%EOF", "application/pdf", true},
		{"image.gif", "GIF89a", "image/gif", true},
		{"attack.pdf", "<!DOCTYPE html><script>alert(1)</script>", "application/octet-stream", false},
		{"attack.svg", "<svg xmlns=\"http://www.w3.org/2000/svg\" onload=\"alert(1)\"/>", "application/octet-stream", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := owner.send(owner.init(tc.name, []byte(tc.content), 0), []byte(tc.content))
			url := fmt.Sprintf("/api/attachments/%d/download?inline=true", f.ID)
			guest.request("GET", url, nil, 401)
			other.request("GET", url, nil, 404)
			res := admin.request("GET", url, nil, 200)
			if res.Body.String() != tc.content || res.Header().Get("Content-Type") != tc.mime {
				t.Fatal("preview content or MIME mismatch", res.Header())
			}
			if tc.inline != (res.Header().Get("Content-Disposition") == "inline") {
				t.Fatal("unsafe inline disposition")
			}
			if tc.mime == "application/pdf" && res.Header().Get("X-Frame-Options") != "SAMEORIGIN" {
				t.Fatal("PDF frame blocked")
			}
			download := admin.request("GET", strings.TrimSuffix(url, "?inline=true"), nil, 200)
			if !strings.HasPrefix(download.Header().Get("Content-Disposition"), "attachment;") {
				t.Fatal("original download no longer available")
			}
			if err := app.DB.Model(&model.Submission{ID: first.Items[0].ID}).Association("Attachments").Append(&f); err != nil {
				t.Fatal(err)
			}
		})
	}
	withFiles := decode[page](t, admin.request("GET", "/api/admin/submissions", nil, 200))
	if len(withFiles.Items[0].Attachments) != len(cases) {
		t.Fatal("submission attachments missing")
	}
	admin.request("DELETE", fmt.Sprintf("/api/admin/submissions/%d?files=false", first.Items[0].ID), nil, 200)
	after := decode[page](t, admin.request("GET", "/api/admin/submissions", nil, 200))
	if after.Total != 11 {
		t.Fatal("deleted submission remains")
	}
}
