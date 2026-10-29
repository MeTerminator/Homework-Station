package app

import (
	"fmt"
	"os"
	"testing"

	"github.com/gin-gonic/gin"

	"homework-station/internal/config"
	"homework-station/internal/model"
)

func TestOwnAttachmentManagement(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("DATA_DIR", t.TempDir())
	t.Setenv("DB_DRIVER", "sqlite")
	t.Setenv("DB_DSN", "")
	t.Setenv("JWT_SECRET", "test-session-secret-32-characters-long")
	application, err := New(config.Load())
	if err != nil {
		t.Fatal(err)
	}
	defer application.Close()
	admin := &testClient{t: t, r: application.Router}
	owner := &testClient{t: t, r: application.Router}
	other := &testClient{t: t, r: application.Router}
	guest := &testClient{t: t, r: application.Router}
	admin.request("POST", "/api/login", gin.H{"username": "admin", "password": "123456"}, 200)
	user := owner.register("attachment-owner")
	other.register("attachment-other")
	assignment := decode[model.Assignment](t, admin.request("POST", "/api/admin/assignments", gin.H{
		"name": "附件管理测试", "status": "published", "allow_attachments": true,
		"extensions": "txt", "max_size": 1024, "multiple": true,
	}, 200))
	content := []byte("shared attachment content")
	file := owner.send(owner.init("answer.txt", content, assignment.ID), content)
	shared := other.send(other.init("other.txt", content, assignment.ID), content)
	partial := owner.init("unfinished.txt", []byte("partial"), assignment.ID)
	owner.request("PUT", fmt.Sprintf("/api/uploads/%d/chunks/0", partial.ID), []byte("partial"), 200)
	owner.request("GET", fmt.Sprintf("/api/attachments/%d/download", partial.ID), nil, 409)
	owner.request("POST", fmt.Sprintf("/api/assignments/%d/submissions", assignment.ID), gin.H{"attachment_ids": []uint{file.ID}}, 201)
	if err := application.DB.Model(&model.User{}).Where("id = ?", user.ID).Update("avatar_id", file.ID).Error; err != nil {
		t.Fatal(err)
	}
	files := decode[[]model.Attachment](t, owner.request("GET", "/api/attachments?user_id=1", nil, 200))
	if len(files) != 2 {
		t.Fatalf("own list returned %d files, want 2", len(files))
	}
	for _, f := range files {
		if f.UserID != user.ID {
			t.Fatal("own list leaked another user's attachment")
		}
		if f.ID == file.ID && len(f.Usage) != 2 {
			t.Fatal("submission or avatar usage missing")
		}
		if f.ID == partial.ID && len(f.Received) != 1 {
			t.Fatal("received chunk progress missing")
		}
	}
	path := fmt.Sprintf("/api/attachments/%d", file.ID)
	guest.request("DELETE", path+"?force=true", nil, 401)
	other.request("DELETE", path+"?force=true", nil, 404)
	// Even an admin must use the admin endpoint to delete someone else's file.
	admin.request("DELETE", path+"?force=true", nil, 404)
	owner.request("DELETE", path, nil, 409)
	download := owner.request("GET", path+"/download", nil, 200)
	if download.Body.String() != string(content) {
		t.Fatal("download content mismatch")
	}
	owner.request("DELETE", path+"?force=true", nil, 200)
	owner.request("GET", path+"/download", nil, 404)
	me := decode[model.User](t, owner.request("GET", "/api/me", nil, 200))
	if me.AvatarID != nil {
		t.Fatal("deleted avatar still linked to user")
	}
	var links int64
	application.DB.Model(&model.SubmissionAttachment{}).Where("attachment_id = ?", file.ID).Count(&links)
	if links != 0 {
		t.Fatal("deleted attachment still linked to submission")
	}
	preserved := other.request("GET", fmt.Sprintf("/api/attachments/%d/download", shared.ID), nil, 200)
	if preserved.Body.String() != string(content) {
		t.Fatal("deletion damaged shared file")
	}
	owner.request("DELETE", fmt.Sprintf("/api/attachments/%d", partial.ID), nil, 200)
	if _, err := os.Stat(application.Handler.Files.ChunkDir(partial.ID)); !os.IsNotExist(err) {
		t.Fatalf("partial chunks remain: %v", err)
	}
	files = decode[[]model.Attachment](t, owner.request("GET", "/api/attachments", nil, 200))
	if len(files) != 0 {
		t.Fatal("deleted attachments still listed")
	}
}
