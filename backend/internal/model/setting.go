package model

import "time"

type Setting struct {
	Key   string `gorm:"primaryKey;size:100"`
	Value string `gorm:"type:text"`
}

type UsedCode struct {
	ID        string `gorm:"primaryKey;size:64"`
	ExpiresAt time.Time
}
