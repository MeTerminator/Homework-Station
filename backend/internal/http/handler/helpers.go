package handler

import (
	"encoding/base64"
	"errors"
	"strconv"

	"homework-station/internal/http/middleware"
	"homework-station/internal/http/response"
	"homework-station/internal/model"
	"homework-station/internal/service"

	"github.com/gin-gonic/gin"
)

func fail(c *gin.Context, status int, message string) {
	response.Error(c, status, message)
}
func current(c *gin.Context) model.User { return middleware.CurrentUser(c) }
func parseID(s string) uint             { n, _ := strconv.ParseUint(s, 10, 32); return uint(n) }
func pagination(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	if page < 1 {
		page = 1
	}
	switch size {
	case 10, 20, 50, 100:
	default:
		size = 20
	}
	return page, size
}
func base64String(b []byte) string { return base64.StdEncoding.EncodeToString(b) }
func respondError(c *gin.Context, err error) {
	var applicationError *service.Error
	if errors.As(err, &applicationError) {
		fail(c, applicationError.Status, applicationError.Message)
	} else {
		fail(c, 500, "操作失败")
	}
}
