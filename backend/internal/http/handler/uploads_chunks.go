package handler

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"homework-station/internal/model"

	"github.com/gin-gonic/gin"
)

func (h *Handler) PutChunk(c *gin.Context) {
	h.Files.MutationMu.Lock()
	defer h.Files.MutationMu.Unlock()
	f, ok := h.OwnedAttachment(c)
	if !ok {
		return
	}
	i, e := strconv.Atoi(c.Param("index"))
	if e != nil || i < 0 || i >= f.TotalChunks || f.Status != "uploading" {
		fail(c, 400, "分片序号或状态无效")
		return
	}
	expected := f.ChunkSize
	if i == f.TotalChunks-1 {
		expected = f.Size - int64(i)*f.ChunkSize
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, expected+1)
	defer c.Request.Body.Close()
	tmp := h.Files.ChunkPath(f.ID, i) + ".tmp"
	out, e := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if e != nil {
		fail(c, 500, "无法保存分片")
		return
	}
	n, e := io.Copy(out, c.Request.Body)
	closeErr := out.Close()
	if e != nil || closeErr != nil || n != expected {
		os.Remove(tmp)
		fail(c, 400, "分片大小不正确或传输中断")
		return
	}
	if e = os.Rename(tmp, h.Files.ChunkPath(f.ID, i)); e != nil {
		os.Remove(tmp)
		fail(c, 500, "保存失败")
		return
	}
	c.JSON(200, gin.H{"ok": true, "index": i})
}

func (h *Handler) CompleteUpload(c *gin.Context) {
	h.Files.MutationMu.Lock()
	defer h.Files.MutationMu.Unlock()
	f, ok := h.OwnedAttachment(c)
	if !ok {
		return
	}
	if f.Status == "complete" {
		c.JSON(200, f)
		return
	}
	if len(h.Files.Received(f)) != f.TotalChunks {
		fail(c, 409, "仍有分片未上传")
		return
	}
	tmp, e := os.CreateTemp(filepath.Join(h.Config.DataDir, "blobs"), "merge-*")
	if e != nil {
		fail(c, 500, "无法合并文件")
		return
	}
	defer os.Remove(tmp.Name())
	checksum := sha256.New()
	var size int64
	for i := 0; i < f.TotalChunks; i++ {
		part, e := os.Open(h.Files.ChunkPath(f.ID, i))
		if e != nil {
			tmp.Close()
			fail(c, 409, "分片缺失")
			return
		}
		n, e := io.Copy(io.MultiWriter(tmp, checksum), part)
		part.Close()
		size += n
		if e != nil {
			tmp.Close()
			fail(c, 500, "合并失败")
			return
		}
	}
	if e := tmp.Close(); e != nil {
		fail(c, 500, "合并失败")
		return
	}
	if size != f.Size || hex.EncodeToString(checksum.Sum(nil)) != f.SHA256 {
		os.RemoveAll(h.Files.ChunkDir(f.ID))
		os.MkdirAll(h.Files.ChunkDir(f.ID), 0700)
		fail(c, 422, "SHA256 校验失败，已清空损坏分片，请重新选择原文件续传")
		return
	}
	path := filepath.Join(h.Config.DataDir, "blobs", f.SHA256)
	var b model.Blob
	e = h.DB.First(&b, "sha256 = ?", f.SHA256).Error
	if e != nil {
		if e = os.Rename(tmp.Name(), path); e != nil {
			fail(c, 500, "保存文件失败")
			return
		}
		b = model.Blob{SHA256: f.SHA256, Size: f.Size, Path: path}
		if e = h.DB.Create(&b).Error; e != nil {
			os.Remove(path)
			fail(c, 500, "保存文件信息失败")
			return
		}
	} else if b.Size != f.Size {
		fail(c, 409, "文件校验冲突")
		return
	}
	if e = h.DB.Model(&f).Update("status", "complete").Error; e != nil {
		fail(c, 500, "保存上传状态失败")
		return
	}
	os.RemoveAll(h.Files.ChunkDir(f.ID))
	f.Status = "complete"
	c.JSON(200, f)
}
