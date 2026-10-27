package middleware

import (
	"sync"
	"time"

	"homework-station/internal/http/response"

	"github.com/gin-gonic/gin"
)

type rateEntry struct {
	N  int
	At time.Time
}
type RateLimiter struct {
	mu      sync.Mutex
	entries map[string]rateEntry
}

func NewRateLimiter() *RateLimiter { return &RateLimiter{entries: map[string]rateEntry{}} }
func (limiter *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		limiter.mu.Lock()
		now := time.Now()
		for k, v := range limiter.entries {
			if now.Sub(v.At) > time.Minute {
				delete(limiter.entries, k)
			}
		}
		k := c.ClientIP()
		v := limiter.entries[k]
		if v.N == 0 {
			v.At = now
		}
		v.N++
		limiter.entries[k] = v
		limiter.mu.Unlock()
		if v.N > 20 {
			response.Error(c, 429, "请求过于频繁，请稍后重试")
			return
		}
		c.Next()
	}
}
