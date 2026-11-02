package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand/v2"
	"net/http"
	"sync"
	"time"

	"gorm.io/gorm"

	"homework-station/internal/model"
)

// Review rows form a durable queue, including legacy admin reviews after migration.
// One worker serializes sends and resend requests within the application.
type NotificationService struct {
	db       *gorm.DB
	settings *SettingsService
	mu       sync.Mutex
	client   *http.Client
}

func NewNotificationService(db *gorm.DB, settings *SettingsService) *NotificationService {
	return &NotificationService{db: db, settings: settings, client: &http.Client{Timeout: 15 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}
}
func (s *NotificationService) Requeue(id uint) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	q := s.db.Model(&model.ReviewMessage{}).Where("admin = ?", true)
	if id != 0 {
		q = q.Where("id = ?", id)
	}
	// MySQL may report zero affected rows when an already-pending row is unchanged.
	var count int64
	if err := q.Session(&gorm.Session{}).Count(&count).Error; err != nil {
		return 0, Reject(500, "读取通知失败")
	}
	if id != 0 && count == 0 {
		return 0, Reject(404, "管理员评论不存在")
	}
	result := q.Updates(map[string]interface{}{"notification_pending": true, "notification_next_at": nil, "notification_status": "pending"})
	if result.Error != nil {
		return 0, Reject(500, "通知入队失败")
	}
	return count, nil
}
func (s *NotificationService) Run(ctx context.Context, output *log.Logger) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.ProcessNext(ctx, true); err != nil && ctx.Err() == nil {
				output.Printf("评论通知队列处理失败: %v", err)
			}
		}
	}
}
func (s *NotificationService) ProcessNext(ctx context.Context, delay bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	endpoint, secret := s.settings.Get("qq_notification_url"), s.settings.Get("qq_notification_secret")
	if endpoint == "" || secret == "" {
		return nil
	}
	var m model.ReviewMessage
	err := s.db.Where("admin = ? AND notification_pending = ? AND (notification_next_at IS NULL OR notification_next_at <= ?)", true, true, time.Now()).Order("id asc").First(&m).Error
	if err == gorm.ErrRecordNotFound {
		return nil
	}
	if err != nil {
		return err
	}
	if delay {
		timer := time.NewTimer(time.Duration(1000+rand.IntN(2001)) * time.Millisecond)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
		}
	}
	// Read recipient preferences immediately before sending, rather than at enqueue time.
	var submission model.Submission
	var user, author model.User
	if err := s.db.First(&submission, m.SubmissionID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return s.finish(m.ID, "skipped_missing_submission", false)
		}
		return err
	}
	if err := s.db.First(&user, submission.UserID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return s.finish(m.ID, "skipped_missing_user", false)
		}
		return err
	}
	if user.QQ == nil || *user.QQ == "" {
		return s.finish(m.ID, "skipped_unbound", false)
	}
	if !user.QQCommentNotifications {
		return s.finish(m.ID, "skipped_disabled", false)
	}
	if err := s.db.First(&author, m.UserID).Error; err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	if author.Username == "" {
		author.Username = "已删除用户"
	}
	zone := time.FixedZone("Asia/Shanghai", 8*60*60)
	message := fmt.Sprintf("# Homework Comments\nStaff: %s\nTime: %s\n\n%s", author.Username, m.CreatedAt.In(zone).Format("2006-01-02 15:04:05"), m.Content)
	payload, err := json.Marshal(map[string]interface{}{"review_id": m.ID, "qq": *user.QQ, "message": message})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+secret)
	resp, err := s.client.Do(req)
	if err == nil {
		defer resp.Body.Close()
		var result struct {
			Status string `json:"status"`
		}
		if resp.StatusCode == http.StatusOK && json.NewDecoder(io.LimitReader(resp.Body, 4096)).Decode(&result) == nil {
			switch result.Status {
			case "sent":
				return s.finish(m.ID, "sent", true)
			case "not_friend":
				return s.finish(m.ID, "skipped_not_friend", false)
			}
		}
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}
	next := time.Now().Add(time.Minute)
	// Failed requests remain queued and cannot starve newer comments.
	return s.db.Model(&model.ReviewMessage{}).Where("id = ?", m.ID).Updates(map[string]interface{}{"notification_status": "retry", "notification_next_at": next}).Error
}
func (s *NotificationService) finish(id uint, status string, sent bool) error {
	changes := map[string]interface{}{"notification_pending": false, "notification_status": status, "notification_next_at": nil}
	if sent {
		changes["last_notified_at"] = time.Now().UTC()
	}
	return s.db.Model(&model.ReviewMessage{}).Where("id = ?", id).Updates(changes).Error
}
