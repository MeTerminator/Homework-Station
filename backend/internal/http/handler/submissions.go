package handler

import (
	"github.com/gin-gonic/gin"

	"homework-station/internal/model"
	"homework-station/internal/service"
)

func (h *Handler) ListSubmissions(c *gin.Context) {
	h.listAssignmentSubmissions(c, false)
}

func (h *Handler) ListAdminAssignmentSubmissions(c *gin.Context) {
	h.listAssignmentSubmissions(c, true)
}

func (h *Handler) listAssignmentSubmissions(c *gin.Context, admin bool) {
	a, ok := h.VisibleAssignment(c)
	if !ok {
		return
	}
	u := current(c)
	q := h.DB.Preload("Attachments").Where("assignment_id = ?", a.ID)
	if !admin {
		q = q.Where("user_id = ?", u.ID)
	}
	if id := parseID(c.Query("user_id")); admin && id > 0 {
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

// ListAllSubmissions returns recent submissions across every assignment for admins.
func (h *Handler) ListAllSubmissions(c *gin.Context) {
	page, size := pagination(c)
	q := h.DB.Model(&model.Submission{})
	var total int64
	if err := q.Count(&total).Error; err != nil {
		fail(c, 500, "加载提交失败")
		return
	}
	items := []model.Submission{}
	if err := q.Select("submissions.*").
		Preload("Attachments").Order("submissions.created_at DESC, submissions.id DESC").
		Offset((page - 1) * size).Limit(size).Find(&items).Error; err != nil {
		fail(c, 500, "加载提交失败")
		return
	}
	userIDs, assignmentIDs := []uint{}, []uint{}
	for _, s := range items {
		userIDs = append(userIDs, s.UserID)
		assignmentIDs = append(assignmentIDs, s.AssignmentID)
	}
	users, assignments := []model.User{}, []model.Assignment{}
	if len(items) > 0 {
		if err := h.DB.Select("id", "username").Where("id IN ?", userIDs).Find(&users).Error; err != nil {
			fail(c, 500, "加载用户失败")
			return
		}
		if err := h.DB.Select("id", "name").Where("id IN ?", assignmentIDs).Find(&assignments).Error; err != nil {
			fail(c, 500, "加载作业失败")
			return
		}
	}
	names, titles := map[uint]string{}, map[uint]string{}
	for _, u := range users {
		names[u.ID] = u.Username
	}
	for _, a := range assignments {
		titles[a.ID] = a.Name
	}
	for i := range items {
		items[i].Username = names[items[i].UserID]
		items[i].AssignmentName = titles[items[i].AssignmentID]
	}
	c.JSON(200, gin.H{"items": items, "total": total})
}
