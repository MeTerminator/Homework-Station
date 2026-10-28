package model

import "time"

type Assignment struct {
	ID               uint       `json:"id" gorm:"primaryKey"`
	Name             string     `json:"name" gorm:"size:200"`
	CategoryID       *uint      `json:"category_id"`
	Status           string     `json:"status" gorm:"size:20"`
	Deadline         *time.Time `json:"deadline"`
	Sort             int        `json:"sort"`
	Description      string     `json:"description" gorm:"type:text"`
	AllowAttachments bool       `json:"allow_attachments"`
	Extensions       string     `json:"extensions"`
	MaxSize          int64      `json:"max_size"`
	Multiple         bool       `json:"multiple"`
	Withdraw         bool       `json:"withdraw"`
	RequireRealName  bool       `json:"require_real_name" gorm:"not null;default:false"`
	RequireQQ        bool       `json:"require_qq" gorm:"not null;default:false"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	Submitters       int64      `json:"submitters" gorm:"-"`
	Submitted        bool       `json:"submitted" gorm:"-"`
}

type AssignmentAttachment struct {
	AssignmentID uint `gorm:"primaryKey"`
	AttachmentID uint `gorm:"primaryKey"`
}
