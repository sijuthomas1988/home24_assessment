package analyzer

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func TestDetectHTMLVersion(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "HTML5 doctype",
			content:  "<!DOCTYPE html><html></html>",
			expected: "HTML5",
		},
		{
			name:     "HTML5 lowercase",
			content:  "<!doctype html><html></html>",
			expected: "HTML5",
		},
		{
			name:     "HTML 4.01 Strict",
			content:  `<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 4.01//EN" "http://www.w3.org/TR/html4/strict.dtd">`,
			expected: "HTML 4.01 Strict",
		},
		{
			name:     "HTML 4.01 Transitional",
			content:  `<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 4.01 Transitional//EN">`,
			expected: "HTML 4.01 Transitional",
		},
		{
			name:     "XHTML 1.0 Strict",
			content:  `<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Strict//EN">`,
			expected: "XHTML 1.0 Strict",
		},
		{
			name:     "XHTML 1.1",
			content:  `<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.1//EN">`,
			expected: "XHTML 1.1",
		},
		{
			name:     "No doctype",
			content:  "<html></html>",
			expected: "Unknown or HTML5 (no explicit DOCTYPE)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectHTMLVersion(tt.content)
			if result != tt.expected {
				t.Errorf("detectHTMLVersion() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsLoginForm(t *testing.T) {
	tests := []struct {
		name     string
		htmlStr  string
		expected bool
	}{
		{
			name: "Valid login form with username and password",
			htmlStr: `
				<form>
					<input type="text" name="username">
					<input type="password" name="password">
				</form>
			`,
			expected: true,
		},
		{
			name: "Valid login form with email and password",
			htmlStr: `
				<form>
					<input type="email" name="email">
					<input type="password" name="pass">
				</form>
			`,
			expected: true,
		},
		{
			name: "Form with only password field",
			htmlStr: `
				<form>
					<input type="password" name="password">
				</form>
			`,
			expected: false,
		},
		{
			name: "Form with only username field",
			htmlStr: `
				<form>
					<input type="text" name="username">
				</form>
			`,
			expected: false,
		},
		{
			name: "Form with unrelated fields",
			htmlStr: `
				<form>
					<input type="text" name="search">
					<input type="submit">
				</form>
			`,
			expected: false,
		},
		{
			name: "Login form with implicit text type",
			htmlStr: `
				<form>
					<input name="user">
					<input type="password">
				</form>
			`,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := html.Parse(strings.NewReader(tt.htmlStr))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}

			var formNode *html.Node
			var findForm func(*html.Node)
			findForm = func(n *html.Node) {
				if n.Type == html.ElementNode && n.Data == "form" {
					formNode = n
					return
				}
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					findForm(c)
				}
			}
			findForm(doc)

			if formNode == nil {
				t.Fatal("No form element found in test HTML")
			}

			result := isLoginForm(formNode)
			if result != tt.expected {
				t.Errorf("isLoginForm() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestAnalyzeURL_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		html := `
<!DOCTYPE html>
<html>
<head>
	<title>Test Page</title>
</head>
<body>
	<h1>Main Heading</h1>
	<h2>Sub Heading 1</h2>
	<h2>Sub Heading 2</h2>
	<h3>Sub Sub Heading</h3>

	<a href="/internal">Internal Link</a>
	<a href="https://external.com">External Link</a>
	<a href="#anchor">Anchor Link</a>
	<a href="javascript:void(0)">JS Link</a>

	<form>
		<input type="text" name="username">
		<input type="password" name="password">
	</form>
</body>
</html>
`
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(html))
	}))
	defer server.Close()

	result, err := AnalyzeURL(server.URL)
	if err != nil {
		t.Fatalf("AnalyzeURL() failed: %v", err)
	}

	if result.URL != server.URL {
		t.Errorf("URL = %v, want %v", result.URL, server.URL)
	}

	if result.HTMLVersion != "HTML5" {
		t.Errorf("HTMLVersion = %v, want HTML5", result.HTMLVersion)
	}

	if result.Title != "Test Page" {
		t.Errorf("Title = %v, want 'Test Page'", result.Title)
	}

	expectedHeadings := map[string]int{"h1": 1, "h2": 2, "h3": 1}
	for tag, count := range expectedHeadings {
		if result.Headings[tag] != count {
			t.Errorf("Headings[%s] = %v, want %v", tag, result.Headings[tag], count)
		}
	}

	if result.InternalLinks != 1 {
		t.Errorf("InternalLinks = %v, want 1", result.InternalLinks)
	}

	if result.ExternalLinks != 1 {
		t.Errorf("ExternalLinks = %v, want 1", result.ExternalLinks)
	}

	if !result.HasLoginForm {
		t.Error("HasLoginForm = false, want true")
	}
}

func TestAnalyzeURL_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not Found"))
	}))
	defer server.Close()

	_, err := AnalyzeURL(server.URL)
	if err == nil {
		t.Error("AnalyzeURL() expected error for 404, got nil")
	}

	if !strings.Contains(err.Error(), "HTTP 404") {
		t.Errorf("Error message should contain 'HTTP 404', got: %v", err.Error())
	}
}

func TestAnalyzeURL_InvalidURL(t *testing.T) {
	_, err := AnalyzeURL("not-a-valid-url://test")
	if err == nil {
		t.Error("AnalyzeURL() expected error for invalid URL, got nil")
	}
}

func TestAnalyzeURL_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(""))
	}))
	defer server.Close()

	result, err := AnalyzeURL(server.URL)
	if err != nil {
		t.Fatalf("AnalyzeURL() failed: %v", err)
	}

	if result.Title != "" {
		t.Errorf("Title should be empty, got: %v", result.Title)
	}

	if len(result.Headings) != 0 {
		t.Errorf("Headings should be empty, got: %v", result.Headings)
	}
}

func TestCategorizeLinks(t *testing.T) {
	baseURL, _ := url.Parse("https://example.com")

	tests := []struct {
		name             string
		links            []string
		expectedInternal int
		expectedExternal int
	}{
		{
			name:             "Internal and external links",
			links:            []string{"https://example.com/page1", "https://other.com/page"},
			expectedInternal: 1,
			expectedExternal: 1,
		},
		{
			name:             "Relative links",
			links:            []string{"/page1", "/page2", "https://other.com"},
			expectedInternal: 2,
			expectedExternal: 1,
		},
		{
			name:             "Duplicate links",
			links:            []string{"/page1", "/page1", "/page1"},
			expectedInternal: 1,
			expectedExternal: 0,
		},
		{
			name:             "Ignored links",
			links:            []string{"#anchor", "javascript:void(0)", "mailto:test@test.com"},
			expectedInternal: 0,
			expectedExternal: 0,
		},
		{
			name:             "Empty links",
			links:            []string{"", "  ", "\t"},
			expectedInternal: 0,
			expectedExternal: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			internal, external, _ := categorizeLinks(tt.links, baseURL)

			if internal != tt.expectedInternal {
				t.Errorf("Internal links = %v, want %v", internal, tt.expectedInternal)
			}

			if external != tt.expectedExternal {
				t.Errorf("External links = %v, want %v", external, tt.expectedExternal)
			}
		})
	}
}

func TestGetTotalHeadings(t *testing.T) {
	tests := []struct {
		name     string
		headings map[string]int
		expected int
	}{
		{
			name:     "Multiple headings",
			headings: map[string]int{"h1": 1, "h2": 3, "h3": 2},
			expected: 6,
		},
		{
			name:     "Empty headings",
			headings: map[string]int{},
			expected: 0,
		},
		{
			name:     "Single heading type",
			headings: map[string]int{"h1": 5},
			expected: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getTotalHeadings(tt.headings)
			if result != tt.expected {
				t.Errorf("getTotalHeadings() = %v, want %v", result, tt.expected)
			}
		})
	}
}
