package handler

import (
	"github.com/gin-gonic/gin"
)

func (h *Handler) GetSettings(c *gin.Context) { c.JSON(200, h.Settings.All()) }
func (h *Handler) PutSettings(c *gin.Context) {
	var values map[string]string
	if c.ShouldBindJSON(&values) != nil {
		fail(c, 400, "请求格式不正确")
		return
	}
	if err := h.Settings.Update(values); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(200, gin.H{"ok": true})
}
