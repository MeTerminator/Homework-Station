package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"homework-station/internal/model"
)

func TestNotificationQueue(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "queue.db")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	defer sqlDB.Close()
	if err := db.AutoMigrate(&model.User{}, &model.Submission{}, &model.ReviewMessage{}, &model.Setting{}); err != nil {
		t.Fatal(err)
	}
	qq := "123456789"
	owner := model.User{Username: "owner", QQ: &qq}
	staff := model.User{Username: "staff", Admin: true}
	db.Create(&owner)
	db.Create(&staff)
	submission := model.Submission{UserID: owner.ID}
	db.Create(&submission)
	created := time.Date(2026, 11, 2, 0, 0, 0, 0, time.UTC)
	review := model.ReviewMessage{SubmissionID: submission.ID, UserID: staff.ID, Admin: true, Content: "请修改 [CQ:at,qq=all]", CreatedAt: created}
	db.Create(&review)
	reply := model.ReviewMessage{SubmissionID: submission.ID, UserID: owner.ID, Content: "reply"}
	db.Create(&reply)
	calls := 0
	status := "sent"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Header.Get("Authorization") != "Bearer "+strings.Repeat("s", 32) {
			t.Error("missing auth")
		}
		var body struct {
			QQ, Message string
			ReviewID    uint `json:"review_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		if body.QQ != qq || body.ReviewID != review.ID || body.Message != "# 作业评论\nstaff\n请修改 [CQ:at,qq=all]\n\n回复此消息可在平台上进行回复\n2026-11-02 08:00:00" {
			t.Errorf("unexpected payload: %+v", body)
		}
		if status == "error" {
			w.WriteHeader(503)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"status": status})
	}))
	defer server.Close()
	settings := NewSettingsService(db)
	db.Create(&model.Setting{Key: "qq_notification_url", Value: server.URL})
	db.Create(&model.Setting{Key: "qq_notification_secret", Value: strings.Repeat("s", 32)})
	queue := NewNotificationService(db, settings)
	process := func() model.ReviewMessage {
		t.Helper()
		if err := queue.ProcessNext(context.Background(), false); err != nil {
			t.Fatal(err)
		}
		var m model.ReviewMessage
		db.First(&m, review.ID)
		return m
	}
	// Cancellation during the randomized wait must preserve the durable pending row.
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	if err := queue.ProcessNext(cancelled, true); err != context.Canceled {
		t.Fatal("worker did not stop", err)
	}
	if calls != 0 {
		t.Fatal("cancelled worker sent a notification")
	}
	m := process()
	if calls != 1 || m.LastNotifiedAt == nil || m.NotificationPending || m.NotificationStatus != "sent" {
		t.Fatal("not sent", m, calls)
	}
	last := *m.LastNotifiedAt
	process()
	if calls != 1 {
		t.Fatal("duplicate delivery")
	}
	if _, err := queue.Requeue(reply.ID); err == nil {
		t.Fatal("student reply requeued")
	}
	if _, err := queue.Requeue(review.ID); err != nil {
		t.Fatal(err)
	}
	status = "not_friend"
	m = process()
	if calls != 2 || m.NotificationPending || m.NotificationStatus != "skipped_not_friend" || !m.LastNotifiedAt.Equal(last) {
		t.Fatal("friend skip changed timestamp", m)
	}
	queue.Requeue(review.ID)
	db.Model(&owner).Update("qq_comment_notifications", false)
	m = process()
	if calls != 2 || m.NotificationStatus != "skipped_disabled" {
		t.Fatal("disabled recipient contacted", m)
	}
	queue.Requeue(review.ID)
	db.Model(&owner).Updates(map[string]interface{}{"qq_comment_notifications": true, "qq": nil})
	m = process()
	if calls != 2 || m.NotificationStatus != "skipped_unbound" {
		t.Fatal("unbound recipient contacted", m)
	}
	db.Model(&owner).Update("qq", qq)
	queue.Requeue(review.ID)
	status = "error"
	m = process()
	if calls != 3 || !m.NotificationPending || m.NotificationStatus != "retry" || m.NotificationNextAt == nil {
		t.Fatal("failure not scheduled", m)
	}
	process()
	if calls != 3 {
		t.Fatal("retried too soon")
	}
	// Recreate the service to prove queue state survives process restarts.
	queue = NewNotificationService(db, settings)
	status = "sent"
	db.Model(&review).Update("notification_next_at", time.Now().Add(-time.Second))
	m = process()
	if calls != 4 || m.NotificationPending || !m.LastNotifiedAt.After(last) {
		t.Fatal("retry not completed", m)
	}
	if n, err := queue.Requeue(0); err != nil || n != 1 {
		t.Fatal("bulk requeue includes student replies", n, err)
	}
}
