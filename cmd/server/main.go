package main

import (
	"log"
	"net/http"
	"time"

	"webpage-analyzer/internal/handlers"
	"webpage-analyzer/internal/middleware"
	"webpage-analyzer/internal/observability"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	log.Println("=====================================")
	log.Println("  Web Page Analyzer Server")
	log.Println("=====================================")

	rateLimiter := middleware.NewRateLimiter(20, 5)

	mux := http.NewServeMux()
	mux.HandleFunc("/", handlers.HomeHandler)
	mux.Handle("/metrics", promhttp.Handler())

	// Apply middleware chain: Metrics -> RateLimit -> Handlers
	handler := observability.MetricsMiddleware(rateLimiter.Limit(mux))

	// Configure HTTP server with timeouts for security
	server := &http.Server{
		Addr:              ":8080",
		Handler:           handler,
		ReadTimeout:       15 * time.Second,  // Time to read entire request
		WriteTimeout:      60 * time.Second,  // Time to write response (longer for analysis)
		IdleTimeout:       120 * time.Second, // Time to keep idle connections
		ReadHeaderTimeout: 5 * time.Second,   // Time to read headers (prevents slowloris)
		MaxHeaderBytes:    1 << 20,           // 1 MB max header size
	}

	log.Println("[INFO] Server configuration:")
	log.Println("[INFO]   - Address: :8080")
	log.Println("[INFO]   - Rate limit: 20 requests/minute, burst: 5")
	log.Println("[INFO]   - Max response size: 10MB")
	log.Println("[INFO]   - Concurrent link workers: 10")
	log.Println("[INFO]   - Read timeout: 15s")
	log.Println("[INFO]   - Write timeout: 60s")
	log.Println("[INFO]   - Idle timeout: 120s")
	log.Println("[INFO] Endpoints:")
	log.Println("[INFO]   - Web Interface: http://localhost:8080")
	log.Println("[INFO]   - Prometheus Metrics: http://localhost:8080/metrics")
	log.Println("[INFO] Server ready and listening")
	log.Println("=====================================")

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("[FATAL] Server failed to start: %v", err)
	}
}
