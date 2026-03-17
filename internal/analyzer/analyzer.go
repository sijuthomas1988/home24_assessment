package analyzer

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/html"
)

type AnalysisResult struct {
	URL               string
	HTMLVersion       string
	Title             string
	Headings          map[string]int
	InternalLinks     int
	ExternalLinks     int
	InaccessibleLinks int
	HasLoginForm      bool
	Error             string
}

func AnalyzeURL(targetURL string) (*AnalysisResult, error) {
	result := &AnalysisResult{
		URL:      targetURL,
		Headings: make(map[string]int),
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Get(targetURL)
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch URL: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: Unable to access the URL. Please check if the URL is correct and accessible", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to read response body: %v", err)
	}

	doc, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("Failed to parse HTML: %v", err)
	}

	result.HTMLVersion = detectHTMLVersion(string(body))

	var links []string
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch n.Data {
			case "title":
				if n.FirstChild != nil {
					result.Title = n.FirstChild.Data
				}
			case "h1", "h2", "h3", "h4", "h5", "h6":
				result.Headings[n.Data]++
			case "a":
				for _, attr := range n.Attr {
					if attr.Key == "href" {
						links = append(links, attr.Val)
						break
					}
				}
			case "form":
				if isLoginForm(n) {
					result.HasLoginForm = true
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse target URL: %v", err)
	}

	result.InternalLinks, result.ExternalLinks, result.InaccessibleLinks = categorizeLinks(links, parsedURL)

	return result, nil
}

func detectHTMLVersion(htmlContent string) string {
	htmlContent = strings.ToLower(htmlContent)

	if strings.Contains(htmlContent, "<!doctype html>") {
		return "HTML5"
	}

	if strings.Contains(htmlContent, "html 4.01") {
		if strings.Contains(htmlContent, "strict") {
			return "HTML 4.01 Strict"
		} else if strings.Contains(htmlContent, "transitional") {
			return "HTML 4.01 Transitional"
		} else if strings.Contains(htmlContent, "frameset") {
			return "HTML 4.01 Frameset"
		}
		return "HTML 4.01"
	}

	if strings.Contains(htmlContent, "xhtml 1.0") {
		if strings.Contains(htmlContent, "strict") {
			return "XHTML 1.0 Strict"
		} else if strings.Contains(htmlContent, "transitional") {
			return "XHTML 1.0 Transitional"
		} else if strings.Contains(htmlContent, "frameset") {
			return "XHTML 1.0 Frameset"
		}
		return "XHTML 1.0"
	}

	if strings.Contains(htmlContent, "xhtml 1.1") {
		return "XHTML 1.1"
	}

	return "Unknown or HTML5 (no explicit DOCTYPE)"
}

func isLoginForm(n *html.Node) bool {
	hasPasswordField := false
	hasUsernameField := false

	var checkFields func(*html.Node)
	checkFields = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "input" {
			inputType := ""
			inputName := ""

			for _, attr := range node.Attr {
				if attr.Key == "type" {
					inputType = strings.ToLower(attr.Val)
				}
				if attr.Key == "name" {
					inputName = strings.ToLower(attr.Val)
				}
			}

			if inputType == "password" {
				hasPasswordField = true
			}

			if inputType == "text" || inputType == "email" || inputType == "" {
				if strings.Contains(inputName, "user") ||
					strings.Contains(inputName, "email") ||
					strings.Contains(inputName, "login") {
					hasUsernameField = true
				}
			}
		}

		for c := node.FirstChild; c != nil; c = c.NextSibling {
			checkFields(c)
		}
	}

	checkFields(n)
	return hasPasswordField && hasUsernameField
}

func categorizeLinks(links []string, baseURL *url.URL) (internal, external, inaccessible int) {
	client := &http.Client{
		Timeout: 5 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	checkedURLs := make(map[string]bool)

	for _, link := range links {
		link = strings.TrimSpace(link)
		if link == "" || strings.HasPrefix(link, "#") ||
			strings.HasPrefix(link, "javascript:") ||
			strings.HasPrefix(link, "mailto:") {
			continue
		}

		parsedLink, err := url.Parse(link)
		if err != nil {
			continue
		}

		absoluteURL := baseURL.ResolveReference(parsedLink)

		if absoluteURL.Host == baseURL.Host {
			internal++
		} else {
			external++
		}

		urlStr := absoluteURL.String()
		if _, checked := checkedURLs[urlStr]; checked {
			continue
		}
		checkedURLs[urlStr] = true

		resp, err := client.Head(urlStr)
		if err != nil {
			inaccessible++
			continue
		}
		resp.Body.Close()

		if resp.StatusCode >= 400 {
			inaccessible++
		}
	}

	return
}