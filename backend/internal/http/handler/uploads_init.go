package handler

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"homework-station/internal/model"
	"homework-station/internal/service"

	"github.com/gin-gonic/gin"
)

func (h *Handler) OwnedAttachment(c *gin.Context) (model.Attachment, bool) {
	var f model.Attachment
	if h.DB.First(&f, parseID(c.Param("id"))).Error != nil || f.UserID != current(c).ID {
		fail(c, 404, "上传记录不存在")
		return f, false
	}
	return f, true
}

func (h *Handler) InitUpload(c *gin.Context) {
	h.Files.MutationMu.Lock()
	defer h.Files.MutationMu.Unlock()
	var v struct {
		Name         string `json:"name"`
		Size         int64  `json:"size"`
		SHA256       string `json:"sha256"`
		AssignmentID uint   `json:"assignment_id"`
	}
	if c.ShouldBindJSON(&v) != nil || !regexp.MustCompile(`^[a-f0-9]{64}$`).MatchString(v.SHA256) || v.Size < 0 || v.Size > 10*1024*1024*1024 {
		fail(c, 400, "文件信息无效（最大 10GB）")
		return
	}
	v.Name = filepath.Base(strings.ReplaceAll(v.Name, `\`, "/"))
	if v.Name == "." || v.Name == "/" || len(v.Name) > 255 || strings.ContainsAny(v.Name, "\r\n\x00") {
		fail(c, 400, "文件名不正确")
		return
	}
	u := current(c)
	if !u.Admin {
		max, _ := strconv.ParseInt(h.Settings.Get("default_max_size"), 10, 64)
		ext := h.Settings.Get("default_extensions")
		if v.AssignmentID != 0 {
			var a model.Assignment
			if h.DB.First(&a, v.AssignmentID).Error != nil || a.Status != "published" || !a.AllowAttachments {
				fail(c, 400, "作业不允许附件上传")
				return
			}
			max = a.MaxSize
			ext = a.Extensions
		}
		if v.Size > max || !service.AllowedExtension(v.Name, ext) {
			fail(c, 400, "文件大小或后缀超出允许范围")
			return
		}
	}
	// Reuse only the current user's proven copy; other users must upload bytes before deduplication.
	var existing model.Attachment
	if h.DB.Where("user_id = ? AND sha256 = ? AND size = ? AND name = ?", u.ID, v.SHA256, v.Size, v.Name).Order("id desc").First(&existing).Error == nil {
		if existing.Status == "uploading" {
			if err := h.DB.Model(&existing).Update("upload_activity_at", time.Now()).Error; err != nil {
				fail(c, 500, "更新上传时间失败")
				return
			}
		}
		existing.Received = h.Files.Received(existing)
		c.JSON(200, existing)
		return
	}
	total := int((v.Size + service.ChunkBytes - 1) / service.ChunkBytes)
	now := time.Now()
	f := model.Attachment{UploadActivityAt: &now, UserID: u.ID, Name: v.Name, Size: v.Size, SHA256: v.SHA256, Status: "uploading", ChunkSize: service.ChunkBytes, TotalChunks: total}
	if h.DB.Create(&f).Error != nil {
		fail(c, 500, "创建上传失败")
		return
	}
	if e := os.MkdirAll(h.Files.ChunkDir(f.ID), 0700); e != nil {
		h.DB.Delete(&f)
		fail(c, 500, "创建分片目录失败")
		return
	}
	f.Received = []int{}
	c.JSON(201, f)
}

func (h *Handler) UploadStatus(c *gin.Context) {
	h.Files.MutationMu.Lock()
	defer h.Files.MutationMu.Unlock()
	f, ok := h.OwnedAttachment(c)
	if !ok {
		return
	}
	f.Received = h.Files.Received(f)
	c.JSON(200, f)
}
