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

# Run end-to-end tests
e2e:
	@echo "Running end-to-end tests..."
	E2E_TESTS=1 go test -tags=e2e -v ./e2e_test.go

# Run integration tests
integration:
	@echo "Running integration tests..."
	INTEGRATION_TESTS=1 go test -tags=integration -v ./integration_test.go

# Run all tests including e2e and integration
test-all: test integration e2e