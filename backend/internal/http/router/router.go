package router

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"homework-station/internal/http/handler"
	"homework-station/internal/http/middleware"
	"homework-station/internal/http/response"

	"github.com/gin-gonic/gin"
)

func New(h *handler.Handler, output *log.Logger) *gin.Engine {
	r := gin.New()
	limiter := middleware.NewRateLimiter()
	r.SetTrustedProxies(nil)
	r.Use(gin.Recovery(), gin.LoggerWithWriter(output.Writer()), middleware.Security())
	r.GET("/api/health", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })
	r.GET("/api/public-settings", func(c *gin.Context) {
		c.JSON(200, gin.H{"site_title": h.Settings.Get("site_title"), "browser_title": h.Settings.Get("browser_title"), "site_logo": h.Settings.Get("site_logo"), "site_icon": h.Settings.Get("site_icon"), "registration_enabled": h.Settings.Get("registration_enabled"), "realname_enabled": strings.TrimSpace(h.Settings.Get("realname_api_url")) != ""})
	})
	r.GET("/api/branding/:name", h.BrandingImage)
	r.GET("/api/captcha", h.CaptchaHandler)
	r.POST("/api/register", limiter.Middleware(), h.Register)
	r.POST("/api/login", limiter.Middleware(), h.Login)
	r.POST("/api/logout", h.Logout)
	a := r.Group("/api", middleware.Auth(h.DB, h.SessionSecret))
	a.GET("/me", h.Me)
	a.PATCH("/me", h.UpdateMe)
	a.POST("/me/password", h.ChangePassword)
	a.POST("/me/avatar", h.UploadAvatar)
	a.POST("/me/bind/real", h.BindReal)
	a.POST("/me/bind/qq", h.BindQQ)
	a.GET("/categories", h.ListCategories)
	a.GET("/assignments", h.ListAssignments)
	a.GET("/assignments/:id", h.GetAssignment)
	a.GET("/assignments/:id/submissions", h.ListSubmissions)
	a.POST("/assignments/:id/submissions", h.CreateSubmission)
	a.POST("/submissions/:id/withdraw", h.WithdrawSubmission)
	a.POST("/uploads", h.InitUpload)
	a.GET("/uploads/:id", h.UploadStatus)
	a.PUT("/uploads/:id/chunks/:index", h.PutChunk)
	a.POST("/uploads/:id/complete", h.CompleteUpload)
	a.GET("/attachments", h.ListOwnAttachments)
	a.GET("/attachments/:id/download", h.DownloadAttachment)
	ad := a.Group("/admin", middleware.AdminOnly())
	ad.GET("/overview", h.Overview)
	ad.GET("/settings", h.GetSettings)
	ad.PUT("/settings", h.PutSettings)
	ad.POST("/settings/images", h.UploadBrandingImage)
	ad.POST("/categories", h.SaveCategory)
	ad.PUT("/categories/:id", h.SaveCategory)
	ad.DELETE("/categories/:id", h.DeleteCategory)
	ad.GET("/assignments", h.ListAssignments)
	ad.POST("/assignments", h.SaveAssignment)
	ad.PUT("/assignments/:id", h.SaveAssignment)
	ad.DELETE("/assignments/:id", h.DeleteAssignment)
	ad.GET("/users", h.ListUsers)
	ad.GET("/users/:id", h.GetUser)
	ad.PATCH("/users/:id", h.UpdateUser)
	ad.POST("/users/:id/unbind/:kind", h.Unbind)
	ad.GET("/attachments", h.ListAllAttachments)
	ad.DELETE("/attachments/:id", h.DeleteAttachment)
	ad.DELETE("/submissions/:id", h.DeleteSubmission)
	dist := h.Config.WebDist
	r.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			response.Error(c, 404, "接口不存在")
			return
		}
		p := filepath.Join(dist, filepath.Clean("/"+c.Request.URL.Path))
		if stat, e := os.Stat(p); e == nil && !stat.IsDir() {
			c.File(p)
			return
		}
		if _, e := os.Stat(filepath.Join(dist, "index.html")); e == nil {
			c.File(filepath.Join(dist, "index.html"))
		} else {
			response.Error(c, 404, "请先构建前端，开发时访问 http://localhost:5173")
		}
	})
	return r
}
