package app

import (
	"fmt"
	"testing"

	"homework-station/internal/config"
	"homework-station/internal/model"

	"github.com/gin-gonic/gin"
)

func TestReviewConversation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("DATA_DIR", t.TempDir())
	t.Setenv("DB_DRIVER", "sqlite")
	t.Setenv("DB_DSN", "")
	t.Setenv("JWT_SECRET", "test-session-secret-32-characters-long")
	app, err := New(config.Load())
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()
	admin := &testClient{t: t, r: app.Router}
	owner := &testClient{t: t, r: app.Router}
	other := &testClient{t: t, r: app.Router}
	guest := &testClient{t: t, r: app.Router}
	admin.request("POST", "/api/login", gin.H{"username": "admin", "password": "123456"}, 200)
	owner.register("review-owner")
	second := other.register("review-admin")
	a := model.Assignment{Name: "Review assignment", Status: "published", Multiple: true, Withdraw: true}
	if err := app.DB.Create(&a).Error; err != nil {
		t.Fatal(err)
	}
	base := fmt.Sprintf("/api/assignments/%d", a.ID)
	s := decode[model.Submission](t, owner.request("POST", base+"/submissions", gin.H{"content": "answer"}, 201))
	path := fmt.Sprintf("/api/submissions/%d/reviews", s.ID)
	status := func(want string) {
		t.Helper()
		var current model.Submission
		app.DB.First(&current, s.ID)
		if current.ReviewStatus != want {
			t.Fatalf("status = %s want %s", current.ReviewStatus, want)
		}
		detail := decode[struct {
			Assignment model.Assignment `json:"assignment"`
		}](t, owner.request("GET", base, nil, 200))
		if detail.Assignment.ReviewStatus != want {
			t.Fatalf("assignment status = %s want %s", detail.Assignment.ReviewStatus, want)
		}
	}
	status("submitted")
	guest.request("GET", path, nil, 401)
	other.request("GET", path, nil, 404)
	other.request("POST", path, gin.H{"content": "intrude"}, 404)
	owner.request("POST", path, gin.H{"content": "too early"}, 409)
	admin.request("POST", path, gin.H{"content": "  "}, 400)
	payload := []byte("review attachment")
	f := admin.send(admin.init("feedback.txt", payload, 0), payload)
	owner.request("GET", fmt.Sprintf("/api/attachments/%d/download", f.ID), nil, 404)
	admin.request("POST", path, gin.H{"attachment_ids": []uint{f.ID, f.ID}}, 400)
	m := decode[model.ReviewMessage](t, admin.request("POST", path, gin.H{"content": "# Feedback\n**Please revise**", "attachment_ids": []uint{f.ID}}, 201))
	if !m.Admin || m.Username != "admin" || len(m.Attachments) != 1 {
		t.Fatal("missing author or attachment")
	}
	status("reviewed")
	// Authors expose only their current public name/avatar; private account fields stay out.
	app.DB.Model(&model.User{}).Where("id = ?", m.UserID).Update("avatar_id", f.ID)
	withAvatar := decode[[]model.ReviewMessage](t, owner.request("GET", path, nil, 200))
	if withAvatar[0].AvatarID == nil || *withAvatar[0].AvatarID != f.ID {
		t.Fatal("author avatar missing")
	}
	app.DB.Model(&model.User{}).Where("id = ?", m.UserID).Update("avatar_id", nil)
	owner.request("GET", fmt.Sprintf("/api/attachments/%d/download", f.ID), nil, 200)
	other.request("GET", fmt.Sprintf("/api/attachments/%d/download", f.ID), nil, 404)
	owner.request("POST", path, gin.H{"attachment_ids": []uint{f.ID}}, 400)
	replyFile := owner.send(owner.init("reply.txt", []byte("reply"), 0), []byte("reply"))
	owner.request("POST", path, gin.H{"content": "Updated reasoning", "attachment_ids": []uint{replyFile.ID}}, 201)
	status("pending")
	// Multiple admins can participate; roles remain the role at send time.
	if err := app.DB.Model(&second).Update("admin", true).Error; err != nil {
		t.Fatal(err)
	}
	other.request("POST", path, gin.H{"content": "Second reviewer"}, 201)
	status("reviewed")
	if err := app.DB.Model(&second).Update("admin", false).Error; err != nil {
		t.Fatal(err)
	}
	messages := decode[[]model.ReviewMessage](t, owner.request("GET", path, nil, 200))
	if len(messages) != 3 || messages[0].Content != m.Content || messages[1].Admin || !messages[2].Admin || messages[2].Username != second.Username {
		t.Fatal("conversation ordering/author mismatch")
	}
	other.request("GET", path, nil, 404)
	if len(app.Handler.Files.AttachmentUsage(f.ID)) != 1 {
		t.Fatal("review attachment usage missing")
	}
	admin.request("DELETE", fmt.Sprintf("/api/admin/attachments/%d", f.ID), nil, 409)
	// Closed assignments can still be discussed, hidden assignments stay private.
	app.DB.Model(&a).Update("status", "closed")
	owner.request("POST", path, gin.H{"content": "After deadline"}, 201)
	status("pending")
	app.DB.Model(&a).Update("status", "hidden")
	owner.request("GET", path, nil, 404)
	owner.request("POST", path, gin.H{"content": "hidden"}, 404)
	owner.request("GET", fmt.Sprintf("/api/attachments/%d/download", f.ID), nil, 404)
	admin.request("GET", path, nil, 200)
	app.DB.Model(&a).Update("status", "published")
	// A new submission has its own unreviewed conversation.
	fresh := decode[model.Submission](t, owner.request("POST", base+"/submissions", gin.H{"content": "new answer"}, 201))
	if fresh.ReviewStatus != "submitted" {
		t.Fatal("new submission inherited state")
	}
	freshPath := fmt.Sprintf("/api/submissions/%d/reviews", fresh.ID)
	if len(decode[[]model.ReviewMessage](t, owner.request("GET", freshPath, nil, 200))) != 0 {
		t.Fatal("conversation leaked into new submission")
	}
	owner.request("POST", fmt.Sprintf("/api/submissions/%d/withdraw", fresh.ID), nil, 200)
	owner.request("POST", fmt.Sprintf("/api/submissions/%d/withdraw", s.ID), nil, 200)
	admin.request("POST", path, gin.H{"content": "withdrawn"}, 409)
	owner.request("POST", path, gin.H{"content": "withdrawn"}, 409)
	owner.request("GET", path, nil, 200)
	// Forced attachment deletion detaches its review references without deleting the conversation.
	forced := admin.send(admin.init("temporary-review.txt", []byte("temporary review"), 0), []byte("temporary review"))
	// Use another active submission, since the first one has now been withdrawn.
	active := model.Submission{AssignmentID: a.ID, UserID: second.ID, Content: "other user's answer"}
	app.DB.Create(&active)
	activePath := fmt.Sprintf("/api/submissions/%d/reviews", active.ID)
	admin.request("POST", activePath, gin.H{"attachment_ids": []uint{forced.ID, f.ID}}, 201)
	admin.request("DELETE", fmt.Sprintf("/api/admin/attachments/%d?force=true", forced.ID), nil, 200)
	kept := decode[[]model.ReviewMessage](t, admin.request("GET", activePath, nil, 200))
	if len(kept) != 1 || len(kept[0].Attachments) != 1 || kept[0].Attachments[0].ID != f.ID {
		t.Fatal("force delete damaged conversation")
	}
	// Deletion removes messages and links, but preserves files referenced by another conversation.
	admin.request("DELETE", fmt.Sprintf("/api/admin/submissions/%d?files=true", s.ID), nil, 200)
	var count int64
	app.DB.Model(&model.ReviewMessage{}).Where("submission_id = ?", s.ID).Count(&count)
	if count != 0 {
		t.Fatal("orphaned messages")
	}
	app.DB.Model(&model.ReviewAttachment{}).Where("review_message_id = ?", m.ID).Count(&count)
	if count != 0 {
		t.Fatal("orphaned review attachment links")
	}
	app.DB.Model(&model.Attachment{}).Where("id IN ?", []uint{f.ID, replyFile.ID}).Count(&count)
	if count != 1 {
		t.Fatal("shared review file should survive deletion")
	}
	app.DB.Model(&model.Attachment{}).Where("id = ?", replyFile.ID).Count(&count)
	if count != 0 {
		t.Fatal("unreferenced reply file not removed")
	}
	admin.request("DELETE", fmt.Sprintf("/api/admin/submissions/%d?files=true", active.ID), nil, 200)
	app.DB.Model(&model.Attachment{}).Where("id = ?", f.ID).Count(&count)
	if count != 0 {
		t.Fatal("last review file reference not cleaned up")
	}
	owner.request("GET", path, nil, 404)
}
