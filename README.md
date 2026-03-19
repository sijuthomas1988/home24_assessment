# 🔍 Web Page Analyzer

<div align="center">

**A powerful Go web application that analyzes web pages and provides detailed insights about their structure and content**

[![Go Version](https://img.shields.io/badge/Go-1.16+-00ADD8?style=flat&logo=go)](https://golang.org)
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

## 📋 Prerequisites

- Go 1.16 or higher

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
- 📁 **Pre-configured Dashboard** (`grafana-dashboard.json`)
- ⚙️ **Prometheus Config** (`prometheus.yml`)

**Access Points:**
- 🌐 Web Analyzer: http://localhost:8080
- 📊 Prometheus: http://localhost:9090
- 📈 Grafana: http://localhost:3000 (admin/admin)

## 🔄 CI/CD Pipeline

The project includes a comprehensive CI/CD pipeline using GitHub Actions for automated testing, building, and deployment.

### 🎯 Continuous Integration

**Automated on every push and pull request:**

1. **🔍 Linting & Formatting**
   - golangci-lint with comprehensive checks
   - gofmt validation
   - Code style enforcement

2. **🔒 Security Scanning**
   - Gosec static security analyzer
   - SARIF reports uploaded to GitHub Security
   - Trivy container vulnerability scanning

3. **🧪 Multi-Version Testing**
   - Tests run on Go 1.21, 1.22, and 1.23
   - Race condition detection
   - Code coverage reporting (Codecov integration)

4. **🏗️ Multi-Platform Builds**
   - Linux (amd64, arm64)
   - macOS (amd64, arm64)
   - Windows (amd64)

5. **🐳 Docker Build & Push**
   - Multi-arch images (amd64, arm64)
   - Build cache optimization
   - Automated tagging (branch, PR, SHA)

6. **✅ Integration Tests**
   - Docker Compose stack validation
   - Health check verification
   - End-to-end analysis tests

### 🚀 Continuous Deployment

**Automated on version tags (v*.*.*):**

- Multi-platform binary releases
- Docker image publishing with version tags
- Automated changelog generation
- GitHub Release creation with artifacts
- Checksum generation for binaries

### 📋 Workflow Files

- **[.github/workflows/ci.yml](.github/workflows/ci.yml)** - Main CI pipeline
- **[.github/workflows/release.yml](.github/workflows/release.yml)** - Release automation

### 🎨 Pipeline Visualization

```
┌─────────────┐
│  Git Push   │
└──────┬──────┘
       │
       ├─────▶ Lint ────────────────┐
       ├─────▶ Security Scan ───────┤
       ├─────▶ Test (Go 1.21) ──────┤
       ├─────▶ Test (Go 1.22) ──────┼──▶ All Checks Pass
       ├─────▶ Test (Go 1.23) ──────┤
       ├─────▶ Build Binaries ──────┤
       ├─────▶ Docker Build ─────────┤
       └─────▶ Integration Tests ────┘
                     │
                     ▼
              ┌──────────────┐
              │  Merge/Deploy│
              └──────────────┘
```

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

## 🤝 Contributing

Contributions are welcome! Please ensure:

1. ✅ All tests pass (`make test`)
2. ✅ Code is formatted (`gofmt -s -w .`)
3. ✅ Linter passes (`golangci-lint run`)
4. ✅ Coverage doesn't decrease
5. ✅ CI pipeline passes

## 📄 License

MIT License - See LICENSE file for details

## 🙏 Acknowledgments

- Go standard library team for excellent HTTP and HTML packages
- Prometheus team for robust metrics framework
- GitHub Actions for CI/CD platform
