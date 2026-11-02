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
