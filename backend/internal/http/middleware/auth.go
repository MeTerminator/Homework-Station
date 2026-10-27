package middleware

import (
	"strconv"

	"homework-station/internal/http/response"
	"homework-station/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

func Auth(db *gorm.DB, sessionSecret []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		s, e := c.Cookie("session")
		if e != nil {
			response.Error(c, 401, "请先登录")
			return
		}
		t, e := jwt.Parse(s, func(t *jwt.Token) (interface{}, error) { return sessionSecret, nil }, jwt.WithValidMethods([]string{"HS256"}), jwt.WithExpirationRequired())
		if e != nil || !t.Valid {
			response.Error(c, 401, "登录已过期")
			return
		}
		claims := t.Claims.(jwt.MapClaims)
		id, _ := claims["sub"].(string)
		version, _ := claims["v"].(float64)
		var u model.User
		if db.First(&u, parseID(id)).Error != nil || uint(version) != u.AuthVersion {
			response.Error(c, 401, "请重新登录")
			return
		}
		c.Set("user", u)
		c.Next()
	}
}
func CurrentUser(c *gin.Context) model.User { return c.MustGet("user").(model.User) }
func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !CurrentUser(c).Admin {
			response.Error(c, 403, "需要管理员权限")
			return
		}
		c.Next()
	}
}
func parseID(s string) uint { n, _ := strconv.ParseUint(s, 10, 32); return uint(n) }
