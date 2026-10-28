package app

import (
	"fmt"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"homework-station/internal/config"
	"homework-station/internal/model"
)

func TestAssignmentSubmissionRequirements(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("DATA_DIR", t.TempDir())
	t.Setenv("DB_DRIVER", "sqlite")
	t.Setenv("DB_DSN", "")
	t.Setenv("JWT_SECRET", "test-session-secret-32-characters-long")
	application, err := New(config.Load())
	if err != nil {
		t.Fatal(err)
	}
	defer application.Close()
	admin := &testClient{t: t, r: application.Router}
	student := &testClient{t: t, r: application.Router}
	admin.request("POST", "/api/login", gin.H{"username": "admin", "password": "123456"}, 200)
	user := student.register("requirements-user")
	assignment := decode[model.Assignment](t, admin.request("POST", "/api/admin/assignments", gin.H{
		"name": "身份条件作业", "status": "published", "max_size": 1024, "multiple": true,
	}, 200))
	path := fmt.Sprintf("/api/assignments/%d", assignment.ID)
	editPath := fmt.Sprintf("/api/admin/assignments/%d", assignment.ID)
	studentNo, qq := "2026000001", "123456789"
	for _, requireReal := range []bool{false, true} {
		for _, requireQQ := range []bool{false, true} {
			assignment.RequireRealName, assignment.RequireQQ = requireReal, requireQQ
			saved := decode[model.Assignment](t, admin.request("PUT", editPath, assignment, 200))
			if saved.RequireRealName != requireReal || saved.RequireQQ != requireQQ {
				t.Fatal("requirements not saved")
			}
			detail := decode[struct {
				Assignment model.Assignment `json:"assignment"`
			}](t, student.request("GET", path, nil, 200))
			if detail.Assignment.RequireRealName != requireReal || detail.Assignment.RequireQQ != requireQQ {
				t.Fatal("requirements not returned")
			}
			for _, realBound := range []bool{false, true} {
				for _, qqBound := range []bool{false, true} {
					var realValue, qqValue interface{}
					name := ""
					if realBound {
						realValue, name = studentNo, "实名姓名"
					}
					if qqBound {
						qqValue = qq
					}
					if err := application.DB.Model(&model.User{}).Where("id = ?", user.ID).Updates(map[string]interface{}{
						"student_no": realValue, "real_name": name, "qq": qqValue,
					}).Error; err != nil {
						t.Fatal(err)
					}
					want, message := 201, ""
					if requireReal && !realBound {
						want, message = 403, "实名认证"
					} else if requireQQ && !qqBound {
						want, message = 403, "绑定 QQ"
					}
					var before, after int64
					application.DB.Model(&model.Submission{}).Where("assignment_id = ?", assignment.ID).Count(&before)
					result := student.request("POST", path+"/submissions", gin.H{"content": "我的作业"}, want)
					if message != "" && !strings.Contains(result.Body.String(), message) {
						t.Fatal("missing actionable rejection", result.Body.String())
					}
					application.DB.Model(&model.Submission{}).Where("assignment_id = ?", assignment.ID).Count(&after)
					if (want == 403 && after != before) || (want == 201 && after != before+1) {
						t.Fatal("incorrect submission count")
					}
				}
			}
		}
	}
	// Administrators must meet the same requirements when submitting their own work.
	admin.request("POST", path+"/submissions", gin.H{"content": "管理员作业"}, 403)
	// Turning requirements off again restores submission access for unbound accounts.
	assignment.RequireRealName, assignment.RequireQQ = false, false
	admin.request("PUT", editPath, assignment, 200)
	admin.request("POST", path+"/submissions", gin.H{"content": "管理员作业"}, 201)
}
