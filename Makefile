# Makefile for easyGit

# Binary name
BINARY_NAME=easyGit
# Main package path
MAIN_PACKAGE=./cmd/easygit
# Versioning
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev-$(shell date +%Y%m%d)")
# Go build flags
LDFLAGS=-ldflags="-s -w -X github.com/KevinYouu/easyGit/internal/version.Version=$(VERSION)"

# Extract arguments for dev and run targets
ifeq ($(firstword $(MAKECMDGOALS)),$(filter $(firstword $(MAKECMDGOALS)),dev run))
  RUN_ARGS := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
  $(eval $(RUN_ARGS):;@:)
endif

.PHONY: all build test vet fmt clean install help

# Default target
all: vet test build ## Run vet, test, and build (default)

dev: ## Run the application with go run (usage: make dev <args>)
	go run $(LDFLAGS) $(MAIN_PACKAGE) $(RUN_ARGS)

run: build ## Build and run the binary (usage: make run <args>)
	./$(BINARY_NAME) $(RUN_ARGS)

build: ## Build the binary
	@echo "Building $(BINARY_NAME) version: $(VERSION)"
	go build $(LDFLAGS) -o $(BINARY_NAME) $(MAIN_PACKAGE)
	@echo "Build completed: $(BINARY_NAME)"

test: ## Run tests
	@echo "Running tests..."
	go test ./...

vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...

fmt: ## Run go fmt
	@echo "Running go fmt..."
	go fmt ./...

clean: ## Remove build artifacts
	@echo "Cleaning up..."
	rm -f $(BINARY_NAME)
	@echo "Done."

install: build ## Install the binary to $(GOPATH)/bin
	@echo "Installing $(BINARY_NAME) to $(shell go env GOPATH)/bin"
	cp $(BINARY_NAME) $(shell go env GOPATH)/bin/$(BINARY_NAME)
	@echo "Installation completed."

help: ## Show this help message
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'
