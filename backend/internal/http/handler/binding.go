package handler

import (
	"regexp"
	"strings"

	"homework-station/internal/model"

	"github.com/gin-gonic/gin"
)

func (h *Handler) BindReal(c *gin.Context) {
	var v struct {
		StudentNo string `json:"student_no"`
		Name      string `json:"name"`
	}
	if c.ShouldBindJSON(&v) != nil || !regexp.MustCompile(`^[0-9]{10}$`).MatchString(v.StudentNo) || strings.TrimSpace(v.Name) == "" {
		fail(c, 400, "请输入 10 位数字学号和姓名")
		return
	}
	u := current(c)
	if u.StudentNo != nil {
		fail(c, 409, "已实名，请联系管理员解绑")
		return
	}
	fields, err := h.Settings.IdentityConfig()
	if err != nil {
		respondError(c, err)
		return
	}
	realName, err := h.Identity.Verify(fields, v.StudentNo, strings.TrimSpace(v.Name))
	if err != nil {
		respondError(c, err)
		return
	}
	resultDB := h.DB.Model(&model.User{}).Where("id = ? AND student_no IS NULL", u.ID).Updates(map[string]interface{}{"student_no": v.StudentNo, "real_name": realName})
	if resultDB.Error != nil || resultDB.RowsAffected != 1 {
		fail(c, 409, "此学号已绑定其他用户")
		return
	}
	h.DB.First(&u, u.ID)
	c.JSON(200, u)
}
func (h *Handler) BindQQ(c *gin.Context) {
	var input struct {
		Code string `json:"code"`
	}
	if c.ShouldBindJSON(&input) != nil {
		fail(c, 400, "请求格式不正确")
		return
	}
	user, err := h.Users.BindQQ(current(c), input.Code)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(200, user)
}
