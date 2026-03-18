package handlers

import (
	"html/template"
	"log"
	"net/http"

	"webpage-analyzer/internal/analyzer"
)

var templates = template.Must(template.ParseGlob("templates/*.html"))

func init() {
	log.Println("[INFO] Templates loaded successfully from templates/*.html")
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		log.Printf("[INFO] Serving home page to %s", r.RemoteAddr)
		err := templates.ExecuteTemplate(w, "index.html", nil)
		if err != nil {
			log.Printf("[ERROR] Failed to render template: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	if r.Method == http.MethodPost {
		handleAnalysis(w, r)
		return
	}

	log.Printf("[WARN] Method not allowed: %s from %s", r.Method, r.RemoteAddr)
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func handleAnalysis(w http.ResponseWriter, r *http.Request) {
	url := r.FormValue("url")
	if url == "" {
		log.Printf("[WARN] Analysis request with empty URL from %s", r.RemoteAddr)
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	log.Printf("[INFO] Analysis request received from %s for URL: %s", r.RemoteAddr, url)

	result, err := analyzer.AnalyzeURL(url)
	if err != nil {
		log.Printf("[ERROR] Analysis failed for URL %s: %v", url, err)
		data := map[string]interface{}{
			"Error": err.Error(),
		}
		renderErr := templates.ExecuteTemplate(w, "index.html", data)
		if renderErr != nil {
			log.Printf("[ERROR] Failed to render error template: %v", renderErr)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	log.Printf("[INFO] Analysis successful for URL: %s, rendering results", url)
	renderErr := templates.ExecuteTemplate(w, "index.html", result)
	if renderErr != nil {
		log.Printf("[ERROR] Failed to render results template: %v", renderErr)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}