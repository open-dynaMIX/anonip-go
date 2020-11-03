.PHONY: build
build:
	goreleaser --snapshot --skip-sign --skip-publish --rm-dist

.PHONY: test
test:
	@go test -cover -v

.PHONY: lint
lint:
	@golangci-lint run --exclude-use-default=false --enable=golint
