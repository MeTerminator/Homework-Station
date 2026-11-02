package handler

import (
	"crypto/subtle"
	"net/http"

	"github.com/gin-gonic/gin"

	"homework-station/internal/service"
)

func (h *Handler) CreateQQReply(c *gin.Context) {
	secret := h.Settings.Get("qq_notification_secret")
	if secret == "" || subtle.ConstantTimeCompare([]byte(c.GetHeader("Authorization")), []byte("Bearer "+secret)) != 1 {
		fail(c, 401, "通知凭据无效")
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 3*1024*1024)
	var input service.QQReplyInput
	if c.ShouldBindJSON(&input) != nil {
		fail(c, 400, "QQ 回复格式错误")
		return
	}
	message, err := h.Submissions.CreateQQReply(input)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(201, gin.H{"id": message.ID, "username": message.Username})
}
