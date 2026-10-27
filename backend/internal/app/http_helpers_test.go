package app

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"homework-station/internal/model"
)

type testClient struct {
	t      *testing.T
	r      http.Handler
	cookie *http.Cookie
}

func (a *testClient) request(method, path string, body interface{}, want int) *httptest.ResponseRecorder {
	a.t.Helper()
	var b io.Reader
	ct := "application/json"
	switch v := body.(type // A second user proves possession but shares the final physical blob.
	// PNG input becomes a real WebP and is tracked in the user's attachment library.
	// Real-name checks use a reserved example domain and a fully mocked transport.
	// Referenced attachment deletion requires explicit force; shared blob survives.
	) {
	case []byte:
		b = bytes.NewReader(v)
		ct = "application/octet-stream"
	case nil:
	default:
		raw, e := json.Marshal(body)
		if e != nil {
			a.t.Fatal(e)
		}
		b = bytes.NewReader(raw)
	}
	req := httptest.NewRequest(method, path, b)
	req.Header.Set("Content-Type", ct)
	if a.cookie != nil {
		req.AddCookie(a.cookie)
	}
	res := httptest.NewRecorder()
	a.r.ServeHTTP(res, req)
	if res.Code != want {
		a.t.Fatalf("%s %s = %d want %d: %s", method, path, res.Code, want, res.Body.String())
	}
	for _, c := range res.Result().Cookies() {
		if c.Name == "session" {
			a.cookie = c
		}
	}
	return res
}
func decode[T any](t *testing.T, r *httptest.ResponseRecorder) T {
	t.Helper()
	var v T
	if e := json.Unmarshal(r.Body.Bytes(), &v); e != nil {
		t.Fatal(e, r.Body.String())
	}
	return v
}
func (a *testClient) register(name string) model.User {
	a.t.Helper()
	cap := decode[map[string]string](a.t, a.request("GET", "/api/captcha", nil, 200))
	image, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(cap["image"], "data:image/svg+xml;base64,"))
	if err != nil {
		a.t.Fatal(err)
	}
	digits := regexp.MustCompile(`>([0-9])</text>`).FindAllSubmatch(image, -1)
	code := ""
	for _, digit := range digits {
		code += string(digit[1])
	}
	return decode[model.User](a.t, a.request("POST", "/api/register", gin.H{"username": name, "password": "password123", "confirm": "password123", "captcha_id": cap["id"], "captcha": code}, 201))
}
func sha(b []byte) string {
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}
func (a *testClient) init(name string, b []byte, assignment uint) model.Attachment {
	return decode[model.Attachment](a.t, a.request("POST", "/api/uploads", gin.H{"name": name, "size": len(b), "sha256": sha(b), "assignment_id": assignment}, 201))
}
func (a *testClient) send(f model.Attachment, b []byte) model.Attachment {
	for i := 0; i < f.TotalChunks; i++ {
		end := int64(i+1) * f.ChunkSize
		if end > int64(len(b)) {
			end = int64(len(b))
		}
		a.request("PUT", fmt.Sprintf("/api/uploads/%d/chunks/%d", f.ID, i), b[int64(i)*f.ChunkSize:end], 200)
	}
	return decode[model.Attachment](a.t, a.request("POST", fmt.Sprintf("/api/uploads/%d/complete", f.ID), nil, 200))
}

type roundTrip func(*http.Request) (*http.Response, error)

func (f roundTrip) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}
