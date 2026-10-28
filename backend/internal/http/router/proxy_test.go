package router

import (
	"io"
	"log"
	"net/http/httptest"
	"testing"

	"homework-station/internal/config"
	"homework-station/internal/http/handler"
	"homework-station/internal/http/middleware"

	"github.com/gin-gonic/gin"
)

func TestProxyClientIP(t *testing.T) {
	for _, tc := range []struct {
		name                          string
		proxies                       []string
		peer, forwarded, realIP, want string
	}{
		{name: "disabled", peer: "127.0.0.1:1234", forwarded: "203.0.113.5", realIP: "203.0.113.6", want: "127.0.0.1"},
		{name: "trusted forwarded", proxies: []string{"127.0.0.1"}, peer: "127.0.0.1:1234", forwarded: "203.0.113.5", want: "203.0.113.5"},
		{name: "trusted real IP", proxies: []string{"127.0.0.1"}, peer: "127.0.0.1:1234", realIP: "203.0.113.6", want: "203.0.113.6"},
		{name: "untrusted spoof", proxies: []string{"127.0.0.1"}, peer: "198.51.100.9:1234", forwarded: "203.0.113.5", realIP: "203.0.113.6", want: "198.51.100.9"},
		{name: "proxy chain ignores forged prefix", proxies: []string{"10.0.0.0/24"}, peer: "10.0.0.2:1234", forwarded: "192.0.2.99, 203.0.113.5, 10.0.0.1", want: "203.0.113.5"},
		{name: "IPv6", proxies: []string{"::1"}, peer: "[::1]:1234", forwarded: "2001:db8::1", want: "2001:db8::1"},
		{name: "invalid forwarded", proxies: []string{"127.0.0.1"}, peer: "127.0.0.1:1234", forwarded: "not-an-ip", want: "127.0.0.1"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, err := New(&handler.Handler{}, log.New(io.Discard, "", 0), config.Config{TrustedProxies: tc.proxies})
			if err != nil {
				t.Fatal(err)
			}
			r.GET("/test-ip", func(c *gin.Context) { c.String(200, c.ClientIP()) })
			req := httptest.NewRequest("GET", "/test-ip", nil)
			req.RemoteAddr = tc.peer
			req.Header.Set("X-Forwarded-For", tc.forwarded)
			req.Header.Set("X-Real-IP", tc.realIP)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)
			if rec.Body.String() != tc.want {
				t.Fatalf("got %q, want %q", rec.Body.String(), tc.want)
			}
		})
	}
}

func TestInvalidTrustedProxy(t *testing.T) {
	if _, err := New(&handler.Handler{}, log.New(io.Discard, "", 0), config.Config{TrustedProxies: []string{"invalid"}}); err == nil {
		t.Fatal("invalid trusted proxy accepted")
	}
}

func TestProxyRateLimit(t *testing.T) {
	r, err := New(&handler.Handler{}, log.New(io.Discard, "", 0), config.Config{TrustedProxies: []string{"127.0.0.1"}})
	if err != nil {
		t.Fatal(err)
	}
	r.GET("/test-limit", middleware.NewRateLimiter().Middleware(), func(c *gin.Context) { c.Status(204) })
	for i := 0; i < 22; i++ {
		req := httptest.NewRequest("GET", "/test-limit", nil)
		req.RemoteAddr = "127.0.0.1:1234"
		req.Header.Set("X-Forwarded-For", "203.0.113.5")
		want := 204
		if i == 20 {
			want = 429
		}
		if i == 21 {
			req.Header.Set("X-Forwarded-For", "203.0.113.6")
		}
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != want {
			t.Fatalf("request %d: got %d, want %d", i, rec.Code, want)
		}
	}
}

func TestProxyOriginValidation(t *testing.T) {
	r, err := New(&handler.Handler{}, log.New(io.Discard, "", 0), config.Config{TrustedProxies: []string{"127.0.0.1"}})
	if err != nil {
		t.Fatal(err)
	}
	r.POST("/test-origin", func(c *gin.Context) { c.Status(204) })
	for _, tc := range []struct {
		host, origin string
		want         int
	}{
		{"homework.example.com", "https://homework.example.com", 204},
		{"homework.example.com:8443", "https://homework.example.com:8443", 204},
		{"127.0.0.1:8080", "https://homework.example.com", 403},
		{"homework.example.com", "https://evil.example", 403},
	} {
		req := httptest.NewRequest("POST", "/test-origin", nil)
		req.Host = tc.host
		req.RemoteAddr = "127.0.0.1:1234"
		req.Header.Set("Origin", tc.origin)
		req.Header.Set("X-Forwarded-Host", "evil.example")
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != tc.want {
			t.Fatalf("host %s, origin %s: got %d, want %d", tc.host, tc.origin, rec.Code, tc.want)
		}
	}
}
