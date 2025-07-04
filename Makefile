# Makefile for valkey-ai-tasks project

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt

# Project directories
TEST_DIR=./tests
INTEG_TEST_DIR=$(TEST_DIR)/integration

# Test parameters
COVERAGE_DIR=./coverage
COVERAGE_FILE=$(COVERAGE_DIR)/coverage.out
COVERAGE_HTML=$(COVERAGE_DIR)/coverage.html

# Default filter is empty (run all tests)
filter?=.

# Default verbosity is off
verbose?=

# Set verbosity flag if verbose is set
ifdef verbose
	VERBOSE_FLAG=-v
else
	VERBOSE_FLAG=
endif

.PHONY: all build test integ-test clean fmt tidy coverage lint lint-install

# Default target
all: build test lint

# Build the application
build:
	@echo "Building application..."
	@$(GOBUILD) ./...

# Run all tests
test:
	@echo "Running all tests..."
	@$(GOTEST) $(VERBOSE_FLAG) ./... -run $(filter)

# Run only integration tests
integ-test:
	@echo "Running integration tests..."
	@$(GOTEST) $(VERBOSE_FLAG) ./tests/integration/... -run $(filter)

# Run tests with coverage
coverage:
	@echo "Running tests with coverage..."
	@mkdir -p $(COVERAGE_DIR)
	@$(GOTEST) -coverprofile=$(COVERAGE_FILE) ./...
	@$(GOCMD) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "Coverage report generated at $(COVERAGE_HTML)"

# Format code
fmt:
	@echo "Formatting code..."
	gofumpt -w .
	golines -w --shorten-comments -m 127 .

# Update dependencies
tidy:
	@echo "Updating dependencies..."
	@$(GOMOD) tidy

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(COVERAGE_DIR)
	@find . -type f -name "*.test" -delete

run:
	@echo "Running application..."
	@go run cmd/mcpserver/main.go

# Lint code using golangci-lint
lint:
	@echo "Linting code..."
	@golangci-lint run $(if $(verbose),-v,) $(if $(fix),--fix,)

# Help target
help:
	@echo "Available targets:"
	@echo "  all         : Build, test, and lint the application"
	@echo "  build       : Build the application"
	@echo "  test        : Run all tests"
	@echo "                 Usage: make test [filter=TestName] [verbose=1]"
	@echo "  integ-test  : Run integration tests only"
	@echo "                 Usage: make integ-test [filter=TestName] [verbose=1]"
	@echo "  coverage    : Generate test coverage report"
	@echo "  lint        : Run linters on the code"
	@echo "                 Usage: make lint [verbose=1] [fix=1]"
	@echo "  lint-install: Install golangci-lint"
	@echo "  fmt         : Format code"
	@echo "  tidy        : Update dependencies"
	@echo "  clean       : Clean build artifacts"
	@echo "  help        : Show this help message"
