#!/bin/bash

# FL-GO Test Runner
# This script runs all tests for the FL-GO project

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
COVERAGE_DIR="$PROJECT_ROOT/coverage"
TEST_RESULTS_DIR="$PROJECT_ROOT/test-results"

# Functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

setup_test_env() {
    log_info "Setting up test environment..."
    
    # Create directories
    mkdir -p "$COVERAGE_DIR"
    mkdir -p "$TEST_RESULTS_DIR"
    
    # Set test environment variables
    export FL_GO_TEST_MODE=true
    export FL_GO_LOG_LEVEL=error
}

run_unit_tests() {
    log_info "Running unit tests..."
    
    local test_packages=(
        "./pkg/aggregator/..."
        "./pkg/collaborator/..."
        "./pkg/federation/..."
        "./pkg/monitoring/..."
        "./pkg/security/..."
        "./pkg/cli/..."
    )
    
    local failed_tests=()
    
    for package in "${test_packages[@]}"; do
        log_info "Testing package: $package"
        
        if go test -v -race -coverprofile="$COVERAGE_DIR/$(basename "$package").out" "$package"; then
            log_success "Package $package passed"
        else
            log_error "Package $package failed"
            failed_tests+=("$package")
        fi
    done
    
    if [ ${#failed_tests[@]} -gt 0 ]; then
        log_error "Unit tests failed for: ${failed_tests[*]}"
        return 1
    fi
    
    log_success "All unit tests passed"
}

run_integration_tests() {
    log_info "Running integration tests..."
    
    # Check if integration test directory exists
    if [ -d "$PROJECT_ROOT/tests/integration" ]; then
        cd "$PROJECT_ROOT/tests/integration"
        
        for test_file in *.sh; do
            if [ -f "$test_file" ] && [ -x "$test_file" ]; then
                log_info "Running integration test: $test_file"
                
                if ./"$test_file"; then
                    log_success "Integration test $test_file passed"
                else
                    log_error "Integration test $test_file failed"
                    return 1
                fi
            fi
        done
        
        cd "$PROJECT_ROOT"
    else
        log_warning "Integration test directory not found, skipping"
    fi
    
    log_success "All integration tests passed"
}

run_e2e_tests() {
    log_info "Running end-to-end tests..."
    
    # Check if e2e test directory exists
    if [ -d "$PROJECT_ROOT/tests/e2e" ]; then
        cd "$PROJECT_ROOT/tests/e2e"
        
        for test_file in *.sh; do
            if [ -f "$test_file" ] && [ -x "$test_file" ]; then
                log_info "Running E2E test: $test_file"
                
                if ./"$test_file"; then
                    log_success "E2E test $test_file passed"
                else
                    log_error "E2E test $test_file failed"
                    return 1
                fi
            fi
        done
        
        cd "$PROJECT_ROOT"
    else
        log_warning "E2E test directory not found, skipping"
    fi
    
    log_success "All E2E tests passed"
}

run_security_tests() {
    log_info "Running security tests..."
    
    # Run gosec security scanner
    if command -v gosec &> /dev/null; then
        log_info "Running gosec security scan..."
        
        if gosec -fmt sarif -out "$TEST_RESULTS_DIR/gosec.sarif" -exclude=G103,G104,G301,G302,G304 -exclude-dir=api ./...; then
            log_success "Security scan completed"
        else
            log_warning "Security scan found issues (non-blocking)"
        fi
    else
        log_warning "gosec not found, skipping security scan"
    fi
    
    # Run govulncheck
    if command -v govulncheck &> /dev/null; then
        log_info "Running govulncheck..."
        
        if govulncheck ./...; then
            log_success "Vulnerability check completed"
        else
            log_warning "Vulnerability check found issues (non-blocking)"
        fi
    else
        log_warning "govulncheck not found, skipping vulnerability check"
    fi
}

run_performance_tests() {
    log_info "Running performance benchmarks..."
    
    local benchmark_packages=(
        "./pkg/aggregator/..."
        "./pkg/collaborator/..."
        "./pkg/monitoring/..."
    )
    
    for package in "${benchmark_packages[@]}"; do
        log_info "Running benchmarks for: $package"
        
        if go test -bench=. -benchmem "$package"; then
            log_success "Benchmarks for $package completed"
        else
            log_warning "Benchmarks for $package failed (non-blocking)"
        fi
    done
}

generate_coverage_report() {
    log_info "Generating coverage report..."
    
    # Merge coverage files
    if [ -f "$COVERAGE_DIR/coverage.out" ]; then
        rm "$COVERAGE_DIR/coverage.out"
    fi
    
    # Find all coverage files and merge them
    find "$COVERAGE_DIR" -name "*.out" -exec echo "{}" \; | while read -r file; do
        if [ -f "$file" ]; then
            cat "$file" >> "$COVERAGE_DIR/coverage.out"
        fi
    done
    
    # Generate HTML report
    if [ -f "$COVERAGE_DIR/coverage.out" ]; then
        go tool cover -html="$COVERAGE_DIR/coverage.out" -o "$COVERAGE_DIR/coverage.html"
        log_success "Coverage report generated: $COVERAGE_DIR/coverage.html"
        
        # Show coverage summary
        go tool cover -func="$COVERAGE_DIR/coverage.out" | tail -1
    else
        log_warning "No coverage data found"
    fi
}

run_linting() {
    log_info "Running code linting..."
    
    # Run goimports
    if command -v goimports &> /dev/null; then
        log_info "Running goimports..."
        
        if goimports -l . | grep -q .; then
            log_error "Code formatting issues found"
            goimports -l .
            return 1
        else
            log_success "Code formatting is correct"
        fi
    else
        log_warning "goimports not found, skipping formatting check"
    fi
    
    # Run go vet
    log_info "Running go vet..."
    
    if go vet ./...; then
        log_success "Go vet passed"
    else
        log_error "Go vet found issues"
        return 1
    fi
    
    # Run staticcheck (if available)
    if command -v staticcheck &> /dev/null; then
        log_info "Running staticcheck..."
        
        if staticcheck ./...; then
            log_success "Staticcheck passed"
        else
            log_warning "Staticcheck found issues (non-blocking)"
        fi
    else
        log_warning "staticcheck not found, skipping"
    fi
}

cleanup() {
    log_info "Cleaning up test artifacts..."
    
    # Remove temporary files
    find . -name "*.test" -delete
    find . -name "*.prof" -delete
    
    log_success "Cleanup completed"
}

show_help() {
    echo "FL-GO Test Runner"
    echo ""
    echo "Usage: $0 [OPTION]"
    echo ""
    echo "Options:"
    echo "  all       Run all tests (default)"
    echo "  unit      Run only unit tests"
    echo "  integration  Run only integration tests"
    echo "  e2e       Run only end-to-end tests"
    echo "  security  Run only security tests"
    echo "  performance  Run only performance tests"
    echo "  lint      Run only linting"
    echo "  coverage  Generate coverage report"
    echo "  help      Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0              # Run all tests"
    echo "  $0 unit         # Run only unit tests"
    echo "  $0 security     # Run only security tests"
    echo "  $0 coverage     # Generate coverage report"
}

# Main script
main() {
    local test_type="${1:-all}"
    
    log_info "Starting FL-GO test suite..."
    
    # Setup
    setup_test_env
    
    # Run tests based on type
    case "$test_type" in
        "all")
            run_linting
            run_unit_tests
            run_integration_tests
            run_e2e_tests
            run_security_tests
            run_performance_tests
            generate_coverage_report
            ;;
        "unit")
            run_unit_tests
            ;;
        "integration")
            run_integration_tests
            ;;
        "e2e")
            run_e2e_tests
            ;;
        "security")
            run_security_tests
            ;;
        "performance")
            run_performance_tests
            ;;
        "lint")
            run_linting
            ;;
        "coverage")
            generate_coverage_report
            ;;
        "help"|"-h"|"--help")
            show_help
            exit 0
            ;;
        *)
            log_error "Unknown test type: $test_type"
            show_help
            exit 1
            ;;
    esac
    
    # Cleanup
    cleanup
    
    log_success "Test suite completed successfully!"
}

# Run main function
main "$@"
