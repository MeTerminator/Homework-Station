package handler

import (
	"crypto/sha256"
	"fmt"
	"image"
	"image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"

	"github.com/gin-gonic/gin"
)

const brandingMaxSize = 2 * 1024 * 1024

var brandingName = regexp.MustCompile(`^[a-f0-9]{64}\.png$`)

// Branding assets are public, separate from private user attachments.
func (h *Handler) BrandingImage(c *gin.Context) {
	name := c.Param("name")
	if !brandingName.MatchString(name) {
		fail(c, 404, "图片不存在")
		return
	}
	path := filepath.Join(h.Config.DataDir, "branding", name)
	if _, err := os.Stat(path); err != nil {
		fail(c, 404, "图片不存在")
		return
	}
	c.Header("Cache-Control", "public, max-age=31536000, immutable")
	c.File(path)
}

func (h *Handler) UploadBrandingImage(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, brandingMaxSize+1024*1024)
	if err := c.Request.ParseMultipartForm(brandingMaxSize); err != nil {
		fail(c, 400, "图片超过大小限制或格式无效")
		return
	}
	defer c.Request.MultipartForm.RemoveAll()
	f, err := c.FormFile("image")
	if err != nil || f.Size > brandingMaxSize {
		fail(c, 400, "请上传不超过 2MB 的图片")
		return
	}
	in, err := f.Open()
	if err != nil {
		fail(c, 400, "无法读取图片")
		return
	}
	defer in.Close()
	cfg, _, err := image.DecodeConfig(in)
	if err != nil || cfg.Width < 1 || cfg.Height < 1 || int64(cfg.Width)*int64(cfg.Height) > 4_000_000 {
		fail(c, 400, "请上传 PNG、JPEG、GIF 或 WebP 图片，最多 400 万像素")
		return
	}
	if _, err = in.Seek(0, 0); err != nil {
		fail(c, 400, "无法读取图片")
		return
	}
	img, _, err := image.Decode(in)
	if err != nil {
		fail(c, 400, "图片解码失败")
		return
	}
	dir := filepath.Join(h.Config.DataDir, "branding")
	if err = os.MkdirAll(dir, 0700); err != nil {
		fail(c, 500, "保存图片失败")
		return
	}
	tmp, err := os.CreateTemp(dir, "upload-*")
	if err != nil {
		fail(c, 500, "保存图片失败")
		return
	}
	defer os.Remove(tmp.Name())
	hash := sha256.New()
	err = png.Encode(io.MultiWriter(tmp, hash), img)
	closeErr := tmp.Close()
	if err != nil || closeErr != nil {
		fail(c, 500, "保存图片失败")
		return
	}
	name := fmt.Sprintf("%x.png", hash.Sum(nil))
	if err = os.Rename(tmp.Name(), filepath.Join(dir, name)); err != nil {
		fail(c, 500, "保存图片失败")
		return
	}
	c.JSON(201, gin.H{"url": "/api/branding/" + name})
}
