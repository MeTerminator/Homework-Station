package middleware

import (
	"net/http"
	"net/url"
	"strings"

	"homework-station/internal/http/response"

	"github.com/gin-gonic/gin"
)

func Security() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		if strings.HasPrefix(c.GetHeader("Content-Type"), "application/json") {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 3*1024*1024)
		}
		c.Header("Referrer-Policy", "same-origin")
		c.Header("X-Frame-Options", "DENY")
		if c.Request.Method != "GET" && c.Request.Method != "HEAD" {
			if origin := c.GetHeader("Origin"); origin != "" {
				u, e := url.Parse(origin)
				if e != nil || u.Host != c.Request.Host {
					response.Error(c, 403, "来源校验失败")
					return
				}
			}
		}
		c.Next()
	}
}
