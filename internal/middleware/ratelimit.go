// Package middleware provides HTTP middleware components for the webpage analyzer service
package middleware

import (
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"webpage-analyzer/internal/observability"

	"golang.org/x/time/rate"
)

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimiter implements per-IP rate limiting for HTTP requests
type RateLimiter struct {
	visitors map[string]*visitor
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
}

// NewRateLimiter creates a new rate limiter with the specified requests per minute and burst capacity
func NewRateLimiter(requestsPerMinute, burst int) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate.Limit(requestsPerMinute) / 60,
		burst:    burst,
	}

	log.Printf("[INFO] Rate limiter initialized: %d requests/minute, burst: %d", requestsPerMinute, burst)

	go rl.cleanupVisitors()

	return rl
}

func (rl *RateLimiter) getVisitor(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	if !exists {
		limiter := rate.NewLimiter(rl.rate, rl.burst)
		rl.visitors[ip] = &visitor{
			limiter:  limiter,
			lastSeen: time.Now(),
		}
		log.Printf("[INFO] New visitor registered: %s (total active: %d)", ip, len(rl.visitors))

		// Update active visitors metric
		observability.UpdateActiveVisitors(len(rl.visitors))

		return limiter
	}

	v.lastSeen = time.Now()
	return v.limiter
}

func (rl *RateLimiter) cleanupVisitors() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	log.Println("[INFO] Visitor cleanup goroutine started")

	for range ticker.C {
		rl.mu.Lock()
		initialCount := len(rl.visitors)
		cleaned := 0

		for ip, v := range rl.visitors {
			if time.Since(v.lastSeen) > 10*time.Minute {
				delete(rl.visitors, ip)
				cleaned++
			}
		}

		currentCount := len(rl.visitors)
		rl.mu.Unlock()

		if cleaned > 0 {
			log.Printf("[INFO] Cleaned up %d inactive visitors (active: %d -> %d)", cleaned, initialCount, currentCount)

			// Update active visitors metric
			observability.UpdateActiveVisitors(currentCount)
		}
	}
}

// Limit is an HTTP middleware that enforces rate limiting based on client IP address
func (rl *RateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getIP(r)
		limiter := rl.getVisitor(ip)

		if !limiter.Allow() {
			log.Printf("[WARN] Rate limit exceeded for IP: %s, Method: %s, Path: %s", ip, r.Method, r.URL.Path)

			// Record rate limit violation
			observability.RecordRateLimitExceeded(ip)

			http.Error(w, "Rate limit exceeded. Please try again later.", http.StatusTooManyRequests)
			return
		}

		log.Printf("[INFO] Request allowed: IP: %s, Method: %s, Path: %s", ip, r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func getIP(r *http.Request) string {
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		ips := parseForwardedFor(forwarded)
		if len(ips) > 0 {
			log.Printf("[DEBUG] IP from X-Forwarded-For: %s", ips[0])
			return ips[0]
		}
	}

	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		log.Printf("[DEBUG] IP from X-Real-IP: %s", realIP)
		return realIP
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		log.Printf("[WARN] Failed to parse RemoteAddr '%s': %v, using as-is", r.RemoteAddr, err)
		return r.RemoteAddr
	}
	log.Printf("[DEBUG] IP from RemoteAddr: %s", ip)
	return ip
}

func parseForwardedFor(header string) []string {
	var ips []string
	for i := 0; i < len(header); {
		end := i
		for end < len(header) && header[end] != ',' {
			end++
		}

		ip := header[i:end]
		for ip != "" && ip[0] == ' ' {
			ip = ip[1:]
		}
		for ip != "" && ip[len(ip)-1] == ' ' {
			ip = ip[:len(ip)-1]
		}

		if ip != "" {
			ips = append(ips, ip)
		}

		i = end + 1
	}
	return ips
}
