package main

import (
	"log"
	"net/http"

	"webpage-analyzer/internal/handlers"
	"webpage-analyzer/internal/middleware"
)

func main() {
	log.Println("=====================================")
	log.Println("  Web Page Analyzer Server")
	log.Println("=====================================")

	rateLimiter := middleware.NewRateLimiter(20, 5)

	mux := http.NewServeMux()
	mux.HandleFunc("/", handlers.HomeHandler)

	handler := rateLimiter.Limit(mux)

	log.Println("[INFO] Server configuration:")
	log.Println("[INFO]   - Address: :8080")
	log.Println("[INFO]   - Rate limit: 20 requests/minute, burst: 5")
	log.Println("[INFO]   - Max response size: 10MB")
	log.Println("[INFO]   - Concurrent link workers: 10")
	log.Println("[INFO] Server ready and listening on http://localhost:8080")
	log.Println("=====================================")

	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatalf("[FATAL] Server failed to start: %v", err)
	}
}
