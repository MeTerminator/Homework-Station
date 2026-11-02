package database

import (
	"io"
	"log"
	"path/filepath"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"homework-station/internal/model"
)

type oldNotificationUser struct {
	ID       uint `gorm:"primaryKey"`
	Username string
}

func (oldNotificationUser) TableName() string { return "users" }

type oldNotificationReview struct {
	SubmissionID uint `gorm:"index;not null"`
	UserID       uint
	ID           uint `gorm:"primaryKey"`
	Admin        bool
	Content      string
	CreatedAt    time.Time
}

func (oldNotificationReview) TableName() string { return "review_messages" }
func TestNotificationMigration(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "legacy.db")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	defer sqlDB.Close()
	if err := db.AutoMigrate(&oldNotificationUser{}, &oldNotificationReview{}); err != nil {
		t.Fatal(err)
	}
	db.Create(&oldNotificationUser{ID: 1, Username: "legacy"})
	db.Create(&oldNotificationReview{ID: 1, Admin: true, Content: "historical feedback", CreatedAt: time.Now()})
	for i := 0; i < 2; i++ {
		if err := Initialize(db, "", log.New(io.Discard, "", 0)); err != nil {
			t.Fatal(err)
		}
		var user model.User
		var review model.ReviewMessage
		db.First(&user, 1)
		db.First(&review, 1)
		if !user.QQCommentNotifications || review.LastNotifiedAt != nil || !review.NotificationPending || review.Content != "historical feedback" {
			t.Fatal("legacy defaults incorrect", user, review)
		}
	}
	db.Model(&model.User{}).Where("id = ?", 1).Update("qq_comment_notifications", false)
	now := time.Now().UTC()
	db.Model(&model.ReviewMessage{}).Where("id = ?", 1).Updates(map[string]interface{}{"last_notified_at": now, "notification_pending": false})
	if err := Initialize(db, "", log.New(io.Discard, "", 0)); err != nil {
		t.Fatal(err)
	}
	var user model.User
	var review model.ReviewMessage
	db.First(&user, 1)
	db.First(&review, 1)
	if user.QQCommentNotifications || review.LastNotifiedAt == nil || review.NotificationPending {
		t.Fatal("migration reset notification state")
	}
}
