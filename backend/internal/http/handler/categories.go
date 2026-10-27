package handler

import (
	"github.com/gin-gonic/gin"

	"homework-station/internal/model"
)

func (h *Handler) ListCategories(c *gin.Context) {
	items, err := h.Categories.List()
	if err != nil {
		fail(c, 500, "读取分类失败")
		return
	}
	c.JSON(200, items)
}
func (h *Handler) SaveCategory(c *gin.Context) {
	var input model.Category
	if c.ShouldBindJSON(&input) != nil {
		fail(c, 400, "请求格式不正确")
		return
	}
	input.ID = parseID(c.Param("id"))
	category, err := h.Categories.Save(input)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(200, category)
}
func (h *Handler) DeleteCategory(c *gin.Context) {
	if err := h.Categories.Delete(parseID(c.Param("id"))); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(200, gin.H{"ok": true})
}
