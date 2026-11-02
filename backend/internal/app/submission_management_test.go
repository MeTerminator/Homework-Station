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

func TestAssignmentSubmissionHistoryScope(t *testing.T) {
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
	adminUser := decode[model.User](t, admin.request("POST", "/api/login", gin.H{"username": "admin", "password": "123456"}, 200))
	ownerUser := owner.register("history-owner")
	other.register("history-other")
	assignment := decode[model.Assignment](t, admin.request("POST", "/api/admin/assignments", model.Assignment{Name: "History", Status: "published", Multiple: true, Withdraw: true, MaxSize: 1024}, 200))
	path := fmt.Sprintf("/api/assignments/%d/submissions", assignment.ID)
	adminPath := fmt.Sprintf("/api/admin/assignments/%d/submissions", assignment.ID)
	ownerSubmission := decode[model.Submission](t, owner.request("POST", path, gin.H{"content": "owner answer"}, 201))
	owner.request("POST", fmt.Sprintf("/api/submissions/%d/withdraw", ownerSubmission.ID), nil, 200)
	owner.request("POST", path, gin.H{"content": "owner resubmission"}, 201)

	// Admins without their own submissions also receive an empty personal history.
	empty := decode[[]model.Submission](t, admin.request("GET", path, nil, 200))
	if empty == nil || len(empty) != 0 {
		t.Fatalf("expected empty admin personal history, got %+v", empty)
	}
	adminSubmission := decode[model.Submission](t, admin.request("POST", path, gin.H{"content": "admin answer"}, 201))
	for _, tc := range []struct {
		name   string
		client *testClient
		userID uint
		count  int
	}{
		{"admin", admin, adminUser.ID, 1},
		{"owner", owner, ownerUser.ID, 2},
		{"other", other, 0, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, query := range []string{"", fmt.Sprintf("?user_id=%d", ownerUser.ID), fmt.Sprintf("?user_id=%d", adminUser.ID)} {
				items := decode[[]model.Submission](t, tc.client.request("GET", path+query, nil, 200))
				if len(items) != tc.count {
					t.Fatalf("expected %d personal submissions, got %+v", tc.count, items)
				}
				for _, s := range items {
					if s.UserID != tc.userID {
						t.Fatalf("personal history contains another user's submission: %+v", s)
					}
				}
			}
		})
	}
	guest.request("GET", adminPath, nil, 401)
	owner.request("GET", adminPath, nil, 403)
	all := decode[[]model.Submission](t, admin.request("GET", adminPath, nil, 200))
	if len(all) != 3 || all[0].ID != adminSubmission.ID || !all[2].Withdrawn {
		t.Fatalf("admin history must include all users and withdrawn submissions: %+v", all)
	}
	filtered := decode[[]model.Submission](t, admin.request("GET", fmt.Sprintf("%s?user_id=%d", adminPath, ownerUser.ID), nil, 200))
	if len(filtered) != 2 {
		t.Fatalf("admin user filter returned unexpected submissions: %+v", filtered)
	}
	for _, s := range filtered {
		if s.UserID != ownerUser.ID {
			t.Fatalf("admin user filter contains another user's submission: %+v", s)
		}
	}
}
