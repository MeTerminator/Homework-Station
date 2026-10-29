package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"gorm.io/gorm/clause"

	"homework-station/internal/model"
)

func uploadRetention(value, unit string) (time.Duration, error) {
	units := map[string]time.Duration{
		"second": time.Second, "minute": time.Minute, "hour": time.Hour,
		"day": 24 * time.Hour, "week": 7 * 24 * time.Hour, "month": 30 * 24 * time.Hour,
	}
	multiplier, ok := units[unit]
	if !ok {
		return 0, fmt.Errorf("上传临时文件清理单位无效")
	}
	amount, err := strconv.ParseInt(value, 10, 64)
	if err != nil || amount < 1 || amount > int64(time.Duration(1<<63-1)/multiplier) {
		return 0, fmt.Errorf("上传临时文件保留时长需为正整数，且总时长不能超过 292 年")
	}
	return time.Duration(amount) * multiplier, nil
}

func (s *SettingsService) UploadRetention() (time.Duration, error) {
	var rows []model.Setting
	if err := s.db.Where(clause.IN{Column: "key", Values: []interface{}{"upload_cleanup_value", "upload_cleanup_unit"}}).Find(&rows).Error; err != nil {
		return 0, err
	}
	values := map[string]string{}
	for _, row := range rows {
		values[row.Key] = row.Value
	}
	return uploadRetention(values["upload_cleanup_value"], values["upload_cleanup_unit"])
}

// CleanupUploads only removes abandoned upload records and their chunk directories.
// Completed files, blobs, submission associations and avatars are outside its scope.
func (s *FileService) CleanupUploads(now time.Time, retention time.Duration) (int, error) {
	if retention <= 0 {
		return 0, fmt.Errorf("上传临时文件保留时长必须大于零")
	}
	s.MutationMu.Lock()
	defer s.MutationMu.Unlock()
	cutoff := now.Add(-retention)
	var files []model.Attachment
	if err := s.db.Where("status = ? AND (upload_activity_at <= ? OR (upload_activity_at IS NULL AND created_at <= ?))", "uploading", cutoff, cutoff).Find(&files).Error; err != nil {
		return 0, err
	}
	removed := 0
	var failures []error
	for _, file := range files {
		// Older installations have no activity timestamp. Preserve a recently
		// written chunk even if the upload record itself was created long ago.
		if file.UploadActivityAt == nil {
			activity := file.CreatedAt
			entries, err := os.ReadDir(s.ChunkDir(file.ID))
			if err != nil && !os.IsNotExist(err) {
				failures = append(failures, fmt.Errorf("附件 #%d 分片检查失败: %w", file.ID, err))
				continue
			}
			valid := true
			for _, entry := range entries {
				info, err := entry.Info()
				if err != nil {
					failures = append(failures, fmt.Errorf("附件 #%d 分片检查失败: %w", file.ID, err))
					valid = false
					break
				}
				if info.ModTime().After(activity) {
					activity = info.ModTime()
				}
			}
			if !valid {
				continue
			}
			if activity.After(cutoff) {
				if err := s.db.Model(&file).Update("upload_activity_at", activity).Error; err != nil {
					failures = append(failures, err)
				}
				continue
			}
		}
		if err := os.RemoveAll(s.ChunkDir(file.ID)); err != nil {
			failures = append(failures, fmt.Errorf("附件 #%d 分片清理失败: %w", file.ID, err))
			continue
		}
		if err := s.db.Delete(&file).Error; err != nil {
			failures = append(failures, fmt.Errorf("附件 #%d 上传记录清理失败: %w", file.ID, err))
			continue
		}
		removed++
	}
	return removed, errors.Join(failures...)
}

// Settings changes trigger a sweep immediately. Between changes, short retention
// periods are checked each second; longer ones are checked at least each minute.
func (s *FileService) RunUploadCleanup(ctx context.Context, settings *SettingsService, output *log.Logger) {
	for {
		if ctx.Err() != nil {
			return
		}
		interval := time.Minute
		retention, err := settings.UploadRetention()
		if err == nil {
			if retention < interval {
				interval = time.Second
			}
			var removed int
			removed, err = s.CleanupUploads(time.Now(), retention)
			if removed > 0 {
				output.Printf("上传临时文件自动清理：已删除 %d 个未完成附件及其分片", removed)
			}
		}
		if err != nil {
			output.Printf("上传临时文件自动清理失败：%v", err)
		}
		timer := time.NewTimer(interval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-settings.uploadCleanupChanged:
			timer.Stop()
		case <-timer.C:
		}
	}
}
