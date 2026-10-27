package handler

import (
	"github.com/gin-gonic/gin"

	"homework-station/internal/service"
)

func (h *Handler) Me(c *gin.Context) { c.JSON(200, current(c)) }
func (h *Handler) UpdateMe(c *gin.Context) {
	var input service.ProfileInput
	if c.ShouldBindJSON(&input) != nil {
		fail(c, 400, "请求格式不正确")
		return
	}
	user, err := h.Users.UpdateProfile(current(c), input)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(200, user)
}
func (h *Handler) ChangePassword(c *gin.Context) {
	var input service.PasswordInput
	if c.ShouldBindJSON(&input) != nil {
		fail(c, 400, "旧密码不正确")
		return
	}
	user, err := h.Users.ChangePassword(current(c), input)
	if err != nil {
		respondError(c, err)
		return
	}
	h.IssueSession(c, user)
	c.JSON(200, gin.H{"ok": true})
}
