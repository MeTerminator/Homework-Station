package handler

import (
	"homework-station/internal/service"

	"github.com/gin-gonic/gin"
)

func (h *Handler) ListReviews(c *gin.Context) {
	messages, err := h.Submissions.Reviews(current(c), parseID(c.Param("id")))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(200, messages)
}

func (h *Handler) CreateReview(c *gin.Context) {
	var input service.SubmissionInput
	if c.ShouldBindJSON(&input) != nil {
		fail(c, 400, "评价格式错误")
		return
	}
	message, err := h.Submissions.CreateReview(current(c), parseID(c.Param("id")), input)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(201, message)
}

func (h *Handler) ResendReviewNotification(c *gin.Context) {
	id := parseID(c.Param("id"))
	if id == 0 {
		fail(c, 400, "评论 ID 无效")
		return
	}
	n, err := h.Notifications.Requeue(id)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(200, gin.H{"queued": n})
}
func (h *Handler) ResendAllReviewNotifications(c *gin.Context) {
	var input struct {
		Confirmation string `json:"confirmation"`
	}
	if c.ShouldBindJSON(&input) != nil || input.Confirmation != "重新发送所有评论通知" {
		fail(c, 400, "请输入确认文字：重新发送所有评论通知")
		return
	}
	n, err := h.Notifications.Requeue(0)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(200, gin.H{"queued": n})
}
