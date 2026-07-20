package middleware

import (
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type ipBucket struct {
	tokens    float64
	lastCheck time.Time
}

type RateLimiter struct {
	mu       sync.Mutex
	buckets  map[string]*ipBucket
	rate     float64
	burst    float64
	interval time.Duration
}

func NewRateLimiter(rps float64, burst int) *RateLimiter {
	rl := &RateLimiter{
		buckets:  make(map[string]*ipBucket),
		rate:     rps,
		burst:    float64(burst),
		interval: time.Minute,
	}
	go rl.gc()
	return rl
}

func (rl *RateLimiter) gc() {
	ticker := time.NewTicker(rl.interval)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, b := range rl.buckets {
			if now.Sub(b.lastCheck) > 5*time.Minute {
				delete(rl.buckets, ip)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	b, ok := rl.buckets[ip]
	if !ok {
		b = &ipBucket{tokens: rl.burst, lastCheck: now}
		rl.buckets[ip] = b
	}
	elapsed := now.Sub(b.lastCheck).Seconds()
	b.tokens += elapsed * rl.rate
	if b.tokens > rl.burst {
		b.tokens = rl.burst
	}
	b.lastCheck = now
	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := clientIP(c)
		if !rl.Allow(ip) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests"})
			return
		}
		c.Next()
	}
}

func clientIP(c *gin.Context) string {
	// Prefer the connection's remote address. Only honor X-Forwarded-For
	// if the request came from a trusted proxy CIDR — without this guard
	// an attacker can rotate XFF to bypass per-IP rate limits.
	host, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err != nil {
		host = c.Request.RemoteAddr
	}
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" && isTrustedProxy(host) {
		if i := strings.IndexByte(xff, ','); i >= 0 {
			return strings.TrimSpace(xff[:i])
		}
		return strings.TrimSpace(xff)
	}
	return host
}

// trustedProxyCIDRs is populated from TRUSTED_PROXY_CIDRS (comma-sep).
// Empty by default — XFF is only honored when explicitly configured.
var (
	trustedProxyCIDRsOnce sync.Once
	trustedProxyCIDRs     []*net.IPNet
)

func isTrustedProxy(ip string) bool {
	trustedProxyCIDRsOnce.Do(func() {
		for _, cidr := range strings.Split(os.Getenv("TRUSTED_PROXY_CIDRS"), ",") {
			cidr = strings.TrimSpace(cidr)
			if cidr == "" {
				continue
			}
			if _, ipNet, err := net.ParseCIDR(cidr); err == nil {
				trustedProxyCIDRs = append(trustedProxyCIDRs, ipNet)
			}
		}
	})
	if len(trustedProxyCIDRs) == 0 {
		return false
	}
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return false
	}
	for _, ipNet := range trustedProxyCIDRs {
		if ipNet.Contains(parsed) {
			return true
		}
	}
	return false
}
