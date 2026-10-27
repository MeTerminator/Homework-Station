package model

import "time"

type Submission struct {
	ID           uint         `json:"id" gorm:"primaryKey"`
	AssignmentID uint         `json:"assignment_id" gorm:"index"`
	UserID       uint         `json:"user_id" gorm:"index"`
	Content      string       `json:"content" gorm:"type:text"`
	Withdrawn    bool         `json:"withdrawn"`
	CreatedAt    time.Time    `json:"created_at"`
	Attachments  []Attachment `json:"attachments" gorm:"many2many:submission_attachments"`
	Username     string       `json:"username" gorm:"-"`
}

type SubmissionAttachment struct {
	SubmissionID uint `gorm:"primaryKey"`
	AttachmentID uint `gorm:"primaryKey"`
}
