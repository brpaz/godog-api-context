
.PHONY: dev-setup help fmt lint build test test-coverage
.DEFAULT: help

DIR := ${CURDIR}

dev-setup: ## Setup the Development envrionment
	pre-commit install

fmt: ## Formats the go code using gofmt
	@gofmt -w -s .

lint: ## Lint code
	@docker run --rm -v $(DIR):/app -w /app golangci/golangci-lint:v1.30.0 golangci-lint run -v

build: ## Build the app
	@go build -o build/app .

test: ## Run package unit tests
	@go test -v -race -count=1 -short  ./...

test-cover: ## Run tests with coverage
	@mkdir -p cover
	@go test -count=1 -short -coverprofile cover/cover.out -covermode=atomic  ./...
	@go tool cover -html=cover/cover.out

help: ## Displays help menu
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
