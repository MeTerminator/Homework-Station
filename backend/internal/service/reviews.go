package service

import (
	"strconv"
	"strings"

	"homework-station/internal/model"

	"gorm.io/gorm"
)

func (svc *SubmissionService) reviewSubmission(u model.User, id uint) (model.Submission, error) {
	var s model.Submission
	if svc.db.First(&s, id).Error != nil || (!u.Admin && s.UserID != u.ID) {
		return s, Reject(404, "提交记录不存在")
	}
	var a model.Assignment
	if svc.db.First(&a, s.AssignmentID).Error != nil || (!u.Admin && a.Status == "hidden") {
		return s, Reject(404, "作业不存在")
	}
	return s, nil
}

func (svc *SubmissionService) Reviews(u model.User, id uint) ([]model.ReviewMessage, error) {
	if _, err := svc.reviewSubmission(u, id); err != nil {
		return nil, err
	}
	messages := []model.ReviewMessage{}
	if err := svc.db.Preload("Attachments").Where("submission_id = ?", id).Order("id asc").Find(&messages).Error; err != nil {
		return nil, Reject(500, "加载评价失败")
	}
	ids := []uint{}
	for _, m := range messages {
		ids = append(ids, m.UserID)
	}
	if len(ids) > 0 {
		users := []model.User{}
		if err := svc.db.Select("id", "username", "avatar_id").Where("id IN ?", ids).Find(&users).Error; err != nil {
			return nil, Reject(500, "加载评价作者失败")
		}
		authors := map[uint]model.User{}
		for _, author := range users {
			authors[author.ID] = author
		}
		for i := range messages {
			author := authors[messages[i].UserID]
			messages[i].Username = author.Username
			messages[i].AvatarID = author.AvatarID
		}
	}
	return messages, nil
}

func (svc *SubmissionService) CreateReview(u model.User, id uint, v SubmissionInput) (model.ReviewMessage, error) {
	svc.files.MutationMu.Lock()
	defer svc.files.MutationMu.Unlock()
	var empty model.ReviewMessage
	s, err := svc.reviewSubmission(u, id)
	if err != nil {
		return empty, err
	}
	if s.Withdrawn {
		return empty, Reject(409, "已撤回的提交不能新增评价或回复")
	}
	if len(v.Content) > 2*1024*1024 || (strings.TrimSpace(v.Content) == "" && len(v.AttachmentIDs) == 0) {
		return empty, Reject(400, "请填写内容或上传附件，文本不能超过 2MB")
	}
	if !u.Admin && s.ReviewStatus == "submitted" {
		return empty, Reject(409, "请等待管理员评价后再回复")
	}
	files := []model.Attachment{}
	if len(v.AttachmentIDs) > 0 {
		if err := svc.db.Where("id IN ? AND user_id = ? AND status = ?", v.AttachmentIDs, u.ID, "complete").Find(&files).Error; err != nil {
			return empty, Reject(500, "加载附件失败")
		}
		if len(files) != len(v.AttachmentIDs) {
			return empty, Reject(400, "附件无效或尚未上传完成")
		}
		if !u.Admin {
			var max int64
			var setting model.Setting
			if svc.db.First(&setting, "key = ?", "default_max_size").Error != nil {
				return empty, Reject(500, "加载附件限制失败")
			}
			// Replies use site upload rules, independently of assignment submission rules.
			max, _ = strconv.ParseInt(setting.Value, 10, 64)
			var ext model.Setting
			if svc.db.First(&ext, "key = ?", "default_extensions").Error != nil {
				return empty, Reject(500, "加载附件限制失败")
			}
			for _, f := range files {
				if f.Size > max || !AllowedExtension(f.Name, ext.Value) {
					return empty, Reject(400, "附件超出站点上传限制")
				}
			}
		}
	}
	m := model.ReviewMessage{SubmissionID: id, UserID: u.ID, Admin: u.Admin, Content: v.Content, Attachments: files}
	status := "pending"
	if u.Admin {
		status = "reviewed"
	}
	if err := svc.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&m).Error; err != nil {
			return err
		}
		if !u.Admin {
			if err := tx.Model(&m).Update("notification_pending", false).Error; err != nil {
				return err
			}
			m.NotificationPending = false
		}
		return tx.Model(&s).Update("review_status", status).Error
	}); err != nil {
		return empty, Reject(500, "发送评价失败")
	}
	m.Username, m.AvatarID = u.Username, u.AvatarID
	return m, nil
}
