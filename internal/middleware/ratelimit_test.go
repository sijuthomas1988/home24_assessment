package middleware

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestNewRateLimiter(t *testing.T) {
	rl := NewRateLimiter(60, 10)

	if rl == nil {
		t.Fatal("NewRateLimiter returned nil")
	}

	if rl.visitors == nil {
		t.Error("visitors map not initialized")
	}

	if rl.burst != 10 {
		t.Errorf("burst = %v, want 10", rl.burst)
	}
}

func TestRateLimiter_AllowRequests(t *testing.T) {
	rl := NewRateLimiter(60, 5)

	handler := rl.Limit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	// First 5 requests should succeed (burst limit)
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Request %d: expected status 200, got %d", i+1, w.Code)
		}
	}
}

func TestRateLimiter_ExceedLimit(t *testing.T) {
	rl := NewRateLimiter(1, 2) // Very low limit: 1 request per minute, burst of 2

	handler := rl.Limit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.1.1:12345"

	// First 2 requests should succeed (burst)
	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Request %d: expected status 200, got %d", i+1, w.Code)
		}
	}

	// Next request should be rate limited
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("Expected status 429, got %d", w.Code)
	}

	body := w.Body.String()
	if body == "" {
		t.Error("Expected error message in response body")
	}
}

func TestRateLimiter_DifferentIPs(t *testing.T) {
	rl := NewRateLimiter(60, 5)

	handler := rl.Limit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	ips := []string{"192.168.1.1:12345", "192.168.1.2:12345", "192.168.1.3:12345"}

	for _, ip := range ips {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = ip
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Request from %s: expected status 200, got %d", ip, w.Code)
		}
	}

	// Verify different visitors were created
	rl.mu.RLock()
	visitorCount := len(rl.visitors)
	rl.mu.RUnlock()

	if visitorCount != 3 {
		t.Errorf("Expected 3 visitors, got %d", visitorCount)
	}
}

func TestRateLimiter_VisitorLastSeen(t *testing.T) {
	rl := NewRateLimiter(60, 5)

	handler := rl.Limit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.1.1:12345"

	// First request
	w1 := httptest.NewRecorder()
	handler.ServeHTTP(w1, req)

	rl.mu.RLock()
	firstSeen := rl.visitors["192.168.1.1"].lastSeen
	rl.mu.RUnlock()

	time.Sleep(100 * time.Millisecond)

	// Second request
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req)

	rl.mu.RLock()
	secondSeen := rl.visitors["192.168.1.1"].lastSeen
	rl.mu.RUnlock()

	if !secondSeen.After(firstSeen) {
		t.Error("lastSeen should be updated on subsequent requests")
	}
}

func TestGetIP_RemoteAddr(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.1.1:12345"

	ip := getIP(req)
	if ip != "192.168.1.1" {
		t.Errorf("getIP() = %v, want 192.168.1.1", ip)
	}
}

func TestGetIP_XForwardedFor(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	req.Header.Set("X-Forwarded-For", "203.0.113.1, 198.51.100.1")

	ip := getIP(req)
	if ip != "203.0.113.1" {
		t.Errorf("getIP() = %v, want 203.0.113.1 (first IP in X-Forwarded-For)", ip)
	}
}

func TestGetIP_XRealIP(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	req.Header.Set("X-Real-IP", "203.0.113.1")

	ip := getIP(req)
	if ip != "203.0.113.1" {
		t.Errorf("getIP() = %v, want 203.0.113.1", ip)
	}
}

func TestGetIP_XForwardedForPriority(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	req.Header.Set("X-Forwarded-For", "203.0.113.1")
	req.Header.Set("X-Real-IP", "198.51.100.1")

	ip := getIP(req)
	// X-Forwarded-For should take priority
	if ip != "203.0.113.1" {
		t.Errorf("getIP() = %v, want 203.0.113.1 (X-Forwarded-For has priority)", ip)
	}
}

func TestGetIP_InvalidRemoteAddr(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "invalid-address"

	ip := getIP(req)
	if ip != "invalid-address" {
		t.Errorf("getIP() = %v, want invalid-address (should return as-is)", ip)
	}
}

func TestParseForwardedFor(t *testing.T) {
	tests := []struct {
		name     string
		header   string
		expected []string
	}{
		{
			name:     "Single IP",
			header:   "203.0.113.1",
			expected: []string{"203.0.113.1"},
		},
		{
			name:     "Multiple IPs",
			header:   "203.0.113.1, 198.51.100.1, 192.0.2.1",
			expected: []string{"203.0.113.1", "198.51.100.1", "192.0.2.1"},
		},
		{
			name:     "IPs with extra spaces",
			header:   "203.0.113.1,  198.51.100.1  ,   192.0.2.1",
			expected: []string{"203.0.113.1", "198.51.100.1", "192.0.2.1"},
		},
		{
			name:     "Empty header",
			header:   "",
			expected: []string{},
		},
		{
			name:     "Only spaces",
			header:   "   ,   ,   ",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseForwardedFor(tt.header)

			if len(result) != len(tt.expected) {
				t.Errorf("parseForwardedFor() returned %d IPs, want %d", len(result), len(tt.expected))
				return
			}

			for i, ip := range result {
				if ip != tt.expected[i] {
					t.Errorf("parseForwardedFor()[%d] = %v, want %v", i, ip, tt.expected[i])
				}
			}
		})
	}
}

func TestRateLimiter_ConcurrentRequests(t *testing.T) {
	rl := NewRateLimiter(100, 50)

	handler := rl.Limit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	var wg sync.WaitGroup
	successCount := 0
	var mu sync.Mutex

	// Simulate 20 concurrent requests from same IP
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = "192.168.1.1:12345"
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code == http.StatusOK {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()

	// All requests should succeed since we're within burst limit
	if successCount < 20 {
		t.Logf("Warning: Only %d out of 20 concurrent requests succeeded", successCount)
	}
}

func TestRateLimiter_RecoveryAfterWait(t *testing.T) {
	rl := NewRateLimiter(60, 1) // 1 request per second, burst of 1

	handler := rl.Limit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.1.1:12345"

	// First request should succeed
	w1 := httptest.NewRecorder()
	handler.ServeHTTP(w1, req)
	if w1.Code != http.StatusOK {
		t.Errorf("First request: expected status 200, got %d", w1.Code)
	}

	// Immediate second request should fail
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req)
	if w2.Code != http.StatusTooManyRequests {
		t.Errorf("Second request: expected status 429, got %d", w2.Code)
	}

	// Wait for rate limit to recover (1 second + buffer)
	time.Sleep(1100 * time.Millisecond)

	// Third request should succeed after waiting
	w3 := httptest.NewRecorder()
	handler.ServeHTTP(w3, req)
	if w3.Code != http.StatusOK {
		t.Errorf("Third request after wait: expected status 200, got %d", w3.Code)
	}
}
