package model

import "time"

// ReviewMessage belongs to one submission; Admin records the author's role at send time.
type ReviewMessage struct {
	QQReplyKey          *string      `json:"-" gorm:"uniqueIndex;size:128"`
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

// QQNotification links a delivered private message to its original review.
type QQNotification struct {
	BotID     string `gorm:"primaryKey;size:20"`
	MessageID string `gorm:"primaryKey;size:32"`
	QQ        string `gorm:"not null;size:20"`
	ReviewID  uint   `gorm:"not null;index"`
}

type ReviewAttachment struct {
	ReviewMessageID uint `gorm:"primaryKey"`
	AttachmentID    uint `gorm:"primaryKey"`
}
