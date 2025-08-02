.PHONY: build install clean test fmt lint

# Build the application
build:
	go build -o sbs

# Install using go install
install:
	go install

# Clean build artifacts
clean:
	rm -f sbs

# Run tests
test:
	go test ./...

# Format code
fmt:
	go fmt ./...

# Lint code (requires golangci-lint)
lint:
	golangci-lint run

# Run with example
demo:
	./sbs --help

# Development build with race detection
dev:
	go build -race -o sbs

# Update dependencies
deps:
	go mod tidy
	go mod download