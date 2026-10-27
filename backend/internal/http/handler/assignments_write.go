package handler

import (
	"github.com/gin-gonic/gin"

	"homework-station/internal/model"
)

func (h *Handler) SaveAssignment(c *gin.Context) {
	var input struct {
		model.Assignment
		AttachmentIDs []uint `json:"attachment_ids"`
	}
	if c.ShouldBindJSON(&input) != nil {
		fail(c, 400, "请求格式不正确")
		return
	}
	input.ID = parseID(c.Param("id"))
	assignment, err := h.Assignments.Save(input.Assignment, input.AttachmentIDs, current(c))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(200, assignment)
}
func (h *Handler) DeleteAssignment(c *gin.Context) {
	if err := h.Assignments.Delete(parseID(c.Param("id"))); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(200, gin.H{"ok": true})
}
