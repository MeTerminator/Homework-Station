package service

import (
	"time"

	"gorm.io/gorm"

	"homework-station/internal/model"
)

type OverviewService struct{ db *gorm.DB }

func NewOverviewService(db *gorm.DB) *OverviewService { return &OverviewService{db: db} }
func (svc *OverviewService) Get() map[string]any {
	var users, attachments, submissions int64
	var capacity int64
	svc.db.Model(&model.User{}).Count(&users)
	svc.db.Model(&model.Attachment{}).Count(&attachments)
	svc.db.Model(&model.Submission{}).Count(&submissions)
	svc.db.Model(&model.Blob{}).Select("COALESCE(SUM(size),0)").Scan(&capacity)
	type Day struct {
		Date        string `json:"date"`
		Users       int64  `json:"users"`
		Attachments int64  `json:"attachments"`
		Capacity    int64  `json:"capacity"`
		Submissions int64  `json:"submissions"`
	}
	days := []Day{}
	now := time.Now().UTC()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, -29)
	for i := 0; i < 30; i++ {
		d := Day{Date: start.AddDate(0, 0, i).Format("2006-01-02")}
		days = append(days, d)
	}
	index := func(t time.Time) int { return int(t.UTC().Sub(start).Hours() / 24) }
	us := []model.User{}
	fs := []model.Attachment{}
	ss := []model.Submission{}
	svc.db.Where("created_at >= ?", start).Find(&us)
	svc.db.Where("created_at >= ?", start).Find(&fs)
	svc.db.Where("created_at >= ?", start).Find(&ss)
	for _, u := range us {
		if i := index(u.CreatedAt); i >= 0 && i < 30 {
			days[i].Users++
		}
	}
	for _, f := range fs {
		if i := index(f.CreatedAt); i >= 0 && i < 30 {
			days[i].Attachments++
			if f.Status == "complete" {
				days[i].Capacity += f.Size
			}
		}
	}
	for _, s := range ss {
		if i := index(s.CreatedAt); i >= 0 && i < 30 {
			days[i].Submissions++
		}
	}
	return map[string]any{"users": users, "attachments": attachments, "capacity": capacity, "submissions": submissions, "days": days, "timezone": "UTC", "capacity_note": "总容量为去重后的实际文件容量；日曲线为当天新增完成附件的逻辑大小"}
}
