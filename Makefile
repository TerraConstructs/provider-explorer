.PHONY: build install clean fmt test test-unit test-integration test-coverage test-short help

# Colors for output
RED    := \033[31m
GREEN  := \033[32m
YELLOW := \033[33m
BLUE   := \033[34m
RESET  := \033[0m

# Default target
.DEFAULT_GOAL := help

build: ## Build the application using GoReleaser snapshot
	@echo "$(BLUE)Building with GoReleaser...$(RESET)"
	@if ! command -v goreleaser >/dev/null 2>&1; then \
		echo "$(RED)Error: goreleaser not found$(RESET)"; \
		echo "Install it with: go install github.com/goreleaser/goreleaser@latest"; \
		exit 1; \
	fi
	goreleaser build --snapshot --clean --single-target

install: ## Install the application to $GOPATH/bin
	@echo "$(BLUE)Installing provider-explorer...$(RESET)"
	go install .

clean: ## Clean build artifacts and cache
	@echo "$(YELLOW)Cleaning build artifacts...$(RESET)"
	rm -rf dist/
	rm -f provider-explorer
	rm -f coverage.out coverage.html
	@echo "$(GREEN)Clean complete$(RESET)"

fmt: ## Format Go code
	@echo "$(BLUE)Formatting Go code...$(RESET)"
	gofmt -s -w .
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
	else \
		echo "$(YELLOW)Note: goimports not found, only gofmt applied$(RESET)"; \
		echo "Install goimports with: go install golang.org/x/tools/cmd/goimports@latest"; \
	fi
	@echo "$(GREEN)Formatting complete$(RESET)"

test: ## Run all tests with race detection
	@echo "$(BLUE)Running all tests...$(RESET)"
	go test -race -v ./...

test-unit: ## Run unit tests only (skip integration tests)
	@echo "$(BLUE)Running unit tests...$(RESET)"
	go test -race -v -short ./...

test-integration: ## Run integration tests only (using teatest)
	@echo "$(BLUE)Running integration tests...$(RESET)"
	go test -race -v ./internal/ui -run "Test.*Integration|TestFullApplicationFlow|TestResponsiveLayout|TestKeyboardNavigation|TestErrorStates|TestLoadingStates|TestViewTransitions"

test-coverage: ## Run tests with coverage report
	@echo "$(BLUE)Running tests with coverage...$(RESET)"
	go test -race -coverprofile=coverage.out ./...
	@if command -v go >/dev/null 2>&1; then \
		go tool cover -html=coverage.out -o coverage.html; \
		echo "$(GREEN)Coverage report generated: coverage.html$(RESET)"; \
	fi

test-short: ## Run quick tests (unit tests only, no race detection)
	@echo "$(BLUE)Running quick tests...$(RESET)"
	go test -v -short ./...

help: ## Show this help message
	@echo "$(GREEN)Provider Explorer - Available Make targets:$(RESET)"
	@echo ""
	@awk 'BEGIN {FS = ":.*?## "}; /^[a-zA-Z_-]+:.*?## / {printf "  $(YELLOW)%-12s$(RESET) %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""
	@echo "$(BLUE)Development workflow:$(RESET)"
	@echo "  1. $(YELLOW)make fmt$(RESET)          - Format code"
	@echo "  2. $(YELLOW)make test-unit$(RESET)    - Run unit tests (fast)"
	@echo "  3. $(YELLOW)make test$(RESET)         - Run all tests"
	@echo "  4. $(YELLOW)make build$(RESET)        - Build binary"
	@echo "  5. $(YELLOW)make install$(RESET)      - Install to Go bin"
	@echo ""
	@echo "$(BLUE)Test options:$(RESET)"
	@echo "  $(YELLOW)make test-unit$(RESET)        - Unit tests only (skip integration)"
	@echo "  $(YELLOW)make test-integration$(RESET) - Integration tests only (teatest)"
	@echo "  $(YELLOW)make test-coverage$(RESET)    - Tests with coverage report"
	@echo "  $(YELLOW)make test-short$(RESET)       - Quick tests (no race detection)"
	@echo ""
	@echo "$(BLUE)Example usage:$(RESET)"
	@echo "  ./dist/provider-explorer_*/provider-explorer fixtures/example-aws"