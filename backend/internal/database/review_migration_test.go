package database

import (
	"io"
	"log"
	"path/filepath"
	"testing"
	"time"

	"homework-station/internal/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Represents the submission schema before review conversations were introduced.
type legacySubmission struct {
	ID           uint   `gorm:"primaryKey"`
	AssignmentID uint   `gorm:"index"`
	UserID       uint   `gorm:"index"`
	Content      string `gorm:"type:text"`
	Withdrawn    bool
	CreatedAt    time.Time
}

func (legacySubmission) TableName() string { return "submissions" }

func TestReviewMigrationPreservesExistingData(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "legacy.db")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	defer sqlDB.Close()
	if err := db.AutoMigrate(&model.User{}, &model.Assignment{}, &model.Attachment{}, &legacySubmission{}, &model.SubmissionAttachment{}, &model.Setting{}); err != nil {
		t.Fatal(err)
	}
	user := model.User{Username: "existing", Password: "preserved-password", Admin: true}
	a := model.Assignment{Name: "existing assignment", Status: "published"}
	db.Create(&user)
	db.Create(&a)
	db.Create(&model.Setting{Key: "site_title", Value: "Existing site"})
	f := model.Attachment{UserID: user.ID, Name: "existing.txt", Status: "complete"}
	db.Create(&f)
	old := legacySubmission{AssignmentID: a.ID, UserID: user.ID, Content: "existing answer", CreatedAt: time.Now().UTC().Truncate(time.Second)}
	db.Create(&old)
	db.Create(&model.SubmissionAttachment{SubmissionID: old.ID, AttachmentID: f.ID})
	for i := 0; i < 2; i++ {
		if err := Initialize(db, "", log.New(io.Discard, "", 0)); err != nil {
			t.Fatal(err)
		}
		var migrated model.Submission
		if err := db.Preload("Attachments").First(&migrated, old.ID).Error; err != nil {
			t.Fatal(err)
		}
		wantStatus := "submitted"
		if i > 0 {
			wantStatus = "reviewed"
		}
		if migrated.ReviewStatus != wantStatus || migrated.Content != old.Content || !migrated.CreatedAt.Equal(old.CreatedAt) || len(migrated.Attachments) != 1 {
			t.Fatal("existing submission altered", migrated)
		}
		var kept model.User
		db.First(&kept, user.ID)
		var setting model.Setting
		db.First(&setting, "key = ?", "site_title")
		if kept.Password != user.Password || setting.Value != "Existing site" {
			t.Fatal("existing account/settings overwritten")
		}
		var count int64
		db.Model(&model.ReviewMessage{}).Count(&count)
		if count != int64(i) {
			t.Fatal("migration changed reviews")
		}
		if i == 0 {
			if err := db.Create(&model.ReviewMessage{SubmissionID: old.ID, UserID: user.ID, Admin: true, Content: "preserved review"}).Error; err != nil {
				t.Fatal(err)
			}
			if err := db.Model(&model.Submission{}).Where("id = ?", old.ID).Update("review_status", "reviewed").Error; err != nil {
				t.Fatal(err)
			}
		}
	}
}
