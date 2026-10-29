package handler

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"homework-station/internal/model"

	"github.com/gin-gonic/gin"
)

func (h *Handler) ListOwnAttachments(c *gin.Context) {
	files := []model.Attachment{}
	if err := h.DB.Where("user_id = ?", current(c).ID).Order("id desc").Find(&files).Error; err != nil {
		fail(c, 500, "加载附件失败")
		return
	}
	for i := range files {
		files[i].Usage = h.Files.AttachmentUsage(files[i].ID)
		if files[i].Status == "uploading" {
			files[i].Received = h.Files.Received(files[i])
		}
	}
	c.JSON(200, files)
}

func (h *Handler) ListAllAttachments(c *gin.Context) {
	page, size := pagination(c)
	q := h.DB.Model(&model.Attachment{})
	if status := c.Query("status"); status == "complete" || status == "uploading" {
		q = q.Where("status = ?", status)
	}
	if id := parseID(c.Query("user_id")); id > 0 {
		q = q.Where("user_id = ?", id)
	}
	if s := c.Query("search"); s != "" {
		q = q.Where("LOWER(name) LIKE ? OR sha256 = ?", "%"+strings.ToLower(s)+"%", s)
	}
	var total int64
	q.Count(&total)
	files := []model.Attachment{}
	q.Order("id desc").Offset((page - 1) * size).Limit(size).Find(&files)
	for i := range files {
		files[i].Usage = h.Files.AttachmentUsage(files[i].ID)
		if files[i].Status == "uploading" {
			files[i].Received = h.Files.Received(files[i])
		}
	}
	c.JSON(200, gin.H{"items": files, "total": total})
}

func (h *Handler) DownloadAttachment(c *gin.Context) {
	var f model.Attachment
	if h.DB.First(&f, parseID(c.Param("id"))).Error != nil || !h.Files.CanDownload(current(c), f) {
		fail(c, 404, "附件不存在或无权下载")
		return
	}
	c.Header("Cache-Control", "private, no-store")
	if f.Status != "complete" {
		if !current(c).Admin {
			fail(c, 409, "文件尚未上传完成")
			return
		}
		h.Files.MutationMu.Lock()
		defer h.Files.MutationMu.Unlock()
		c.Header("Content-Type", "application/zip")
		c.Header("Content-Disposition", `attachment; filename="upload-`+fmt.Sprint(f.ID)+`-partial.zip"`)
		zw := zip.NewWriter(c.Writer)
		meta, _ := zw.Create("manifest.json")
		f.Received = h.Files.Received(f)
		json.NewEncoder(meta).Encode(f)
		for _, i := range f.Received {
			in, e := os.Open(h.Files.ChunkPath(f.ID, i))
			if e != nil {
				continue
			}
			out, e := zw.Create(fmt.Sprintf("chunks/%d", i))
			if e == nil {
				io.Copy(out, in)
			}
			in.Close()
		}
		zw.Close()
		return
	}
	var b model.Blob
	if h.DB.First(&b, "sha256 = ?", f.SHA256).Error != nil {
		fail(c, 404, "文件实体不存在")
		return
	}
	if c.Query("inline") == "true" && strings.HasSuffix(f.Name, ".webp") {
		c.Header("Content-Type", "image/webp")
		c.File(b.Path)
	} else {
		c.Header("Content-Type", "application/octet-stream")
		c.FileAttachment(b.Path, f.Name)
	}
}

func (h *Handler) DeleteOwnAttachment(c *gin.Context) {
	h.deleteAttachment(c, true)
}

func (h *Handler) DeleteAttachment(c *gin.Context) {
	h.deleteAttachment(c, false)
}

func (h *Handler) deleteAttachment(c *gin.Context, ownOnly bool) {
	h.Files.MutationMu.Lock()
	defer h.Files.MutationMu.Unlock()
	var f model.Attachment
	if h.DB.First(&f, parseID(c.Param("id"))).Error != nil || (ownOnly && f.UserID != current(c).ID) {
		fail(c, 404, "附件不存在")
		return
	}
	usage := h.Files.AttachmentUsage(f.ID)
	if len(usage) > 0 && c.Query("force") != "true" {
		c.JSON(409, gin.H{"error": "附件仍被使用，确认删除将解除这些关联", "usage": usage})
		return
	}
	if e := h.Files.RemoveAttachment(f); e != nil {
		fail(c, 500, "删除失败："+e.Error())
		return
	}
	c.JSON(200, gin.H{"ok": true})
}
