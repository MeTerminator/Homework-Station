package service

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"sync"

	"homework-station/internal/model"

	"gorm.io/gorm"
)

const ChunkBytes int64 = 1024 * 1024

type FileService struct {
	db         *gorm.DB
	dataDir    string
	MutationMu sync.Mutex
}

func NewFileService(db *gorm.DB, dataDir string) *FileService {
	return &FileService{db: db, dataDir: dataDir}
}

func (s *FileService) ChunkDir(id uint) string {
	return filepath.Join(s.dataDir, "chunks", fmt.Sprint(id))
}
func (s *FileService) ChunkPath(id uint, i int) string {
	return filepath.Join(s.ChunkDir(id), strconv.Itoa(i))
}
func (s *FileService) Received(f model.Attachment) []int {
	v := []int{}
	entries, _ := os.ReadDir(s.ChunkDir(f.ID))
	for _, e := range entries {
		n, err := strconv.Atoi(e.Name())
		if err == nil && n >= 0 && n < f.TotalChunks {
			v = append(v, n)
		}
	}
	sort.Ints(v)
	return v
}
func (s *FileService) AttachmentUsage(id uint) []string {
	v := []string{}
	users := []model.User{}
	s.db.Where("avatar_id = ?", id).Find(&users)
	for _, u := range users {
		v = append(v, fmt.Sprintf("用户 #%d %s 的头像", u.ID, u.Username))
	}
	subs := []model.Submission{}
	s.db.Joins("JOIN submission_attachments ON submission_attachments.submission_id = submissions.id").Where("submission_attachments.attachment_id = ?", id).Find(&subs)
	for _, s := range subs {
		v = append(v, fmt.Sprintf("作业 #%d · 提交 #%d", s.AssignmentID, s.ID))
	}
	links := []model.AssignmentAttachment{}
	s.db.Where("attachment_id = ?", id).Find(&links)
	for _, l := range links {
		v = append(v, fmt.Sprintf("作业 #%d 的描述", l.AssignmentID))
	}
	return v
}
func (s *FileService) CanDownload(u model.User, f model.Attachment) bool {
	if u.Admin || f.UserID == u.ID {
		return true
	}
	var n int64
	s.db.Model(&model.Assignment{}).Joins("JOIN assignment_attachments ON assignment_attachments.assignment_id = assignments.id").Where("assignment_attachments.attachment_id = ? AND assignments.status <> ?", f.ID, "hidden").Count(&n)
	if n > 0 {
		return true
	}
	s.db.Model(&model.User{}).Where("avatar_id = ?", f.ID).Count(&n)
	return n > 0
}
func (s *FileService) RemoveAttachment(f model.Attachment) error {
	e := s.db.Transaction(func(tx *gorm.DB) error {
		if e := tx.Model(&model.User{}).Where("avatar_id = ?", f.ID).Update("avatar_id", nil).Error; e != nil {
			return e
		}
		if e := tx.Where("attachment_id = ?", f.ID).Delete(&model.SubmissionAttachment{}).Error; e != nil {
			return e
		}
		if e := tx.Where("attachment_id = ?", f.ID).Delete(&model.AssignmentAttachment{}).Error; e != nil {
			return e
		}
		return tx.Delete(&f).Error
	})
	if e != nil {
		return e
	}
	os.RemoveAll(s.ChunkDir(f.ID))
	var n int64
	s.db.Model(&model.Attachment{}).Where("sha256 = ? AND status = ?", f.SHA256, "complete").Count(&n)
	if n == 0 {
		var b model.Blob
		if s.db.First(&b, "sha256 = ?", f.SHA256).Error == nil {
			if e := os.Remove(b.Path); e != nil && !os.IsNotExist(e) {
				return e
			}
			return s.db.Delete(&b).Error
		}
	}
	return nil
}
