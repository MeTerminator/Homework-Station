package router

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/gin-gonic/gin"
)

func TestFrontendRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	const index = "<!doctype html><html>embedded app</html>"
	files := fstest.MapFS{
		"index.html":         {Data: []byte(index)},
		"assets/app-123.js":  {Data: []byte("console.log('embedded');")},
		"assets/app-123.css": {Data: []byte("body{color:red}")},
	}
	router := gin.New()
	router.NoRoute(frontendHandler(files))
	for _, test := range []struct {
		method, path string
		status       int
		body         string
		contentType  string
	}{
		{http.MethodGet, "/", 200, index, "text/html"},
		{http.MethodGet, "/admin/settings", 200, index, "text/html"},
		{http.MethodGet, "/assignments/42", 200, index, "text/html"},
		{http.MethodHead, "/admin/settings", 200, "", "text/html"},
		{http.MethodGet, "/assets/app-123.js", 200, "console.log('embedded');", "javascript"},
		{http.MethodGet, "/assets/app-123.css", 200, "body{color:red}", "text/css"},
		{http.MethodHead, "/assets/app-123.js", 200, "", "javascript"},
		{http.MethodGet, "/assets/missing.js", 404, "", ""},
		{http.MethodGet, "/missing.css", 404, "", ""},
		{http.MethodGet, "/.env", 404, "", ""},
		{http.MethodPost, "/admin/settings", 404, "", ""},
	} {
		t.Run(test.method+" "+test.path, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, httptest.NewRequest(test.method, test.path, nil))
			if recorder.Code != test.status || (test.status == 200 && recorder.Body.String() != test.body) {
				t.Fatalf("got status %d, body %q", recorder.Code, recorder.Body.String())
			}
			if !strings.Contains(recorder.Header().Get("Content-Type"), test.contentType) {
				t.Fatalf("unexpected content type %q", recorder.Header().Get("Content-Type"))
			}
			if test.status == 404 && strings.Contains(recorder.Body.String(), index) {
				t.Fatal("returned SPA HTML for an invalid asset or method")
			}
		})
	}
	for _, path := range []string{"/api", "/api/missing"} {
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))
		if recorder.Code != 404 || !strings.Contains(recorder.Header().Get("Content-Type"), "application/json") {
			t.Fatalf("unknown API path %s did not return JSON 404", path)
		}
	}
}

func TestDevelopmentFrontend(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.NoRoute(frontendHandler(nil))
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/admin", nil))
	if recorder.Code != 404 || !strings.Contains(recorder.Body.String(), "localhost:5173") {
		t.Fatal("development mode did not direct the user to Vite")
	}
}
