.PHONY: test test-verbose test-coverage build run clean deps lint check \
        docker-build docker-run docker-stop docker-clean docker-logs \
        docker-compose-up docker-compose-down docker-compose-logs

# Build the application
build:
	go build -o webpage-analyzer ./cmd/server

# Run the application
run: build
	./webpage-analyzer

# Run all tests
test:
	go test ./internal/...

# Run tests with verbose output
test-verbose:
	go test -v ./internal/...

# Run tests with coverage
test-coverage:
	go test -cover ./internal/...
	go test -coverprofile=coverage.out ./internal/...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Clean build artifacts
clean:
	rm -f webpage-analyzer coverage.out coverage.html
	rm -rf internal/handlers/testdata

# Install dependencies
deps:
	go mod download
	go mod tidy

# Run linter
lint:
	golangci-lint run

# Run all checks (test + lint)
check: test lint

# Docker commands
docker-build:
	docker build -t webpage-analyzer:latest .

docker-run:
	docker run -d --name webpage-analyzer -p 8080:8080 webpage-analyzer:latest

docker-stop:
	docker stop webpage-analyzer || true
	docker rm webpage-analyzer || true

docker-clean: docker-stop
	docker rmi webpage-analyzer:latest || true

docker-logs:
	docker logs -f webpage-analyzer

# Docker Compose commands
docker-compose-up:
	docker-compose up -d

docker-compose-down:
	docker-compose down

docker-compose-logs:
	docker-compose logs -f

docker-compose-rebuild:
	docker-compose down
	docker-compose build --no-cache
	docker-compose up -d
