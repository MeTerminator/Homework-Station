package handler

import (
	"crypto/rand"
	"crypto/subtle"
	"fmt"
	"html"
	"math/big"
	"strings"
	"time"

	"homework-station/internal/security"

	"github.com/gin-gonic/gin"
)

type captchaEntry struct {
	Code    string
	Expires time.Time
}

func (h *Handler) CaptchaHandler(c *gin.Context) {
	n, _ := rand.Int(rand.Reader, big.NewInt(10000))
	code := fmt.Sprintf("%04d", n.Int64())
	id := security.RandomHex(24)
	h.captchaMu.Lock()
	for k, v := range h.captchas {
		if time.Now().After(v.Expires) {
			delete(h.captchas, k)
		}
	}
	if len(h.captchas) > 10000 {
		h.captchaMu.Unlock()
		fail(c, 429, "请求过多")
		return
	}
	h.captchas[id] = captchaEntry{code, time.Now().Add(5 * time.Minute)}
	h.captchaMu.Unlock()
	var b strings.Builder
	b.WriteString(`<svg xmlns="http://www.w3.org/2000/svg" width="160" height="52"><rect width="160" height="52" rx="8" fill="#eef2ff"/>`)
	for i := 0; i < 7; i++ {
		n, _ := rand.Int(rand.Reader, big.NewInt(160))
		m, _ := rand.Int(rand.Reader, big.NewInt(160))
		fmt.Fprintf(&b, `<path d="M %d 0 L %d 52" stroke="#a5b4fc" opacity=".7"/>`, n.Int64(), m.Int64())
	}
	for i, r := range code {
		fmt.Fprintf(&b, `<text x="%d" y="36" transform="rotate(%d %d 30)" font-family="monospace" font-size="29" font-weight="bold" fill="#3730a3">%s</text>`, 18+i*34, (i%3-1)*12, 18+i*34, html.EscapeString(string(r)))
	}
	b.WriteString(`</svg>`)
	c.Header("Cache-Control", "no-store")
	c.JSON(200, gin.H{"id": id, "image": "data:image/svg+xml;base64," + base64String([]byte(b.String()))})
}

func (h *Handler) ConsumeCaptcha(id, code string) bool {
	h.captchaMu.Lock()
	defer h.captchaMu.Unlock()
	v, ok := h.captchas[id]
	delete(h.captchas, id)
	return ok && time.Now().Before(v.Expires) && subtle.ConstantTimeCompare([]byte(v.Code), []byte(code)) == 1
}
