package model

import "time"

// ReviewMessage belongs to one submission; Admin records the author's role at send time.
type ReviewMessage struct {
	LastNotifiedAt      *time.Time   `json:"last_notified_at"`
	NotificationPending bool         `json:"notification_pending" gorm:"not null;default:true;index:idx_review_queue"`
	NotificationNextAt  *time.Time   `json:"-"`
	NotificationStatus  string       `json:"notification_status"`
	ID                  uint         `json:"id" gorm:"primaryKey"`
	SubmissionID        uint         `json:"submission_id" gorm:"index;not null"`
	UserID              uint         `json:"user_id"`
	Admin               bool         `json:"admin"`
	Content             string       `json:"content" gorm:"type:text"`
	CreatedAt           time.Time    `json:"created_at"`
	Attachments         []Attachment `json:"attachments" gorm:"many2many:review_attachments"`
	Username            string       `json:"username" gorm:"-"`
	AvatarID            *uint        `json:"avatar_id" gorm:"-"`
}

type ReviewAttachment struct {
	ReviewMessageID uint `gorm:"primaryKey"`
	AttachmentID    uint `gorm:"primaryKey"`
}
