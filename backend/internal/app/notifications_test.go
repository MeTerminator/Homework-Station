package app

import (
	"fmt"
	"testing"

	"github.com/gin-gonic/gin"

	"homework-station/internal/config"
	"homework-station/internal/model"
)

func TestNotificationPermissionsAndPreferences(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("DATA_DIR", t.TempDir())
	t.Setenv("DB_DRIVER", "sqlite")
	t.Setenv("DB_DSN", "")
	t.Setenv("JWT_SECRET", "test-session-secret-32-characters-long")
	a, err := New(config.Load())
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	admin := &testClient{t: t, r: a.Router}
	user := &testClient{t: t, r: a.Router}
	guest := &testClient{t: t, r: a.Router}
	admin.request("POST", "/api/login", gin.H{"username": "admin", "password": "123456"}, 200)
	owner := user.register("notification-user")
	if !owner.QQCommentNotifications {
		t.Fatal("new users must default to enabled")
	}
	updated := decode[model.User](t, user.request("PATCH", "/api/me", gin.H{"qq_comment_notifications": false}, 200))
	if updated.QQCommentNotifications {
		t.Fatal("opt-out did not persist")
	}
	updated = decode[model.User](t, user.request("PATCH", "/api/me", gin.H{"username": "notification-user2"}, 200))
	if updated.QQCommentNotifications {
		t.Fatal("unrelated edit reset preference")
	}
	assignment := model.Assignment{Name: "test", Status: "published"}
	a.DB.Create(&assignment)
	submission := model.Submission{UserID: owner.ID, AssignmentID: assignment.ID}
	a.DB.Create(&submission)
	review := decode[model.ReviewMessage](t, admin.request("POST", fmt.Sprintf("/api/submissions/%d/reviews", submission.ID), gin.H{"content": "feedback"}, 201))
	if !review.NotificationPending || review.LastNotifiedAt != nil {
		t.Fatal("new admin review not queued")
	}
	path := fmt.Sprintf("/api/admin/reviews/%d/notification/resend", review.ID)
	guest.request("POST", path, gin.H{}, 401)
	user.request("POST", path, gin.H{}, 403)
	admin.request("POST", path, gin.H{}, 200)
	admin.request("POST", "/api/admin/reviews/0/notification/resend", gin.H{}, 400)
	admin.request("POST", "/api/admin/reviews/999999/notification/resend", gin.H{}, 404)
	bulk := "/api/admin/settings/notifications/resend-all"
	user.request("POST", bulk, gin.H{"confirmation": "重新发送所有评论通知"}, 403)
	admin.request("POST", bulk, gin.H{}, 400)
	admin.request("POST", bulk, gin.H{"confirmation": "yes"}, 400)
	admin.request("POST", bulk, gin.H{"confirmation": "重新发送所有评论通知"}, 200)
	admin.request("PUT", "/api/admin/settings", gin.H{"qq_notification_url": "file:///tmp/test"}, 400)
	admin.request("PUT", "/api/admin/settings", gin.H{"qq_notification_secret": "short"}, 400)
	admin.request("PUT", "/api/admin/settings", gin.H{"qq_notification_url": "http://127.0.0.1:8080/homework/notifications", "qq_notification_secret": "test-notification-secret-32-characters-long"}, 200)
}
