package app

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"homework-station/internal/config"
	"homework-station/internal/model"
)

func TestAssignmentSummaryUpgradeAndEditing(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("DATA_DIR", t.TempDir())
	t.Setenv("DB_DRIVER", "sqlite")
	t.Setenv("DB_DSN", "")
	t.Setenv("JWT_SECRET", "test-session-secret-32-characters-long")
	t.Run("sqlite", func(t *testing.T) {
		testAssignmentSummaryUpgradeAndEditing(t, config.Load())
	})
	t.Run("mysql", func(t *testing.T) {
		dsn := os.Getenv("MYSQL_TEST_DSN")
		if dsn == "" {
			t.Skip("MYSQL_TEST_DSN is not configured")
		}
		cfg := config.Load()
		cfg.Database = config.Database{Driver: "mysql", DSN: dsn}
		testAssignmentSummaryUpgradeAndEditing(t, cfg)
	})
}

func testAssignmentSummaryUpgradeAndEditing(t *testing.T, cfg config.Config) {
	t.Helper()
	application, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { application.Close() }()
	admin := &testClient{t: t, r: application.Router}
	admin.request("POST", "/api/login", gin.H{"username": "admin", "password": "123456"}, 200)
	assignment := decode[model.Assignment](t, admin.request("POST", "/api/admin/assignments", gin.H{
		"name": "已有作业", "description": "# 保留完整正文", "status": "published", "max_size": 1024, "multiple": true,
	}, 200))
	path := fmt.Sprintf("/api/assignments/%d", assignment.ID)
	editPath := fmt.Sprintf("/api/admin/assignments/%d", assignment.ID)
	submission := decode[model.Submission](t, admin.request("POST", path+"/submissions", gin.H{"content": "已有提交"}, 201))
	// Remove only the new column to reproduce the schema of the deployed version.
	if err := application.DB.Migrator().DropColumn(&model.Assignment{}, "Summary"); err != nil {
		t.Fatal(err)
	}
	application.Close()
	application, err = New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	admin.r = application.Router
	migrated := decode[struct {
		Assignment model.Assignment `json:"assignment"`
	}](t, admin.request("GET", path, nil, 200)).Assignment
	if migrated.Summary != "" || migrated.Description != assignment.Description || migrated.ID != assignment.ID || !migrated.CreatedAt.Equal(assignment.CreatedAt) {
		t.Fatalf("migration changed existing assignment: %+v", migrated)
	}
	var retained model.Submission
	if err := application.DB.First(&retained, submission.ID).Error; err != nil || retained.Content != "已有提交" {
		t.Fatalf("submission not preserved: %+v, %v", retained, err)
	}
	// Exercise create, edit, Unicode boundary validation and clearing through the real API.
	created := decode[model.Assignment](t, admin.request("POST", "/api/admin/assignments", gin.H{
		"name": "新作业", "summary": "认识 Linux", "description": "详细要求", "status": "published", "max_size": 1024,
	}, 200))
	if created.Summary != "认识 Linux" {
		t.Fatal("summary not saved on create")
	}
	assignment.Summary = "  " + strings.Repeat("简", 200) + "  "
	saved := decode[model.Assignment](t, admin.request("PUT", editPath, assignment, 200))
	if saved.Summary != strings.Repeat("简", 200) {
		t.Fatal("summary not trimmed and saved")
	}
	assignment.Summary = strings.Repeat("简", 201)
	admin.request("PUT", editPath, assignment, 400)
	list := decode[struct {
		Items []model.Assignment `json:"items"`
	}](t, admin.request("GET", "/api/assignments", nil, 200))
	found := false
	for _, item := range list.Items {
		if item.ID == assignment.ID {
			found = true
			if item.Summary != saved.Summary {
				t.Fatal("list did not return persisted summary")
			}
		}
	}
	if !found {
		t.Fatal("existing assignment missing from list")
	}
	// A subsequent startup must retain a filled summary and all existing content.
	application.Close()
	application, err = New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	admin.r = application.Router
	detail := decode[struct {
		Assignment model.Assignment `json:"assignment"`
	}](t, admin.request("GET", path, nil, 200)).Assignment
	if detail.Summary != saved.Summary || detail.Description != assignment.Description {
		t.Fatal("repeated migration changed content")
	}
	assignment.Summary = ""
	cleared := decode[model.Assignment](t, admin.request("PUT", editPath, assignment, 200))
	if cleared.Summary != "" {
		t.Fatal("summary could not be cleared")
	}
}
