package handlers

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
)

func init() {
	// Create test templates directory
	if err := os.MkdirAll("testdata/templates", 0755); err == nil {
		// Create a simple test template
		testTemplate := `<!DOCTYPE html>
<html>
<head><title>Test Template</title></head>
<body>
{{if .Error}}<div class="error">{{.Error}}</div>{{end}}
{{if .URL}}<div>URL: {{.URL}}</div>{{end}}
{{if .Title}}<div>Title: {{.Title}}</div>{{end}}
{{if and (not .Error) .URL}}<div>Analysis Results</div>{{end}}
</body>
</html>`
		os.WriteFile("testdata/templates/index.html", []byte(testTemplate), 0644)
	}

	// Override the templates variable with test template
	if tmpl, err := template.ParseGlob("testdata/templates/*.html"); err == nil {
		templates = tmpl
	}
}

func TestMain(m *testing.M) {
	// Setup
	code := m.Run()

	// Cleanup
	os.RemoveAll("testdata")

	os.Exit(code)
}

func TestHomeHandler_GET(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	HomeHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		t.Errorf("Expected Content-Type to contain 'text/html', got %s", contentType)
	}
}

func TestHomeHandler_POST_EmptyURL(t *testing.T) {
	form := url.Values{}
	form.Add("url", "")

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	HomeHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400 for empty URL, got %d", resp.StatusCode)
	}
}

func TestHomeHandler_POST_InvalidURL(t *testing.T) {
	// Start a test server to handle the analysis
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<!DOCTYPE html><html><head><title>Test</title></head><body></body></html>"))
	}))
	defer testServer.Close()

	form := url.Values{}
	form.Add("url", "not-a-valid-url")

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	HomeHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 (error displayed in template), got %d", resp.StatusCode)
	}

	body := w.Body.String()
	if !strings.Contains(body, "Error") && !strings.Contains(body, "Failed") {
		t.Error("Expected error message in response body")
	}
}

func TestHomeHandler_POST_ValidURL(t *testing.T) {
	// Start a test server to simulate the target website
	targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		html := `<!DOCTYPE html>
<html>
<head><title>Test Page</title></head>
<body>
	<h1>Test Heading</h1>
	<a href="/page">Link</a>
</body>
</html>`
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(html))
	}))
	defer targetServer.Close()

	form := url.Values{}
	form.Add("url", targetServer.URL)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	HomeHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body := w.Body.String()
	if !strings.Contains(body, "Test Page") {
		t.Error("Expected 'Test Page' in response body")
	}

	if !strings.Contains(body, "Analysis Results") {
		t.Error("Expected 'Analysis Results' heading in response")
	}
}

func TestHomeHandler_MethodNotAllowed(t *testing.T) {
	methods := []string{http.MethodPut, http.MethodDelete, http.MethodPatch}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/", nil)
			w := httptest.NewRecorder()

			HomeHandler(w, req)

			resp := w.Result()
			if resp.StatusCode != http.StatusMethodNotAllowed {
				t.Errorf("Expected status 405 for method %s, got %d", method, resp.StatusCode)
			}
		})
	}
}

func TestHandleAnalysis_WithWhitespace(t *testing.T) {
	form := url.Values{}
	form.Add("url", "   ")

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	handleAnalysis(w, req)

	resp := w.Result()
	// Whitespace-only URL still gets processed but returns error in template (status 200)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 (error in template), got %d", resp.StatusCode)
	}

	body := w.Body.String()
	if !strings.Contains(body, "error") && !strings.Contains(body, "Error") && !strings.Contains(body, "Failed") {
		t.Error("Expected error message in response body for whitespace URL")
	}
}

func TestHomeHandler_LargeResponse(t *testing.T) {
	// Test handling of large response (should be within 10MB limit)
	targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<!DOCTYPE html><html><head><title>Large Page</title></head><body>"))

		// Write some content, but keep it under 10MB limit
		for i := 0; i < 1000; i++ {
			w.Write([]byte("<p>This is some content to make the page larger.</p>"))
		}
		w.Write([]byte("</body></html>"))
	}))
	defer targetServer.Close()

	form := url.Values{}
	form.Add("url", targetServer.URL)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	HomeHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 for large but valid response, got %d", resp.StatusCode)
	}
}

func TestHomeHandler_SpecialCharactersInTitle(t *testing.T) {
	targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		html := `<!DOCTYPE html>
<html>
<head><title>Test & "Special" <Characters></title></head>
<body><h1>Content</h1></body>
</html>`
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(html))
	}))
	defer targetServer.Close()

	form := url.Values{}
	form.Add("url", targetServer.URL)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	HomeHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body := w.Body.String()
	// The response should contain escaped or properly handled special characters
	if !strings.Contains(body, "Special") {
		t.Error("Expected title with special characters in response")
	}
}

func TestHomeHandler_Concurrent(t *testing.T) {
	targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		html := `<!DOCTYPE html><html><head><title>Concurrent Test</title></head><body></body></html>`
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(html))
	}))
	defer targetServer.Close()

	// Test concurrent requests
	done := make(chan bool, 5)
	for i := 0; i < 5; i++ {
		go func() {
			form := url.Values{}
			form.Add("url", targetServer.URL)

			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			w := httptest.NewRecorder()

			HomeHandler(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Concurrent request failed with status %d", w.Code)
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 5; i++ {
		<-done
	}
}