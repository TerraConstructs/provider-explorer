.PHONY: build install clean fmt test help

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

test: ## Run tests with race detection
	@echo "$(BLUE)Running tests...$(RESET)"
	go test -race -v ./...

help: ## Show this help message
	@echo "$(GREEN)Provider Explorer - Available Make targets:$(RESET)"
	@echo ""
	@awk 'BEGIN {FS = ":.*?## "}; /^[a-zA-Z_-]+:.*?## / {printf "  $(YELLOW)%-12s$(RESET) %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""
	@echo "$(BLUE)Development workflow:$(RESET)"
	@echo "  1. $(YELLOW)make fmt$(RESET)     - Format code"
	@echo "  2. $(YELLOW)make test$(RESET)    - Run tests"
	@echo "  3. $(YELLOW)make build$(RESET)   - Build binary"
	@echo "  4. $(YELLOW)make install$(RESET) - Install to Go bin"
	@echo ""
	@echo "$(BLUE)Example usage:$(RESET)"
	@echo "  ./dist/provider-explorer_*/provider-explorer fixtures/example-aws"