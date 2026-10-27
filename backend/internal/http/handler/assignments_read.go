package handler

import (
	"strings"
	"time"

	"homework-station/internal/model"

	"github.com/gin-gonic/gin"
)

func (h *Handler) ListAssignments(c *gin.Context) {
	u := current(c)
	admin := strings.Contains(c.FullPath(), "/admin/")
	q := h.DB.Model(&model.Assignment{})
	if !admin {
		q = q.Where("status <> ?", "hidden")
	}
	if s := c.Query("search"); s != "" {
		q = q.Where("LOWER(name) LIKE ?", "%"+strings.ToLower(s)+"%")
	}
	if cat := c.Query("categories"); cat != "" {
		ids := []uint{}
		for _, s := range strings.Split(cat, ",") {
			if id := parseID(s); id > 0 {
				ids = append(ids, id)
			}
		}
		q = q.Where("category_id IN ?", ids)
	}
	var total int64
	q.Count(&total)
	items := []model.Assignment{}
	if admin {
		page, size := pagination(c)
		q.Order("sort asc, id desc").Offset((page - 1) * size).Limit(size).Find(&items)
	} else {
		q.Order("sort asc, id desc").Find(&items)
	}
	for i := range items {
		h.DecorateAssignment(&items[i], u.ID)
	}
	c.JSON(200, gin.H{"items": items, "total": total})
}

func (h *Handler) VisibleAssignment(c *gin.Context) (model.Assignment, bool) {
	var a model.Assignment
	if h.DB.First(&a, parseID(c.Param("id"))).Error != nil || (!current(c).Admin && a.Status == "hidden") {
		fail(c, 404, "作业不存在")
		return a, false
	}
	return a, true
}

func (h *Handler) GetAssignment(c *gin.Context) {
	a, ok := h.VisibleAssignment(c)
	if !ok {
		return
	}
	h.DecorateAssignment(&a, current(c).ID)
	files := []model.Attachment{}
	h.DB.Joins("JOIN assignment_attachments ON assignment_attachments.attachment_id = attachments.id").Where("assignment_attachments.assignment_id = ?", a.ID).Find(&files)
	c.JSON(200, gin.H{"assignment": a, "attachments": files})
}

func (h *Handler) DecorateAssignment(a *model.Assignment, uid uint) {
	h.DB.Model(&model.Submission{}).Where("assignment_id = ? AND withdrawn = ?", a.ID, false).Distinct("user_id").Count(&a.Submitters)
	var n int64
	h.DB.Model(&model.Submission{}).Where("assignment_id = ? AND user_id = ? AND withdrawn = ?", a.ID, uid, false).Count(&n)
	a.Submitted = n > 0
	if a.Status == "published" && a.Deadline != nil && a.Deadline.Before(time.Now()) {
		a.Status = "closed"
	}
}
