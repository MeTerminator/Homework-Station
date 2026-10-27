package service

import (
	"strings"
	"time"

	"gorm.io/gorm"

	"homework-station/internal/model"
)

type AssignmentService struct {
	db    *gorm.DB
	files *FileService
}

func NewAssignmentService(db *gorm.DB, files *FileService) *AssignmentService {
	return &AssignmentService{db: db, files: files}
}
func (svc *AssignmentService) Save(a model.Assignment, attachmentIDs []uint, actor model.User) (model.Assignment, error) {
	svc.files.MutationMu.Lock()
	defer svc.files.MutationMu.Unlock()
	a.CreatedAt = time.Time{}
	a.Name = strings.TrimSpace(a.Name)
	if a.Name == "" || len(a.Name) > 200 || len(a.Description) > 2*1024*1024 {
		return model.Assignment{}, Reject(400, "名称必填且不超过 200 字节，描述不超过 2MB")
	}
	if a.Status != "published" && a.Status != "hidden" && a.Status != "closed" {
		return model.Assignment{}, Reject(400, "作业状态无效")
	}
	if a.MaxSize <= 0 || a.MaxSize > 10*1024*1024*1024 {
		return model.Assignment{}, Reject(400, "附件大小范围为 1 字节至 10GB")
	}
	if a.CategoryID != nil {
		var cat model.Category
		if svc.db.First(&cat, *a.CategoryID).Error != nil {
			return model.Assignment{}, Reject(400, "分类不存在")
		}
	}
	if a.ID != 0 {
		var old model.Assignment
		if svc.db.First(&old, a.ID).Error != nil {
			return model.Assignment{}, Reject(404, "作业不存在")
		}
		a.CreatedAt = old.CreatedAt
	}
	files := []model.Attachment{}
	if len(attachmentIDs) > 0 {
		svc.db.Where("id IN ? AND status = ?", attachmentIDs, "complete").Find(&files)
		for _, file := range files {
			if file.UserID != actor.ID {
				var n int64
				svc.db.Model(&model.AssignmentAttachment{}).Where("assignment_id = ? AND attachment_id = ?", a.ID, file.ID).Count(&n)
				if n == 0 {
					return model.Assignment{}, Reject(400, "附件不属于当前用户或当前作业")
				}
			}
		}
		if len(files) != len(attachmentIDs) {
			return model.Assignment{}, Reject(400, "描述附件不属于当前用户或尚未完成")
		}
	}
	e := svc.db.Transaction(func(tx *gorm.DB) error {
		if e := tx.Save(&a).Error; e != nil {
			return e
		}
		if e := tx.Where("assignment_id = ?", a.ID).Delete(&model.AssignmentAttachment{}).Error; e != nil {
			return e
		}
		for _, f := range files {
			if e := tx.Create(&model.AssignmentAttachment{AssignmentID: a.ID, AttachmentID: f.ID}).Error; e != nil {
				return e
			}
		}
		return nil
	})
	if e != nil {
		return model.Assignment{}, Reject(500, "保存失败")
	}
	return a, nil
}
func (svc *AssignmentService) Delete(id uint) error {
	svc.files.MutationMu.Lock()
	defer svc.files.MutationMu.Unlock()
	var count int64
	svc.db.Model(&model.Submission{}).Where("assignment_id = ?", id).Count(&count)
	if count > 0 {
		return Reject(409, "请先删除此作业的历史提交")
	}
	e := svc.db.Transaction(func(tx *gorm.DB) error {
		if e := tx.Where("assignment_id = ?", id).Delete(&model.AssignmentAttachment{}).Error; e != nil {
			return e
		}
		return tx.Delete(&model.Assignment{}, id).Error
	})
	if e != nil {
		return Reject(500, "删除失败")
	}
	return nil
}
