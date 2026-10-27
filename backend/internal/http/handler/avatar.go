package handler

import (
	"crypto/sha256"
	"encoding/hex"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"homework-station/internal/model"

	"github.com/chai2010/webp"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func (h *Handler) UploadAvatar(c *gin.Context) {
	h.Files.MutationMu.Lock()
	defer h.Files.MutationMu.Unlock()
	max, _ := strconv.ParseInt(h.Settings.Get("avatar_max_size"), 10, 64)
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, max+1024*1024)
	if e := c.Request.ParseMultipartForm(1024 * 1024); e != nil {
		fail(c, 400, "图片超过大小限制或格式无效")
		return
	}
	if c.Request.MultipartForm != nil {
		defer c.Request.MultipartForm.RemoveAll()
	}
	fh, e := c.FormFile("avatar")
	if e != nil || fh.Size > max {
		fail(c, 400, "头像超过大小限制")
		return
	}
	in, e := fh.Open()
	if e != nil {
		fail(c, 400, "无法读取图片")
		return
	}
	defer in.Close()
	cfg, _, e := image.DecodeConfig(in)
	if e != nil || cfg.Width < 1 || cfg.Height < 1 || int64(cfg.Width)*int64(cfg.Height) > 25_000_000 {
		fail(c, 400, "请上传 JPEG、PNG、GIF 或 WebP 图片，最多 2500 万像素")
		return
	}
	in.Seek(0, 0)
	img, _, e := image.Decode(in)
	if e != nil {
		fail(c, 400, "图片解码失败")
		return
	}
	tmp, e := os.CreateTemp(filepath.Join(h.Config.DataDir, "blobs"), "avatar-*")
	if e != nil {
		fail(c, 500, "保存失败")
		return
	}
	defer os.Remove(tmp.Name())
	if e = webp.Encode(tmp, img, &webp.Options{Quality: 85}); e != nil {
		tmp.Close()
		fail(c, 500, "WebP 转换失败")
		return
	}
	tmp.Close()
	bytes, e := os.ReadFile(tmp.Name())
	if e != nil {
		fail(c, 500, "读取头像失败")
		return
	}
	hash := sha256.Sum256(bytes)
	sha := hex.EncodeToString(hash[:])
	path := filepath.Join(h.Config.DataDir, "blobs", sha)
	var b model.Blob
	if h.DB.First(&b, "sha256 = ?", sha).Error != nil {
		if e = os.Rename(tmp.Name(), path); e != nil {
			fail(c, 500, "保存头像失败")
			return
		}
		if h.DB.Create(&model.Blob{SHA256: sha, Size: int64(len(bytes)), Path: path}).Error != nil {
			fail(c, 500, "保存头像信息失败")
			return
		}
	}
	u := current(c)
	f := model.Attachment{UserID: u.ID, Name: "avatar.webp", SHA256: sha, Size: int64(len(bytes)), Status: "complete"}
	e = h.DB.Transaction(func(tx *gorm.DB) error {
		if e := tx.Create(&f).Error; e != nil {
			return e
		}
		return tx.Model(&u).Update("avatar_id", f.ID).Error
	})
	if e != nil {
		fail(c, 500, "保存头像失败")
		return
	}
	h.DB.First(&u, u.ID)
	c.JSON(200, u)
}
