package app

import (
	"testing"

	"github.com/gin-gonic/gin"

	"homework-station/internal/config"
	"homework-station/internal/model"
)

func TestApplicationsOwnIndependentResources(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := config.Load()
	cfg.DataDir = t.TempDir()
	cfg.Database = config.Database{Driver: "sqlite", Name: "homework"}
	cfg.SessionSecret = "first-test-application-session-secret"
	cfg.BindingSecret = ""
	cfg.CookieSecure = false
	first, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer first.Close()
	cfg.DataDir = t.TempDir()
	cfg.SessionSecret = "second-test-application-session-secret"
	second, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()
	client := &testClient{t: t, r: first.Router}
	user := client.register("isolated-user")
	client.request("GET", "/api/me", nil, 200)
	var count int64
	second.DB.Model(&model.User{}).Count(&count)
	if count != 1 {
		t.Fatal("registration leaked into another application's database")
	}
	foreign := &testClient{t: t, r: second.Router, cookie: client.cookie}
	foreign.request("GET", "/api/me", nil, 401)
	if first.Handler.Files == second.Handler.Files || first.Handler.Identity == second.Handler.Identity {
		t.Fatal("applications share mutable services")
	}
	first.Close()
	independent := &testClient{t: t, r: second.Router}
	independent.request("POST", "/api/login", gin.H{"username": "admin", "password": "123456"}, 200)
	independent.request("GET", "/api/me", nil, 200)
	if user.ID == 0 {
		t.Fatal("missing registration")
	}
}
