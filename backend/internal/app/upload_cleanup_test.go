package app

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"homework-station/internal/config"
	"homework-station/internal/model"
)

func TestUploadCleanupScopeAndActivity(t *testing.T) {
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
	// Exercise expiration with a fixed clock without the background worker racing it.
	application.cleanupCancel()
	<-application.cleanupDone
	admin := &testClient{t: t, r: application.Router}
	student := &testClient{t: t, r: application.Router}
	admin.request("POST", "/api/login", gin.H{"username": "admin", "password": "123456"}, 200)
	user := student.register("cleanup-user")
	settings := decode[map[string]string](t, admin.request("GET", "/api/admin/settings", nil, 200))
	if settings["upload_cleanup_value"] != "7" || settings["upload_cleanup_unit"] != "day" {
		t.Fatal("default retention is not 7 days")
	}
	student.request("PUT", "/api/admin/settings", gin.H{"upload_cleanup_value": "1", "upload_cleanup_unit": "second"}, 403)
	for _, values := range []gin.H{
		{"upload_cleanup_value": "0"}, {"upload_cleanup_value": "1.5"},
		{"upload_cleanup_value": "-1"}, {"upload_cleanup_unit": "year"},
		{"upload_cleanup_value": "9223372036854775807", "upload_cleanup_unit": "month"},
	} {
		admin.request("PUT", "/api/admin/settings", values, 400)
	}
	retention, err := application.Handler.Settings.UploadRetention()
	if err != nil || retention != 7*24*time.Hour {
		t.Fatalf("invalid setting changed retention: %v, %v", retention, err)
	}
	now := time.Now()
	old := now.Add(-8 * 24 * time.Hour)
	recent := now.Add(-time.Hour)
	files := application.Handler.Files
	create := func(name, status string, activity *time.Time, chunkTime time.Time) model.Attachment {
		t.Helper()
		file := model.Attachment{UserID: user.ID, Name: name, Status: status, SHA256: sha([]byte(name)), Size: 5, TotalChunks: 1, CreatedAt: old, UploadActivityAt: activity}
		if err := application.DB.Create(&file).Error; err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(files.ChunkDir(file.ID), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(files.ChunkPath(file.ID, 0), []byte("chunk"), 0600); err != nil {
			t.Fatal(err)
		}
		if err := os.Chtimes(files.ChunkPath(file.ID, 0), chunkTime, chunkTime); err != nil {
			t.Fatal(err)
		}
		return file
	}
	expired := create("expired.txt", "uploading", &old, old)
	legacyExpired := create("legacy-expired.txt", "uploading", nil, old)
	legacyActive := create("legacy-active.txt", "uploading", nil, recent)
	active := create("active.txt", "uploading", &recent, recent)
	boundary := now.Add(-retention)
	atBoundary := create("boundary.txt", "uploading", &boundary, boundary)
	resumed := create("resumed.txt", "uploading", &old, old)
	// A resume request refreshes activity, not merely the creation date.
	student.request("POST", "/api/uploads", gin.H{"name": resumed.Name, "size": resumed.Size, "sha256": resumed.SHA256}, 200)
	chunked := create("chunked.txt", "uploading", &old, old)
	chunked.ChunkSize = 1024
	if err := application.DB.Model(&chunked).Update("chunk_size", 1024).Error; err != nil {
		t.Fatal(err)
	}
	student.request("PUT", fmt.Sprintf("/api/uploads/%d/chunks/0", chunked.ID), []byte("chunk"), 200)
	complete := create("complete.txt", "complete", &old, old)
	avatar := create("avatar.webp", "complete", nil, old)
	if err := application.DB.Model(&model.User{}).Where("id = ?", user.ID).Update("avatar_id", avatar.ID).Error; err != nil {
		t.Fatal(err)
	}
	blobPath := application.Config.DataDir + "/blobs/complete"
	if err := os.WriteFile(blobPath, []byte("complete"), 0600); err != nil {
		t.Fatal(err)
	}
	blob := model.Blob{SHA256: complete.SHA256, Size: 8, Path: blobPath}
	if err := application.DB.Create(&blob).Error; err != nil {
		t.Fatal(err)
	}
	removed, err := files.CleanupUploads(now, retention)
	if err != nil || removed != 3 {
		t.Fatalf("cleanup = %d, %v; want 3", removed, err)
	}
	for _, file := range []model.Attachment{expired, legacyExpired, atBoundary} {
		var count int64
		application.DB.Model(&model.Attachment{}).Where("id = ?", file.ID).Count(&count)
		if count != 0 {
			t.Fatalf("expired attachment %s remains", file.Name)
		}
		if _, err := os.Stat(files.ChunkDir(file.ID)); !os.IsNotExist(err) {
			t.Fatalf("expired chunks remain: %v", err)
		}
	}
	for _, file := range []model.Attachment{legacyActive, active, resumed, chunked, complete, avatar} {
		var saved model.Attachment
		if err := application.DB.First(&saved, file.ID).Error; err != nil {
			t.Fatalf("preserved file %s deleted: %v", file.Name, err)
		}
		if _, err := os.Stat(files.ChunkPath(file.ID, 0)); err != nil {
			t.Fatal(err)
		}
		if file.Status == "uploading" && (saved.UploadActivityAt == nil || !saved.UploadActivityAt.After(boundary)) {
			t.Fatalf("activity not refreshed for %s", file.Name)
		}
	}
	if _, err := os.Stat(blobPath); err != nil {
		t.Fatal("completed blob was removed", err)
	}
	me := decode[model.User](t, student.request("GET", "/api/me", nil, 200))
	if me.AvatarID == nil || *me.AvatarID != avatar.ID {
		t.Fatal("completed avatar association changed")
	}
	if removed, err := files.CleanupUploads(now, retention); err != nil || removed != 0 {
		t.Fatalf("second sweep = %d, %v", removed, err)
	}
}

func TestUploadCleanupWorkerAppliesSavedSettings(t *testing.T) {
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
	admin.request("POST", "/api/login", gin.H{"username": "admin", "password": "123456"}, 200)
	old := time.Now().Add(-time.Hour)
	file := model.Attachment{UserID: 1, Name: "abandoned.txt", Status: "uploading", CreatedAt: old, UploadActivityAt: &old}
	if err := application.DB.Create(&file).Error; err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(application.Handler.Files.ChunkDir(file.ID), 0700); err != nil {
		t.Fatal(err)
	}
	admin.request("PUT", "/api/admin/settings", gin.H{"upload_cleanup_value": "1", "upload_cleanup_unit": "second"}, 200)
	waitRemoved := func(id uint) {
		t.Helper()
		deadline := time.Now().Add(5 * time.Second)
		for {
			var count int64
			if err := application.DB.Model(&model.Attachment{}).Where("id = ?", id).Count(&count).Error; err != nil {
				t.Fatal(err)
			}
			if count == 0 {
				break
			}
			if time.Now().After(deadline) {
				t.Fatal("worker did not clean expired upload")
			}
			time.Sleep(10 * time.Millisecond)
		}
		if _, err := os.Stat(application.Handler.Files.ChunkDir(id)); !os.IsNotExist(err) {
			t.Fatal("worker left expired chunks", err)
		}
	}
	waitRemoved(file.ID)
	// This upload expires later, proving the periodic sweep runs without another save.
	now := time.Now()
	fresh := model.Attachment{UserID: 1, Name: "fresh.txt", Status: "uploading", CreatedAt: now, UploadActivityAt: &now}
	if err := application.DB.Create(&fresh).Error; err != nil {
		t.Fatal(err)
	}
	waitRemoved(fresh.ID)
	admin.request("PUT", "/api/admin/settings", gin.H{"upload_cleanup_value": "1", "upload_cleanup_unit": "month"}, 200)
	// Restart must preserve the saved unit and perform its startup cleanup.
	startupOld := time.Now().Add(-31 * 24 * time.Hour)
	expired := model.Attachment{UserID: 1, Name: "startup-expired.txt", Status: "uploading", CreatedAt: startupOld, UploadActivityAt: &startupOld}
	cfg := application.Config
	// Close must cancel a running worker before closing its database, and be idempotent.
	if err := application.Close(); err != nil {
		t.Fatal(err)
	}
	select {
	case <-application.cleanupDone:
	default:
		t.Fatal("cleanup worker did not stop")
	}
	// Seed the old row while the second application's worker is paused, then
	// restart once more to verify the default settings do not overwrite it.
	application, err = New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	application.cleanupCancel()
	<-application.cleanupDone
	if err := application.DB.Create(&expired).Error; err != nil {
		t.Fatal(err)
	}
	if err := application.Close(); err != nil {
		t.Fatal(err)
	}
	application, err = New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer application.Close()
	retention, err := application.Handler.Settings.UploadRetention()
	if err != nil || retention != 30*24*time.Hour {
		t.Fatalf("saved retention lost on restart: %v, %v", retention, err)
	}
	waitRemoved(expired.ID)
}
