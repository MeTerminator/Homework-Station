package service

import (
	"strings"
	"time"

	"gorm.io/gorm"

	"homework-station/internal/model"
)

type SubmissionInput struct {
	Content       string `json:"content"`
	AttachmentIDs []uint `json:"attachment_ids"`
}
type SubmissionService struct {
	db    *gorm.DB
	files *FileService
}

func NewSubmissionService(db *gorm.DB, files *FileService) *SubmissionService {
	return &SubmissionService{db: db, files: files}
}
func (svc *SubmissionService) Create(u model.User, assignmentID uint, v SubmissionInput) (model.Submission, error) {
	svc.files.MutationMu.Lock()
	defer svc.files.MutationMu.Unlock()
	var a model.Assignment
	if svc.db.First(&a, assignmentID).Error != nil || (!u.Admin && a.Status == "hidden") {
		return model.Submission{}, Reject(404, "作业不存在")
	}
	if len(v.Content) > 2*1024*1024 {
		return model.Submission{}, Reject(400, "提交格式错误或文本超过 2MB")
	}
	if a.Status != "published" || (a.Deadline != nil && !a.Deadline.After(time.Now())) {
		return model.Submission{}, Reject(409, "作业尚未发布或已截止")
	}
	if a.RequireRealName && (u.StudentNo == nil || strings.TrimSpace(*u.StudentNo) == "" || strings.TrimSpace(u.RealName) == "") {
		return model.Submission{}, Reject(403, "此作业需要完成实名认证后才能提交")
	}
	if a.RequireQQ && (u.QQ == nil || strings.TrimSpace(*u.QQ) == "") {
		return model.Submission{}, Reject(403, "此作业需要绑定 QQ 后才能提交")
	}
	if strings.TrimSpace(v.Content) == "" && len(v.AttachmentIDs) == 0 {
		return model.Submission{}, Reject(400, "请填写内容或上传附件")
	}
	if !a.Multiple {
		var n int64
		svc.db.Model(&model.Submission{}).Where("assignment_id = ? AND user_id = ? AND withdrawn = ?", a.ID, u.ID, false).Count(&n)
		if n > 0 {
			return model.Submission{}, Reject(409, "此作业不允许重复提交")
		}
	}
	if !a.AllowAttachments && len(v.AttachmentIDs) > 0 {
		return model.Submission{}, Reject(400, "此作业只允许文本提交")
	}
	files := []model.Attachment{}
	if len(v.AttachmentIDs) > 0 {
		svc.db.Where("id IN ? AND user_id = ? AND status = ?", v.AttachmentIDs, u.ID, "complete").Find(&files)
		if len(files) != len(v.AttachmentIDs) {
			return model.Submission{}, Reject(400, "附件无效或尚未上传完成")
		}
		for _, f := range files {
			if f.Size > a.MaxSize || !AllowedExtension(f.Name, a.Extensions) {
				return model.Submission{}, Reject(400, "附件大小或后缀不符合作业要求")
			}
		}
	}
	s := model.Submission{AssignmentID: a.ID, UserID: u.ID, Content: v.Content, Attachments: files}
	if svc.db.Create(&s).Error != nil {
		return model.Submission{}, Reject(500, "提交失败")
	}
	return s, nil
}
func (svc *SubmissionService) Withdraw(u model.User, id uint) error {
	svc.files.MutationMu.Lock()
	defer svc.files.MutationMu.Unlock()
	var s model.Submission
	if svc.db.First(&s, id).Error != nil || s.UserID != u.ID {
		return Reject(404, "提交记录不存在")
	}
	var a model.Assignment
	svc.db.First(&a, s.AssignmentID)
	if !a.Withdraw {
		return Reject(403, "此作业不允许撤回")
	}
	if s.Withdrawn {
		return Reject(409, "已撤回")
	}
	if svc.db.Model(&s).Update("withdrawn", true).Error != nil {
		return Reject(500, "撤回失败")
	}
	return nil
}
func (svc *SubmissionService) Delete(id uint, deleteFiles bool) error {
	svc.files.MutationMu.Lock()
	defer svc.files.MutationMu.Unlock()
	var s model.Submission
	if svc.db.Preload("Attachments").First(&s, id).Error != nil {
		return Reject(404, "提交记录不存在")
	}
	e := svc.db.Transaction(func(tx *gorm.DB) error {
		if e := tx.Where("submission_id = ?", id).Delete(&model.SubmissionAttachment{}).Error; e != nil {
			return e
		}
		return tx.Delete(&model.Submission{}, id).Error
	})
	if e != nil {
		return Reject(500, "删除失败")
	}
	if deleteFiles {
		for _, f := range s.Attachments {
			if len(svc.files.AttachmentUsage(f.ID)) == 0 {
				svc.files.RemoveAttachment(f)
			}
		}
	}
	return nil
}
