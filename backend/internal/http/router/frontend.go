package router

import (
	"bytes"
	"io/fs"
	"net/http"
	"path"
	"strings"
	"time"

	"homework-station/internal/http/response"

	"github.com/gin-gonic/gin"
)

func frontendHandler(files fs.FS) gin.HandlerFunc {
	var index []byte
	var static http.Handler
	if files != nil {
		index, _ = fs.ReadFile(files, "index.html")
		static = http.FileServerFS(files)
	}
	return func(c *gin.Context) {
		urlPath := c.Request.URL.Path
		if urlPath == "/api" || strings.HasPrefix(urlPath, "/api/") {
			response.Error(c, 404, "接口不存在")
			return
		}
		if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead {
			c.Status(http.StatusNotFound)
			return
		}
		if static == nil || len(index) == 0 {
			response.Error(c, 404, "开发时访问 http://localhost:5173，生产部署请使用 make build 构建的程序")
			return
		}
		name := strings.TrimPrefix(path.Clean("/"+urlPath), "/")
		if info, err := fs.Stat(files, name); err == nil && info.Mode().IsRegular() {
			if strings.HasPrefix(name, "assets/") {
				c.Header("Cache-Control", "public, max-age=31536000, immutable")
			} else {
				c.Header("Cache-Control", "no-cache")
			}
			static.ServeHTTP(c.Writer, c.Request)
			return
		}
		if strings.HasPrefix(name, "assets/") || path.Ext(name) != "" {
			c.Status(http.StatusNotFound)
			return
		}
		c.Header("Cache-Control", "no-cache")
		http.ServeContent(c.Writer, c.Request, "index.html", time.Time{}, bytes.NewReader(index))
	}
}
