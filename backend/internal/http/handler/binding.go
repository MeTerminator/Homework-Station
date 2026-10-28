package handler

import (
	"regexp"
	"strings"
	"unicode/utf8"

	"homework-station/internal/model"

	"github.com/gin-gonic/gin"
)

func (h *Handler) BindReal(c *gin.Context) {
	var v struct {
		StudentNo string `json:"student_no"`
		Name      string `json:"name"`
	}
	if c.ShouldBindJSON(&v) != nil || v.StudentNo == "" || utf8.RuneCountInString(v.StudentNo) > 128 || strings.TrimSpace(v.Name) == "" {
		fail(c, 400, "请输入学号（最多 128 个字符）和姓名")
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
	pattern, err := regexp.Compile(`\A(?:` + fields["realname_student_no_pattern"] + `)\z`)
	if err != nil || fields["realname_student_no_pattern"] == "" {
		fail(c, 503, "学号格式配置无效，请联系管理员")
		return
	}
	if !pattern.MatchString(v.StudentNo) {
		fail(c, 400, "学号格式不正确："+fields["realname_student_no_placeholder"])
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
