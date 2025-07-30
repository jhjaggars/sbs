.PHONY: build install clean test fmt lint

# Build the application
build:
	go build -o sbs

# Install to ~/bin (make sure ~/bin is in your PATH)
install: build
	mkdir -p ~/bin
	cp sbs ~/bin/

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