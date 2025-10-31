package middlewares

import (
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Kheav-Kienghok/scholarship_portal/internal/utils"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type visitor struct {
	limiter     *rate.Limiter
	lastSeen    time.Time
	violations  int
	lockedUntil time.Time
}

type RateLimiter struct {
	mu              sync.Mutex
	visitors        map[string]*visitor
	rps             float64
	burst           int
	ttl             time.Duration
	maxViolations   int
	lockoutDuration time.Duration
}

// NewRateLimiter initializes a new rate limiter store.
func NewRateLimiter(rps float64, burst int, ttl time.Duration) *RateLimiter {
	if rps <= 0 {
		rps = 5
	}
	if burst <= 0 {
		burst = 10
	}
	if ttl <= 0 {
		ttl = 3 * time.Minute
	}

	rl := &RateLimiter{
		visitors:        make(map[string]*visitor),
		rps:             rps,
		burst:           burst,
		ttl:             ttl,
		maxViolations:   2,
		lockoutDuration: 3 * time.Hour,
	}
	go rl.cleanupLoop()
	return rl
}

// Middleware returns a Gin handler that enforces rate limiting per client IP.
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := clientIP(c)

		rl.mu.Lock()
		v := rl.getOrCreateVisitor(ip)

		// Check if IP is locked out
		if time.Now().Before(v.lockedUntil) {
			_ = time.Until(v.lockedUntil)
			rl.mu.Unlock()
			utils.JSONIndent(c, 429, "Too many rate limit violations. Account locked.", nil)
			c.Abort()
			return
		}

		limiter := v.limiter
		rl.mu.Unlock()

		allowed := limiter.Allow()

		if !allowed {
			rl.mu.Lock()
			v.violations++
			if v.violations >= rl.maxViolations {
				v.lockedUntil = time.Now().Add(rl.lockoutDuration)
				rl.mu.Unlock()
				utils.JSONIndent(c, 429, "Too many rate limit violations. Account locked for 3 hours.", nil)
				c.Abort()
				return
			}
			rl.mu.Unlock()

			c.Header("Retry-After", rl.retryAfter(limiter))
			utils.JSONIndent(c, 429, "Rate limit exceeded", nil)

			c.Abort()
			return
		}

		c.Next()
	}
}

func (rl *RateLimiter) getOrCreateVisitor(ip string) *visitor {
	now := time.Now()
	v, exists := rl.visitors[ip]
	if !exists {
		lim := rate.NewLimiter(rate.Limit(rl.rps), rl.burst)
		v = &visitor{
			limiter:     lim,
			lastSeen:    now,
			violations:  0,
			lockedUntil: time.Time{},
		}
		rl.visitors[ip] = v
	} else {
		v.lastSeen = now
	}
	return v
}

func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rl.ttl)
	defer ticker.Stop()
	for range ticker.C {
		cutoff := time.Now().Add(-rl.ttl)
		rl.mu.Lock()
		for k, v := range rl.visitors {
			// Don't clean up locked IPs
			if v.lastSeen.Before(cutoff) && time.Now().After(v.lockedUntil) {
				delete(rl.visitors, k)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) retryAfter(l *rate.Limiter) string {
	res := l.Reserve()
	if res.OK() {
		delay := res.Delay()
		res.Cancel()
		if delay > 0 {
			secs := int(delay.Round(time.Second) / time.Second)
			return strconv.Itoa(secs)
		}
	}
	return "0"
}

func clientIP(c *gin.Context) string {
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		ip := strings.TrimSpace(parts[0])
		if ip != "" {
			return ip
		}
	}
	if xrip := c.GetHeader("X-Real-IP"); xrip != "" {
		return strings.TrimSpace(xrip)
	}
	ip, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err != nil {
		return c.Request.RemoteAddr
	}
	return ip
}
