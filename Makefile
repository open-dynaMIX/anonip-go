.DEFAULT_GOAL := help

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort -k 1,1 | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: build
build: ## Build project snapshot with goreleaser
	goreleaser --snapshot --skip-sign --skip-publish --rm-dist

.PHONY: test
test: ## Test the project
	go test -cover -v

.PHONY: test-no-cov
test-no-cov: ## Test the project, do not enforce 100% coverage
	go test -v -cover -ic

.PHONY: lint
lint: ## Lint the project
	golangci-lint run --exclude-use-default=false --enable=golint

.PHONY: html-coverage
coverage: ## Create html coverage and open it in browser
	go test -coverprofile=coverage.out && go tool cover -html=coverage.out
