package model

import "time"

type User struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Username    string    `json:"username" gorm:"uniqueIndex;size:64;not null"`
	Password    string    `json:"-"`
	Admin       bool      `json:"admin"`
	AuthVersion uint      `json:"-"`
	AvatarID    *uint     `json:"avatar_id"`
	StudentNo   *string   `json:"student_no" gorm:"uniqueIndex;size:128"`
	RealName    string    `json:"real_name"`
	QQ          *string   `json:"qq" gorm:"uniqueIndex;size:20"`
	CategoryIDs string    `json:"category_ids"`
	CreatedAt   time.Time `json:"created_at"`
}
