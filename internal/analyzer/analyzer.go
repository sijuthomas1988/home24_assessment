// Package analyzer provides webpage analysis functionality
package analyzer

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"webpage-analyzer/internal/observability"

	"golang.org/x/net/html"
)

const (
	maxResponseBodySize = 10 * 1024 * 1024
	requestTimeout      = 30 * time.Second
	linkCheckTimeout    = 5 * time.Second
	maxIdleConns        = 100
	maxIdleConnsPerHost = 10
	idleConnTimeout     = 90 * time.Second
	dialTimeout         = 10 * time.Second
	tlsHandshakeTimeout = 10 * time.Second
)

var (
	fetchClient *http.Client
	checkClient *http.Client
)

func init() {
	transport := &http.Transport{
		MaxIdleConns:        maxIdleConns,
		MaxIdleConnsPerHost: maxIdleConnsPerHost,
		IdleConnTimeout:     idleConnTimeout,
		DialContext: (&net.Dialer{
			Timeout:   dialTimeout,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   tlsHandshakeTimeout,
		ExpectContinueTimeout: 1 * time.Second,
		DisableCompression:    false,
	}

	fetchClient = &http.Client{
		Timeout:   requestTimeout,
		Transport: transport,
	}

	checkClient = &http.Client{
		Timeout:   linkCheckTimeout,
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	log.Printf("[INFO] Analyzer HTTP clients initialized (MaxIdle: %d, Timeout: %s)", maxIdleConns, requestTimeout)
}

// AnalysisResult contains the results of analyzing a webpage
type AnalysisResult struct {
	Headings          map[string]int
	URL               string
	HTMLVersion       string
	Title             string
	Error             string
	InternalLinks     int
	ExternalLinks     int
	InaccessibleLinks int
	HasLoginForm      bool
}

// AnalyzeURL fetches and analyzes a webpage, returning detailed information about its structure and links
func AnalyzeURL(targetURL string) (*AnalysisResult, error) {
	log.Printf("[INFO] Starting analysis for URL: %s", targetURL)
	startTime := time.Now()

	result := &AnalysisResult{
		URL:      targetURL,
		Headings: make(map[string]int),
	}

	resp, err := fetchClient.Get(targetURL)
	if err != nil {
		log.Printf("[ERROR] Failed to fetch URL %s: %v", targetURL, err)
		observability.RecordError("fetch_failed", "analyze_url")
		return nil, fmt.Errorf("failed to fetch URL: %v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("[WARN] Failed to close response body: %v", err)
		}
	}()

	log.Printf("[INFO] Fetched URL %s with status: %d", targetURL, resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		log.Printf("[ERROR] Non-OK status for URL %s: HTTP %d", targetURL, resp.StatusCode)
		observability.RecordError("http_error", "analyze_url")
		return nil, fmt.Errorf("HTTP %d: Unable to access the URL. Please check if the URL is correct and accessible", resp.StatusCode)
	}

	limitedReader := io.LimitReader(resp.Body, maxResponseBodySize)
	body, err := io.ReadAll(limitedReader)
	if err != nil {
		log.Printf("[ERROR] Failed to read response body from %s: %v", targetURL, err)
		observability.RecordError("read_failed", "analyze_url")
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	log.Printf("[INFO] Read %d bytes from %s", len(body), targetURL)

	if len(body) >= maxResponseBodySize {
		log.Printf("[ERROR] Response body too large for %s: %d bytes (max: %d)", targetURL, len(body), maxResponseBodySize)
		observability.RecordError("body_too_large", "analyze_url")
		return nil, fmt.Errorf("response body too large (max %d bytes)", maxResponseBodySize)
	}

	doc, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		log.Printf("[ERROR] Failed to parse HTML for %s: %v", targetURL, err)
		observability.RecordError("parse_failed", "analyze_url")
		return nil, fmt.Errorf("failed to parse HTML: %v", err)
	}

	log.Printf("[INFO] Successfully parsed HTML document for %s", targetURL)

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

	log.Printf("[INFO] Extracted %d links from %s", len(links), targetURL)

	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		log.Printf("[ERROR] Failed to parse target URL %s: %v", targetURL, err)
		observability.RecordError("url_parse_failed", "analyze_url")
		return nil, fmt.Errorf("failed to parse target URL: %v", err)
	}

	log.Printf("[INFO] Starting link categorization for %d links from %s", len(links), targetURL)
	result.InternalLinks, result.ExternalLinks, result.InaccessibleLinks = categorizeLinks(links, parsedURL)

	duration := time.Since(startTime)
	log.Printf(
		"[INFO] Analysis completed for %s in %v - Title: '%s', Headings: %d, Internal: %d, External: %d, Inaccessible: %d, LoginForm: %v",
		targetURL, duration, result.Title, getTotalHeadings(result.Headings),
		result.InternalLinks, result.ExternalLinks, result.InaccessibleLinks, result.HasLoginForm,
	)

	return result, nil
}

func getTotalHeadings(headings map[string]int) int {
	total := 0
	for _, count := range headings {
		total += count
	}
	return total
}

func detectHTMLVersion(htmlContent string) string {
	htmlContent = strings.ToLower(htmlContent)

	if strings.Contains(htmlContent, "<!doctype html>") {
		return "HTML5"
	}

	if strings.Contains(htmlContent, "html 4.01") {
		switch {
		case strings.Contains(htmlContent, "strict"):
			return "HTML 4.01 Strict"
		case strings.Contains(htmlContent, "transitional"):
			return "HTML 4.01 Transitional"
		case strings.Contains(htmlContent, "frameset"):
			return "HTML 4.01 Frameset"
		default:
			return "HTML 4.01"
		}
	}

	if strings.Contains(htmlContent, "xhtml 1.0") {
		switch {
		case strings.Contains(htmlContent, "strict"):
			return "XHTML 1.0 Strict"
		case strings.Contains(htmlContent, "transitional"):
			return "XHTML 1.0 Transitional"
		case strings.Contains(htmlContent, "frameset"):
			return "XHTML 1.0 Frameset"
		default:
			return "XHTML 1.0"
		}
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

type linkResult struct {
	isInternal     bool
	isInaccessible bool
}

func categorizeLinks(links []string, baseURL *url.URL) (internal, external, inaccessible int) {
	const maxWorkers = 10

	log.Printf("[INFO] Categorizing %d raw links", len(links))

	uniqueLinks := make(map[string]bool)
	var validLinks []string

	for _, link := range links {
		link = strings.TrimSpace(link)
		if link == "" || strings.HasPrefix(link, "#") ||
			strings.HasPrefix(link, "javascript:") ||
			strings.HasPrefix(link, "mailto:") {
			continue
		}

		parsedLink, err := url.Parse(link)
		if err != nil {
			log.Printf("[WARN] Failed to parse link '%s': %v", link, err)
			continue
		}

		absoluteURL := baseURL.ResolveReference(parsedLink)
		urlStr := absoluteURL.String()

		if !uniqueLinks[urlStr] {
			uniqueLinks[urlStr] = true
			validLinks = append(validLinks, urlStr)

			if absoluteURL.Host == baseURL.Host {
				internal++
			} else {
				external++
			}
		}
	}

	log.Printf("[INFO] Found %d unique valid links (%d internal, %d external)", len(validLinks), internal, external)

	if len(validLinks) == 0 {
		return internal, external, inaccessible
	}

	jobs := make(chan string, len(validLinks))
	results := make(chan linkResult, len(validLinks))

	var wg sync.WaitGroup
	workerCount := maxWorkers
	if len(validLinks) < maxWorkers {
		workerCount = len(validLinks)
	}

	log.Printf("[INFO] Starting %d concurrent workers to validate links", workerCount)
	checkStartTime := time.Now()

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			checked := 0
			for urlStr := range jobs {
				parsedURL, _ := url.Parse(urlStr)
				isInternal := parsedURL.Host == baseURL.Host

				resp, err := checkClient.Head(urlStr)
				isInaccessible := false
				if err != nil {
					log.Printf("[DEBUG] Worker %d: Link check failed for %s: %v", workerID, urlStr, err)
					isInaccessible = true
				} else {
					if closeErr := resp.Body.Close(); closeErr != nil {
						log.Printf("[WARN] Worker %d: Failed to close response body: %v", workerID, closeErr)
					}
					if resp.StatusCode >= 400 {
						log.Printf("[DEBUG] Worker %d: Link inaccessible %s (status: %d)", workerID, urlStr, resp.StatusCode)
						isInaccessible = true
					}
				}

				results <- linkResult{
					isInternal:     isInternal,
					isInaccessible: isInaccessible,
				}
				checked++
			}
			log.Printf("[DEBUG] Worker %d completed checking %d links", workerID, checked)
		}(i)
	}

	for _, link := range validLinks {
		jobs <- link
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(results)
	}()

	for result := range results {
		if result.isInaccessible {
			inaccessible++
		}
	}

	checkDuration := time.Since(checkStartTime)
	log.Printf("[INFO] Link validation completed in %v: %d inaccessible out of %d total links", checkDuration, inaccessible, len(validLinks))

	return internal, external, inaccessible
}
