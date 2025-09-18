#!/bin/bash

# End-to-End Test Runner for SBS
# This script sets up the environment and runs comprehensive E2E tests

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check prerequisites
check_prerequisites() {
    print_info "Checking prerequisites..."
    
    # Check if git is available
    if ! command -v git &> /dev/null; then
        print_error "git is required for E2E tests"
        exit 1
    fi
    
    # Check if tmux is available
    if ! command -v tmux &> /dev/null; then
        print_warning "tmux not found - some tests may be skipped"
    fi
    
    # Check if jq is available (optional)
    if ! command -v jq &> /dev/null; then
        print_warning "jq not found - JSON output won't be pretty-printed"
    fi
    
    print_success "Prerequisites checked"
}

# Build the binary
build_binary() {
    print_info "Building SBS binary..."
    
    if make build; then
        print_success "Binary built successfully"
    else
        print_error "Failed to build binary"
        exit 1
    fi
}

# Run the E2E tests
run_e2e_tests() {
    print_info "Running end-to-end tests..."
    
    # Set test environment
    export E2E_TESTS=1
    export GO_TEST_TIMEOUT=10m
    
    # Run tests with verbose output
    if go test -tags=e2e -v -timeout=${GO_TEST_TIMEOUT} ./e2e_test.go; then
        print_success "All E2E tests passed!"
    else
        print_error "Some E2E tests failed"
        return 1
    fi
}

# Clean up any leftover test resources
cleanup() {
    print_info "Cleaning up test resources..."
    
    # Try to clean any leftover SBS sessions
    if [ -f "./sbs" ]; then
        ./sbs clean --force > /dev/null 2>&1 || true
    fi
    
    # Kill any leftover tmux sessions that might be from tests
    tmux list-sessions 2>/dev/null | grep "sbs-test" | cut -d: -f1 | xargs -I {} tmux kill-session -t {} 2>/dev/null || true
    
    print_success "Cleanup completed"
}

# Main execution
main() {
    echo "======================================"
    echo "       SBS End-to-End Test Runner     "
    echo "======================================"
    echo
    
    # Trap to ensure cleanup on exit
    trap cleanup EXIT
    
    # Run all steps
    check_prerequisites
    echo
    
    build_binary
    echo
    
    run_e2e_tests
    echo
    
    print_success "E2E test run completed successfully!"
}

# Handle command line arguments
case "${1:-}" in
    "help"|"-h"|"--help")
        echo "Usage: $0 [options]"
        echo
        echo "Options:"
        echo "  help, -h, --help    Show this help message"
        echo "  clean               Only run cleanup"
        echo "  build               Only build binary"
        echo "  test                Only run tests (requires pre-built binary)"
        echo
        echo "Environment variables:"
        echo "  E2E_TESTS=1         Enable E2E tests (set automatically)"
        echo "  GO_TEST_TIMEOUT     Timeout for tests (default: 10m)"
        echo
        exit 0
        ;;
    "clean")
        cleanup
        exit 0
        ;;
    "build")
        check_prerequisites
        build_binary
        exit 0
        ;;
    "test")
        run_e2e_tests
        exit $?
        ;;
    "")
        main
        ;;
    *)
        print_error "Unknown option: $1"
        echo "Use '$0 help' for usage information"
        exit 1
        ;;
esac
