package handlers

import (
	"html/template"
	"net/http"

	"webpage-analyzer/internal/analyzer"
)

var templates = template.Must(template.ParseGlob("templates/*.html"))

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		templates.ExecuteTemplate(w, "index.html", nil)
		return
	}

	if r.Method == http.MethodPost {
		handleAnalysis(w, r)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func handleAnalysis(w http.ResponseWriter, r *http.Request) {
	url := r.FormValue("url")
	if url == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	result, err := analyzer.AnalyzeURL(url)
	if err != nil {
		data := map[string]interface{}{
			"Error": err.Error(),
		}
		templates.ExecuteTemplate(w, "index.html", data)
		return
	}

	templates.ExecuteTemplate(w, "index.html", result)
}