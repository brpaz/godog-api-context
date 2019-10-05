
.PHONY: setup help dep format lint vet build test test-coverage
.DEFAULT: help

setup: ## Setup the Development envrionment
	pre-commit install

dep: ## Get build dependencies
	go mod download

fmt: ## Formats the go code using gofmt
	@gofmt -w -s .

lint: ## Lint code
	@revive -config revive.toml -formatter friendly ./...

vet: ## Run go vet
	@go vet ./...

test: ## Run package unit testsS
	@go test -v -race -short  ./...

test-coverage: ## Run tests with coverage
	@go test -short -coverprofile cover.out -covermode=atomic  ./...

help: ## Displays help menu
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
