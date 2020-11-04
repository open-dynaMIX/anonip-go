.PHONY: build
build:
	goreleaser --snapshot --skip-sign --skip-publish --rm-dist

.PHONY: test
test:
	@go test -cover -v

.PHONY: test-no-cov
test-no-cov:
	go test -v -cover -ic

.PHONY: lint
lint:
	@golangci-lint run --exclude-use-default=false --enable=golint

.PHONY: html-coverage
coverage:
	go test -coverprofile=coverage.out && go tool cover -html=coverage.out
