package model

type Category struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	Name     string `json:"name" gorm:"size:100"`
	ParentID *uint  `json:"parent_id"`
	Sort     int    `json:"sort"`
}
