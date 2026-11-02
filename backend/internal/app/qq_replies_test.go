package app

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"homework-station/internal/config"
	"homework-station/internal/model"
	"homework-station/internal/service"

	"github.com/gin-gonic/gin"
)

func TestQQReplyRoundTrip(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("DATA_DIR", t.TempDir())
	t.Setenv("DB_DRIVER", "sqlite")
	t.Setenv("DB_DSN", "")
	t.Setenv("JWT_SECRET", "test-session-secret-32-characters-long")
	a, err := New(config.Load())
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	qq := "123456789"
	owner := model.User{Username: "qq-owner", QQ: &qq}
	staff := model.User{Username: "qq-staff", Admin: true}
	a.DB.Create(&owner)
	a.DB.Create(&staff)
	assignment := model.Assignment{Name: "reply", Status: "published"}
	a.DB.Create(&assignment)
	submission := model.Submission{UserID: owner.ID, AssignmentID: assignment.ID, ReviewStatus: "reviewed"}
	a.DB.Create(&submission)
	review := model.ReviewMessage{SubmissionID: submission.ID, UserID: staff.ID, Admin: true, Content: "feedback"}
	a.DB.Create(&review)
	const secret = "notification-secret-32-characters-long"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer "+secret {
			t.Error("missing notification auth")
		}
		json.NewEncoder(w).Encode(gin.H{"status": "sent", "bot_id": "10000", "message_id": "-123"})
	}))
	defer server.Close()
	if err := a.Handler.Settings.Update(map[string]string{"qq_notification_secret": secret, "qq_notification_url": server.URL}); err != nil {
		t.Fatal(err)
	}
	if err := a.Handler.Notifications.ProcessNext(context.Background(), false); err != nil {
		t.Fatal(err)
	}
	var link model.QQNotification
	if err := a.DB.First(&link).Error; err != nil || link.ReviewID != review.ID {
		t.Fatal("notification link missing", link, err)
	}
	payload := gin.H{"bot_id": "10000", "qq": qq, "message_id": "456", "reply_message_id": "-123", "content": "已修改\n谢谢 [CQ:at,qq=all]"}
	request := func(auth string, want int) uint {
		t.Helper()
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/api/bot/replies", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", auth)
		res := httptest.NewRecorder()
		a.Router.ServeHTTP(res, req)
		if res.Code != want {
			t.Fatalf("reply = %d want %d: %s", res.Code, want, res.Body.String())
		}
		if want != 201 {
			return 0
		}
		return decode[struct {
			ID uint `json:"id"`
		}](t, res).ID
	}
	request("", 401)
	request("Bearer wrong", 401)
	const auth = "Bearer " + secret
	payload["qq"] = "987654321"
	request(auth, 404)
	payload["qq"] = qq
	payload["reply_message_id"] = "999"
	request(auth, 404)
	payload["reply_message_id"] = "-123"
	payload["content"] = " "
	request(auth, 400)
	payload["content"] = "已修改\n谢谢 [CQ:at,qq=all]"
	id := request(auth, 201)
	a.Handler.Submissions = service.NewSubmissionService(a.DB, a.Handler.Files)
	if request(auth, 201) != id {
		t.Fatal("duplicate reply created")
	}
	var reply model.ReviewMessage
	a.DB.First(&reply, id)
	a.DB.First(&submission, submission.ID)
	if reply.UserID != owner.ID || reply.Admin || reply.NotificationPending || reply.Content != payload["content"] || submission.ReviewStatus != "pending" {
		t.Fatal("reply did not follow platform workflow", reply, submission)
	}
	// Recheck binding, assignment visibility and submission state on every reply.
	payload["message_id"] = "457"
	a.DB.Model(&owner).Update("qq", nil)
	request(auth, 403)
	a.DB.Model(&owner).Update("qq", qq)
	a.DB.Model(&assignment).Update("status", "hidden")
	request(auth, 404)
	a.DB.Model(&assignment).Update("status", "published")
	a.DB.Model(&submission).Update("withdrawn", true)
	request(auth, 409)
	a.DB.Model(&submission).Update("withdrawn", false)
	a.DB.Model(&submission).Update("user_id", staff.ID)
	request(auth, 403)
}
