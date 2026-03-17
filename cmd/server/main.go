package main

import (
	"log"
	"net/http"
	"webpage-analyzer/internal/handlers"
)

func main() {
	http.HandleFunc("/", handlers.HomeHandler)

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
