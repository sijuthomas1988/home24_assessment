# 🔍 Web Page Analyzer

<div align="center">

**A powerful Go web application that analyzes web pages and provides detailed insights about their structure and content**

[![Go Version](https://img.shields.io/badge/Go-1.25.0+-00ADD8?style=flat&logo=go)](https://golang.org)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=flat&logo=docker)](https://www.docker.com)
[![Test Coverage](https://img.shields.io/badge/Coverage-85%25-success?style=flat)](README.md#testing)
[![CI/CD](https://img.shields.io/badge/CI%2FCD-GitHub_Actions-2088FF?style=flat&logo=githubactions)](https://github.com)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg?style=flat)](LICENSE)

📖 **[Architecture Documentation](ARCHITECTURE.md)** | 🚀 **[CI/CD Pipeline](.github/workflows/)**

</div>

---

## ✨ Features

The analyzer provides the following information about any given URL:

- 📄 **HTML Version**: Detects the HTML version (HTML5, HTML 4.01, XHTML, etc.)
- 📝 **Page Title**: Extracts the page title
- 📊 **Headings Analysis**: Counts headings by level (H1-H6)
- 🔗 **Link Analysis**:
  - 🏠 Internal links count
  - 🌐 External links count
  - ⚠️ Inaccessible links count (broken links)
- 🔐 **Login Form Detection**: Identifies if the page contains a login form

### ⚡ Performance & Security Features

- 🚀 **Concurrent Link Checking**: Validates up to 10 links in parallel using goroutines
- 🛡️ **Rate Limiting**: Per-IP rate limiting (20 requests/minute, burst of 5) to prevent abuse
- 🔄 **Connection Pooling**: Reuses HTTP connections for improved performance
- 💾 **Response Size Limits**: Protects against memory exhaustion (10MB max)
- ⏱️ **Comprehensive Timeouts**: Multiple timeout layers for reliability

### 🎨 User Interface Features

- ⌛ **Loading Spinner**: Animated spinner overlay during analysis
- 📈 **Progress Indicator**: Step-by-step progress showing current operation:
  1. 📥 Fetching page content
  2. 🔨 Parsing HTML structure
  3. 🔍 Analyzing headings & forms
  4. ✅ Validating links
- 📱 **Responsive Design**: Modern, mobile-friendly interface
- 💬 **Real-time Feedback**: Visual indicators for success, warnings, and errors
- 🎯 **Clean Results Display**: Organized presentation with color-coded badges

### 📊 Observability & Monitoring

- 📈 **Prometheus Metrics**: Production-ready metrics endpoint (`/metrics`)
- 🔍 **Request Tracking**: HTTP request metrics with latency histograms
- ⏱️ **Performance Monitoring**: Analysis duration and throughput metrics
- 🚦 **Rate Limit Metrics**: Track violations and active visitors
- ❌ **Error Tracking**: Comprehensive error metrics by type and operation
- 📝 **Structured Logging**: Multi-level logging (INFO, WARN, ERROR, DEBUG)

## 🎯 Design Philosophy & Key Decisions

This project demonstrates production-grade engineering with explicit trade-off analysis:

**Architecture Decisions**:
- ✅ **Monolith over Microservices**: Simpler operations, faster iteration (split when >20 engineers or different scaling needs)
- ✅ **Synchronous over Async**: Better UX for current scale (migrate to queue at >5K req/min)
- ✅ **In-Memory Rate Limiting**: Sub-microsecond latency vs. Redis at 1-2ms (switch at >10 instances)
- ✅ **Worker Pool (10) over Unbounded Goroutines**: Predictable resources vs. OOM risk on large pages
- ✅ **Templates over SPA**: Faster initial load, better SEO, progressive enhancement

**Why These Matter**:
- Each decision optimized for **current scale** while providing **clear migration paths**
- Trade-offs explicitly documented in [ARCHITECTURE.md](ARCHITECTURE.md#key-design-trade-offs--alternatives-considered)
- Code comments explain **WHY**, not just WHAT

**Scaling Thresholds** (when to evolve architecture):
| Current Approach | Threshold | Next Step |
|------------------|-----------|-----------|
| Synchronous processing | >5,000 req/min | Async queue (RabbitMQ/SQS) |
| In-memory rate limiting | >10 instances | Redis for shared state |
| Fixed worker pool | Background jobs needed | Dynamic pool + job queue |
| No caching | >30% repeat URLs | Redis cache (1hr TTL) |

See [ARCHITECTURE.md](ARCHITECTURE.md) for detailed analysis of alternatives considered and migration paths.

---

## 📋 Prerequisites

- Go 1.25.0 or higher

## 📦 Installation

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

## 🚀 Usage

### 💻 Using Pre-built Binary

1. Run the application:
```bash
./webpage-analyzer
```

2. Open your browser and navigate to:
```
http://localhost:8080
```

3. Enter a URL in the form and click "Analyze"

### 🐳 Using Docker

#### Option 1: Docker Run

```bash
# Build the Docker image
make docker-build
# or
docker build -t webpage-analyzer:latest .

# Run the container
make docker-run
# or
docker run -d --name webpage-analyzer -p 8080:8080 webpage-analyzer:latest

# View logs
make docker-logs
# or
docker logs -f webpage-analyzer

# Stop and remove container
make docker-stop
```

#### Option 2: Docker Compose (Recommended)

```bash
# Start the application
make docker-compose-up
# or
docker-compose up -d

# View logs
make docker-compose-logs
# or
docker-compose logs -f

# Stop the application
make docker-compose-down
# or
docker-compose down

# Rebuild and restart
make docker-compose-rebuild
```

Then open your browser to: `http://localhost:8080`

## 🧪 Example Usage

### 🌐 URLs to Test

Try analyzing these URLs:
- 🌐 https://example.com
- 🐹 https://golang.org
- 🐙 https://github.com

### 📊 Monitoring Endpoints

- **Web Interface**: http://localhost:8080
- **Prometheus Metrics**: http://localhost:8080/metrics

## ❌ Error Handling

If a URL is not reachable, the application will display:
- 🔢 The HTTP status code
- 📝 A descriptive error message

## 📁 Project Structure

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
│   ├── middleware/
│   │   └── ratelimit.go          # Rate limiting middleware
│   └── observability/
│       └── metrics.go            # Prometheus metrics
├── templates/
│   └── index.html                # Web interface template
├── go.mod                        # Go module dependencies
└── README.md                     # This file
```

## 🛠️ Development

To run the application in development mode:

```bash
go run ./cmd/server
```

## 🧪 Testing

The application includes comprehensive test coverage across all packages.

### ▶️ Running Tests

```bash
# Run all tests
make test

# Run tests with verbose output
make test-verbose

# Run tests with coverage report
make test-coverage

# Or using go directly
go test ./internal/...
go test -v ./internal/...
go test -cover ./internal/...
```

### 📊 Test Coverage

Current test coverage:
- ✅ **Analyzer**: 88.8% - Tests for HTML parsing, version detection, link analysis, and login form detection
- ✅ **Handlers**: 81.8% - Tests for HTTP handlers, form validation, and error handling
- ✅ **Middleware**: 85.1% - Tests for rate limiting, IP detection, and concurrent requests

### 📂 Test Structure

```
internal/
├── analyzer/
│   ├── analyzer.go
│   └── analyzer_test.go      # 10 test cases
├── handlers/
│   ├── handlers.go
│   └── handlers_test.go      # 9 test cases
└── middleware/
    ├── ratelimit.go
    └── ratelimit_test.go     # 14 test cases
```

## 🔧 Technical Details

### 📚 Dependencies
- 🌐 Standard Go `net/http` package for the web server
- 🔨 `golang.org/x/net/html` for HTML parsing
- ⏱️ `golang.org/x/time/rate` for rate limiting
- 📊 `github.com/prometheus/client_golang` for metrics and monitoring
- 🎨 Go templates for rendering the web interface

### 🔍 Analyzer Features
- ⏰ Fetches the target URL with a 30-second timeout
- 💾 Limits response body size to 10MB to prevent memory exhaustion
- 📖 Parses the HTML document using streaming parser
- 🌳 Traverses the DOM tree to extract information
- ✅ Validates links concurrently (up to 10 workers) by sending HEAD requests
- 🔐 Detects login forms by looking for password fields combined with username/email fields

### ⚡ Performance Optimizations
- **🔌 HTTP Client Configuration**:
  - 🔄 Connection pooling with 100 max idle connections
  - 🏠 10 max idle connections per host
  - ⏳ 90-second idle connection timeout
  - 🚀 Shared transport layer for all requests
- **⏱️ Timeouts**:
  - 📡 Request timeout: 30 seconds
  - 🔗 Link check timeout: 5 seconds
  - 📞 Dial timeout: 10 seconds
  - 🔒 TLS handshake timeout: 10 seconds
- **🔀 Concurrency**: Worker pool pattern for parallel link validation

### 🛡️ Security Features
- 🚦 **Rate Limiting**: Token bucket algorithm limiting to 20 requests/minute per IP
- 👤 **Per-IP Tracking**: Separate rate limits for each client IP
- 🧹 **Automatic Cleanup**: Removes inactive visitor records after 10 minutes
- 🌐 **Proxy Support**: Reads `X-Forwarded-For` and `X-Real-IP` headers

### 📝 Logging
The application includes comprehensive structured logging at multiple levels:

**📊 Log Levels:**
- ℹ️ `[INFO]` - General informational messages (startup, requests, completions)
- ⚠️ `[WARN]` - Warning messages (rate limits, parsing issues, method errors)
- ❌ `[ERROR]` - Error messages (fetch failures, parsing errors, rendering errors)
- 🐛 `[DEBUG]` - Debug messages (worker activity, IP detection, link checks)
- 💀 `[FATAL]` - Critical errors that cause server shutdown

**📋 What Gets Logged:**
- 🚀 **Server Startup**: Configuration, initialization of components
- 📨 **HTTP Requests**: Incoming requests with IP, method, and path
- 🚦 **Rate Limiting**: New visitors, rate limit violations, cleanup operations
- 🔍 **Analysis Operations**:
  - 📥 URL fetch and status codes
  - 📊 Response body sizes
  - 🔨 HTML parsing results
  - 🔗 Link extraction and categorization
  - 👷 Worker pool activity
  - ✅ Concurrent link validation
  - ⏱️ Analysis completion with timing metrics
- ❌ **Errors**: All failures with context (URL, operation, error details)
- 📈 **Performance Metrics**: Operation durations, worker counts, link statistics

```

## 📊 Observability & Monitoring

The application includes comprehensive observability features for production monitoring and performance analysis.

### 🎯 Prometheus Metrics

The application exposes Prometheus metrics at the `/metrics` endpoint for real-time monitoring and alerting.

**📡 Access Metrics:**
```
http://localhost:8080/metrics
```

### 📈 Available Metrics

**🌐 HTTP Metrics:**
- `http_requests_total` - Total HTTP requests by method, endpoint, and status
- `http_request_duration_seconds` - HTTP request latency histogram
- `http_requests_in_flight` - Current number of requests being processed

**🔍 Analysis Metrics:**
- `analysis_total` - Total page analyses (success/failure)
- `analysis_duration_seconds` - Time taken for analysis operations
- `links_validated_total` - Links checked by type (internal/external/inaccessible)

**🚦 Rate Limiting Metrics:**
- `rate_limit_exceeded_total` - Rate limit violations by IP
- `active_visitors` - Current number of active visitors

**❌ Error Metrics:**
- `errors_total` - Errors by type and operation

### 🔎 Example Prometheus Queries

**Request Rate (per minute):**
```promql
rate(http_requests_total[5m]) * 60
```

**95th Percentile Response Time:**
```promql
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))
```

**Error Rate:**
```promql
rate(errors_total[5m])
```

**Analysis Success Rate:**
```promql
rate(analysis_total{status="success"}[5m]) / rate(analysis_total[5m])
```

**What's Included:**
- 🌐 **Web Analyzer** on port 8080
- 📊 **Prometheus** on port 9090 (metrics collection)
- 📈 **Grafana** on port 3000 (visualization)
- ⚙️ **Prometheus Config** (`prometheus.yml`)

**Access Points:**
- 🌐 Web Analyzer: http://localhost:8080
- 📊 Prometheus: http://localhost:9090

### 📋 Workflow Files

- **[.github/workflows/ci.yml](.github/workflows/ci.yml)** - Main CI/CD pipeline with automated releases

---

## 📦 Versioning & Releases

### Version Scheme

The project uses **calendar-based versioning** with the format:

```
VYY.WeekNumber.Counter
```

**Components:**
- **VYY**: Last 2 digits of the year (e.g., `V26` for 2026)
- **WeekNumber**: ISO week number of the year (01-53)
- **Counter**: Build number for that specific week (resets each week)

**Examples:**
- `V26.12.1` - First build of week 12 in 2026
- `V26.12.2` - Second build of week 12 in 2026
- `V26.13.1` - First build of week 13 in 2026

### Benefits of This Scheme

✅ **Time-based**: Easy to determine when a release was made
✅ **Sequential**: Clear ordering within each week
✅ **Automatic**: No manual version bumping required
✅ **Weekly resets**: Counter resets each week, keeping numbers small

### Automated Releases

Every push to `main` that passes CI automatically:
1. ✅ Runs linting and tests
2. ✅ Builds the binary
3. ✅ Generates a version tag (e.g., `V26.12.1`)
4. ✅ **Generates changelog** from commits
5. ✅ Creates a GitHub Release with changelog
6. ✅ Uploads the compiled binary

### Changelog Generation

Changelogs are **automatically generated** from commit messages, categorized as:

- **Features & Enhancements**: Commits starting with `feat:`, `add:`, `enhance:`
- **Bug Fixes**: Commits starting with `fix:`, `bug:`
- **Documentation**: Commits starting with `doc:`, `docs:`
- **Other Changes**: All other commits

**Commit Message Examples:**
```bash
git commit -m "feat: add concurrent link validation"
git commit -m "fix: resolve memory leak in rate limiter"
git commit -m "docs: update installation instructions"
git commit -m "refactor: optimize HTML parsing"
```

**Changelog Output Example:**
```markdown
## 📝 Changelog

### Features & Enhancements
- feat: add concurrent link validation (a1b2c3d)
- enhance: improve error messages (e4f5g6h)

### Bug Fixes
- fix: resolve memory leak in rate limiter (i7j8k9l)

### Documentation
- docs: update installation instructions (m0n1o2p)

---
Full Changelog: V26.11.5...V26.12.1
```

**Download latest release:**
```bash
# Visit the releases page
https://github.com/yourusername/home24_assessment/releases

# Or use GitHub CLI
gh release download --pattern 'webpage-analyzer-linux-amd64'
```

---

## 📐 Architecture

For detailed architecture documentation, design decisions, and scalability considerations, see:

**📖 [ARCHITECTURE.md](ARCHITECTURE.md)**

### Key Architectural Highlights

**🏗️ Layered Architecture**:
```
Middleware Layer → Handler Layer → Business Logic → External Services
```

**⚡ Performance Optimizations**:
- Connection pooling with 100 max idle connections
- Worker pool pattern for concurrent link validation
- Streaming HTML parser for memory efficiency
- Token bucket rate limiting

**🔒 Security Design**:
- SSRF protection with URL validation
- Resource limits (timeouts, size caps)
- Per-IP rate limiting
- Non-root Docker containers

**📊 Observability**:
- Prometheus metrics at every layer
- Structured logging with multiple levels
- Request tracing and error tracking

**📈 Scalability**:
- Stateless design for horizontal scaling
- Bounded resource usage
- Clear bottleneck identification
- Future-ready architecture
