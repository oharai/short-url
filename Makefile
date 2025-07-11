# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=shorturl-api
BINARY_UNIX=$(BINARY_NAME)_unix
COVERAGE_FILE=coverage.out

# Build flags
LDFLAGS=-ldflags="-s -w"

.PHONY: all build clean test coverage deps lint fmt vet help

all: test build ## Run tests and build

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: ## Build the binary file
	$(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME) -v ./cmd/api

build-linux: ## Build the binary file for Linux
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_UNIX) -v ./cmd/api

clean: ## Remove previous build and coverage files
	$(GOCLEAN)
	rm -rf bin/
	rm -f $(COVERAGE_FILE)
	rm -f coverage.html

test: ## Run tests
	$(GOTEST) -v -race ./...

test-coverage: ## Run tests with coverage
	$(GOTEST) -v -race -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./...
	$(GOCMD) tool cover -html=$(COVERAGE_FILE) -o coverage.html

coverage: test-coverage ## Generate test coverage report
	$(GOCMD) tool cover -func=$(COVERAGE_FILE)

coverage-html: test-coverage ## Generate HTML coverage report
	$(GOCMD) tool cover -html=$(COVERAGE_FILE)

integration-test: ## Run integration tests
	$(GOTEST) -v -tags=integration ./test/...

benchmark: ## Run benchmark tests
	$(GOTEST) -bench=. -benchmem ./...

deps: ## Download and verify dependencies
	$(GOMOD) download
	$(GOMOD) verify

deps-upgrade: ## Upgrade dependencies
	$(GOGET) -u ./...
	$(GOMOD) tidy

lint: ## Run golangci-lint
	golangci-lint run

fmt: ## Run go fmt on all files
	$(GOCMD) fmt ./...

vet: ## Run go vet
	$(GOCMD) vet ./...

check: fmt vet lint ## Run fmt, vet, and lint

security: ## Run gosec security scanner
	gosec ./...

vulncheck: ## Check for known vulnerabilities
	govulncheck ./...

run: build ## Build and run the application
	./bin/$(BINARY_NAME)

docker-build: ## Build docker image
	docker build -t $(BINARY_NAME) .

docker-run: docker-build ## Run docker container
	docker run -p 8080:8080 $(BINARY_NAME)

install-tools: ## Install development tools
	$(GOGET) github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GOGET) github.com/securecodewarrior/github-action-gosec@latest
	$(GOGET) golang.org/x/vuln/cmd/govulncheck@latest

ci: check test-coverage ## Run CI pipeline locally

pre-commit: fmt vet lint test ## Run pre-commit checks

# Create necessary directories
bin:
	mkdir -p bin

.DEFAULT_GOAL := help