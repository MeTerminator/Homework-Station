package model

import "time"

type Blob struct {
	SHA256 string `json:"sha256" gorm:"primaryKey;size:64"`
	Size   int64  `json:"size"`
	Path   string `json:"-"`
}

type Attachment struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	UserID      uint      `json:"user_id" gorm:"index"`
	Name        string    `json:"name" gorm:"size:255"`
	SHA256      string    `json:"sha256" gorm:"index;size:64"`
	Size        int64     `json:"size"`
	Status      string    `json:"status" gorm:"size:20"`
	ChunkSize   int64     `json:"chunk_size"`
	TotalChunks int       `json:"total_chunks"`
	CreatedAt   time.Time `json:"created_at"`
	Usage       []string  `json:"usage" gorm:"-"`
	Received    []int     `json:"received" gorm:"-"`
}
