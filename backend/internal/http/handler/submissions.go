package handler

import (
	"github.com/gin-gonic/gin"

	"homework-station/internal/model"
	"homework-station/internal/service"
)

func (h *Handler) ListSubmissions(c *gin.Context) {
	a, ok := h.VisibleAssignment(c)
	if !ok {
		return
	}
	u := current(c)
	q := h.DB.Preload("Attachments").Where("assignment_id = ?", a.ID)
	if !u.Admin {
		q = q.Where("user_id = ?", u.ID)
	}
	if id := parseID(c.Query("user_id")); u.Admin && id > 0 {
		q = q.Where("user_id = ?", id)
	}
	v := []model.Submission{}
	q.Order("id desc").Find(&v)
	for i := range v {
		var user model.User
		h.DB.First(&user, v[i].UserID)
		v[i].Username = user.Username
	}
	c.JSON(200, v)
}

// Serializes policy checks and submission writes with upload completion/deletion in this process.
func (h *Handler) CreateSubmission(c *gin.Context) {
	var input service.SubmissionInput
	if c.ShouldBindJSON(&input) != nil {
		fail(c, 400, "提交格式错误或文本超过 2MB")
		return
	}
	submission, err := h.Submissions.Create(current(c), parseID(c.Param("id")), input)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(201, submission)
}
func (h *Handler) WithdrawSubmission(c *gin.Context) {
	if err := h.Submissions.Withdraw(current(c), parseID(c.Param("id"))); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(200, gin.H{"ok": true})
}
func (h *Handler) DeleteSubmission(c *gin.Context) {
	if err := h.Submissions.Delete(parseID(c.Param("id")), c.Query("files") == "true"); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(200, gin.H{"ok": true})
}
