VERSION = $(shell git describe --tags --abbrev=0)
export CGO_ENABLED=0

.PHONY: build
build:
	goreleaser --snapshot --skip-sign --skip-publish --rm-dist

.PHONY: test
test:
	@go test -cover -v

.PHONY: lint
lint:
	@golangci-lint run --exclude-use-default=false --enable=golint
