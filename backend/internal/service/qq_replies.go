package service

import (
	"regexp"

	"homework-station/internal/model"
)

type QQReplyInput struct {
	BotID          string `json:"bot_id"`
	QQ             string `json:"qq"`
	MessageID      string `json:"message_id"`
	ReplyMessageID string `json:"reply_message_id"`
	Content        string `json:"content"`
}

func (svc *SubmissionService) CreateQQReply(v QQReplyInput) (model.ReviewMessage, error) {
	var empty model.ReviewMessage
	qqPattern := regexp.MustCompile(`^[1-9][0-9]{4,14}$`)
	messagePattern := regexp.MustCompile(`^-?[0-9]{1,20}$`)
	if !qqPattern.MatchString(v.BotID) || !qqPattern.MatchString(v.QQ) || !messagePattern.MatchString(v.MessageID) || !messagePattern.MatchString(v.ReplyMessageID) {
		return empty, Reject(400, "QQ 回复参数无效")
	}
	var link model.QQNotification
	if err := svc.db.Where("bot_id = ? AND message_id = ? AND qq = ?", v.BotID, v.ReplyMessageID, v.QQ).First(&link).Error; err != nil {
		return empty, Reject(404, "引用的消息不是可回复的作业评论，请在平台上回复")
	}
	var review model.ReviewMessage
	if err := svc.db.First(&review, link.ReviewID).Error; err != nil || !review.Admin {
		return empty, Reject(404, "原评论不存在")
	}
	var user model.User
	if err := svc.db.Where("qq = ?", v.QQ).First(&user).Error; err != nil {
		return empty, Reject(403, "此 QQ 尚未绑定平台账户")
	}
	var submission model.Submission
	if err := svc.db.First(&submission, review.SubmissionID).Error; err != nil || submission.UserID != user.ID {
		return empty, Reject(403, "只能回复本人提交的作业评论")
	}
	key := v.BotID + ":" + v.QQ + ":" + v.MessageID
	return svc.createReview(user, review.SubmissionID, SubmissionInput{Content: v.Content}, &key)
}
