package handler

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"homework-station/internal/model"
	"homework-station/internal/security"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func (h *Handler) Register(c *gin.Context) {
	var v struct {
		Username  string `json:"username"`
		Password  string `json:"password"`
		Confirm   string `json:"confirm"`
		CaptchaID string `json:"captcha_id"`
		Captcha   string `json:"captcha"`
	}
	if c.ShouldBindJSON(&v) != nil {
		fail(c, 400, "请求格式不正确")
		return
	}
	if h.Settings.Get("registration_enabled") != "true" {
		fail(c, 403, "注册已关闭")
		return
	}
	if !h.ConsumeCaptcha(v.CaptchaID, v.Captcha) {
		fail(c, 400, "验证码错误或已过期，请刷新")
		return
	}
	v.Username = strings.TrimSpace(v.Username)
	if !security.ValidUsername(v.Username) || !security.ValidPassword(v.Password) || v.Password != v.Confirm {
		fail(c, 400, "用户名需为 2–32 位文字、数字或 _.-，密码需为 6–72 字节且两次一致")
		return
	}
	pw, _ := bcrypt.GenerateFromPassword([]byte(v.Password), bcrypt.DefaultCost)
	u := model.User{Username: v.Username, Password: string(pw), CategoryIDs: "[]"}
	if h.DB.Create(&u).Error != nil {
		fail(c, 409, "用户名已存在")
		return
	}
	h.IssueSession(c, u)
	c.JSON(201, u)
}

func (h *Handler) Login(c *gin.Context) {
	var v struct {
		Username string
		Password string
	}
	if c.ShouldBindJSON(&v) != nil {
		fail(c, 400, "请求格式不正确")
		return
	}
	var u model.User
	e := h.DB.First(&u, "username = ?", v.Username).Error
	if e != nil || bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(v.Password)) != nil {
		fail(c, 401, "用户名或密码错误")
		return
	}
	h.IssueSession(c, u)
	c.JSON(200, u)
}

func (h *Handler) Logout(c *gin.Context) {
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie("session", "", -1, "/", "", h.Config.CookieSecure, true)
	c.JSON(200, gin.H{"ok": true})
}

func (h *Handler) IssueSession(c *gin.Context, u model.User) {
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": fmt.Sprint(u.ID), "v": u.AuthVersion, "exp": time.Now().Add(7 * 24 * time.Hour).Unix(), "iat": time.Now().Unix()})
	s, _ := t.SignedString(h.SessionSecret)
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie("session", s, 7*86400, "/", "", h.Config.CookieSecure, true)
}
