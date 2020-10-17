
.PHONY: dev-setup help fmt lint build test test-coverage
.DEFAULT_GOAL:=help

DIR := ${CURDIR}

setup: ## Setup the Development environment
	pre-commit install

fmt: ## Formats the go code using gofmt
	@gofmt -w -s .

lint: ## Lint code
	@docker run --rm -v $(DIR):/app -w /app golangci/golangci-lint:v1.30.0 golangci-lint run -v

build: ## Build the app
	@go build -o build/app .

test: ## Run package unit tests
	@go test -v -count=1 -short -coverprofile cover/cover.out -covermode=atomic  ./...

help: ## Displays help menu
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
