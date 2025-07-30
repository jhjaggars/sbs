.PHONY: build install clean test fmt lint

# Build the application
build:
	go build -o work-orchestrator

# Install to ~/bin (make sure ~/bin is in your PATH)
install: build
	mkdir -p ~/bin
	cp work-orchestrator ~/bin/

# Clean build artifacts
clean:
	rm -f work-orchestrator

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
	./work-orchestrator --help

# Development build with race detection
dev:
	go build -race -o work-orchestrator

# Update dependencies
deps:
	go mod tidy
	go mod download