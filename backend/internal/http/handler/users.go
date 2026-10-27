package handler

import (
	"strings"

	"github.com/gin-gonic/gin"

	"homework-station/internal/model"
	"homework-station/internal/service"
)

func (h *Handler) ListUsers(c *gin.Context) {
	page, size := pagination(c)
	q := h.DB.Model(&model.User{})
	if s := c.Query("search"); s != "" {
		q = q.Where("LOWER(username) LIKE ?", "%"+strings.ToLower(s)+"%")
	}
	var n int64
	q.Count(&n)
	v := []model.User{}
	q.Order("id desc").Offset((page - 1) * size).Limit(size).Find(&v)
	c.JSON(200, gin.H{"items": v, "total": n})
}
func (h *Handler) GetUser(c *gin.Context) {
	var u model.User
	if h.DB.First(&u, parseID(c.Param("id"))).Error != nil {
		fail(c, 404, "用户不存在")
		return
	}
	subs := []model.Submission{}
	h.DB.Preload("Attachments").Where("user_id = ?", u.ID).Order("id desc").Find(&subs)
	files := []model.Attachment{}
	h.DB.Where("user_id = ?", u.ID).Order("id desc").Find(&files)
	for i := range files {
		files[i].Usage = h.Files.AttachmentUsage(files[i].ID)
	}
	c.JSON(200, gin.H{"user": u, "submissions": subs, "attachments": files})
}
func (h *Handler) UpdateUser(c *gin.Context) {
	var input service.AdminUserInput
	if c.ShouldBindJSON(&input) != nil {
		fail(c, 400, "请求格式不正确")
		return
	}
	user, err := h.Users.UpdateAdmin(parseID(c.Param("id")), input)
	if err != nil {
		respondError(c, err)
		return
	}
	if user.ID == current(c).ID {
		h.IssueSession(c, user)
	}
	c.JSON(200, user)
}
func (h *Handler) Unbind(c *gin.Context) {
	if err := h.Users.Unbind(parseID(c.Param("id")), c.Param("kind")); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(200, gin.H{"ok": true})
}
