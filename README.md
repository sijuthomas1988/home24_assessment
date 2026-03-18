# Web Page Analyzer

A Go web application that analyzes web pages and provides detailed information about their structure and content.

## Features

The analyzer provides the following information about any given URL:

- **HTML Version**: Detects the HTML version (HTML5, HTML 4.01, XHTML, etc.)
- **Page Title**: Extracts the page title
- **Headings Analysis**: Counts headings by level (H1-H6)
- **Link Analysis**:
  - Internal links count
  - External links count
  - Inaccessible links count (broken links)
- **Login Form Detection**: Identifies if the page contains a login form

### Performance & Security Features

- **Concurrent Link Checking**: Validates up to 10 links in parallel using goroutines
- **Rate Limiting**: Per-IP rate limiting (20 requests/minute, burst of 5) to prevent abuse
- **Connection Pooling**: Reuses HTTP connections for improved performance
- **Response Size Limits**: Protects against memory exhaustion (10MB max)
- **Comprehensive Timeouts**: Multiple timeout layers for reliability

## Prerequisites

- Go 1.16 or higher

## Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd home24_accessment
```

2. Install dependencies:
```bash
go mod download
```

3. Build the application:
```bash
go build -o webpage-analyzer ./cmd/server
```

## Usage

1. Run the application:
```bash
./webpage-analyzer
```

2. Open your browser and navigate to:
```
http://localhost:8080
```

3. Enter a URL in the form and click "Analyze"

## Example URLs to Test

Try analyzing these URLs:
- https://example.com
- https://golang.org
- https://github.com

## Error Handling

If a URL is not reachable, the application will display:
- The HTTP status code
- A descriptive error message

## Project Structure

```
.
├── cmd/
│   └── server/
│       └── main.go               # Main application entry point
├── internal/
│   ├── analyzer/
│   │   └── analyzer.go           # HTML analysis logic
│   ├── handlers/
│   │   └── handlers.go           # HTTP request handlers
│   └── middleware/
│       └── ratelimit.go          # Rate limiting middleware
├── templates/
│   └── index.html                # Web interface template
├── go.mod                        # Go module dependencies
└── README.md                     # This file
```

## Development

To run the application in development mode:

```bash
go run ./cmd/server
```

## Technical Details

### Dependencies
- Standard Go `net/http` package for the web server
- `golang.org/x/net/html` for HTML parsing
- `golang.org/x/time/rate` for rate limiting
- Go templates for rendering the web interface

### Analyzer Features
- Fetches the target URL with a 30-second timeout
- Limits response body size to 10MB to prevent memory exhaustion
- Parses the HTML document using streaming parser
- Traverses the DOM tree to extract information
- Validates links concurrently (up to 10 workers) by sending HEAD requests
- Detects login forms by looking for password fields combined with username/email fields

### Performance Optimizations
- **HTTP Client Configuration**:
  - Connection pooling with 100 max idle connections
  - 10 max idle connections per host
  - 90-second idle connection timeout
  - Shared transport layer for all requests
- **Timeouts**:
  - Request timeout: 30 seconds
  - Link check timeout: 5 seconds
  - Dial timeout: 10 seconds
  - TLS handshake timeout: 10 seconds
- **Concurrency**: Worker pool pattern for parallel link validation

### Security Features
- **Rate Limiting**: Token bucket algorithm limiting to 20 requests/minute per IP
- **Per-IP Tracking**: Separate rate limits for each client IP
- **Automatic Cleanup**: Removes inactive visitor records after 10 minutes
- **Proxy Support**: Reads `X-Forwarded-For` and `X-Real-IP` headers

### Logging
The application includes comprehensive structured logging at multiple levels:

**Log Levels:**
- `[INFO]` - General informational messages (startup, requests, completions)
- `[WARN]` - Warning messages (rate limits, parsing issues, method errors)
- `[ERROR]` - Error messages (fetch failures, parsing errors, rendering errors)
- `[DEBUG]` - Debug messages (worker activity, IP detection, link checks)
- `[FATAL]` - Critical errors that cause server shutdown

**What Gets Logged:**
- **Server Startup**: Configuration, initialization of components
- **HTTP Requests**: Incoming requests with IP, method, and path
- **Rate Limiting**: New visitors, rate limit violations, cleanup operations
- **Analysis Operations**:
  - URL fetch and status codes
  - Response body sizes
  - HTML parsing results
  - Link extraction and categorization
  - Worker pool activity
  - Concurrent link validation
  - Analysis completion with timing metrics
- **Errors**: All failures with context (URL, operation, error details)
- **Performance Metrics**: Operation durations, worker counts, link statistics

**Example Log Output:**
```
[INFO] Rate limiter initialized: 20 requests/minute, burst: 5
[INFO] Analyzer HTTP clients initialized (MaxIdle: 100, Timeout: 30s)
[INFO] Templates loaded successfully from templates/*.html
[INFO] Server ready and listening on http://localhost:8080
[INFO] Request allowed: IP: 127.0.0.1, Method: POST, Path: /
[INFO] Starting analysis for URL: https://example.com
[INFO] Fetched URL https://example.com with status: 200
[INFO] Read 1256 bytes from https://example.com
[INFO] Analysis completed for https://example.com in 2.3s
```