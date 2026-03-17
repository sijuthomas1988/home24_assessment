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
│   └── handlers/
│       └── handlers.go           # HTTP request handlers
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

The application uses:
- Standard Go `net/http` package for the web server
- `golang.org/x/net/html` for HTML parsing
- Go templates for rendering the web interface

The analyzer:
- Fetches the target URL with a 30-second timeout
- Parses the HTML document
- Traverses the DOM tree to extract information
- Validates links by sending HEAD requests
- Detects login forms by looking for password fields combined with username/email fields