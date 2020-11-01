VERSION = $(shell git describe --tags --abbrev=0)
export CGO_ENABLED=0

.PHONY: build
build:
	make build-linux-amd64

.PHONY: build-all
build-all:
	make build-linux-amd64
	make build-linux-386
	make build-linux-arm
	make build-netbsd-386
	make build-netbsd-amd64
	make build-netbsd-arm
	make build-freebsd-386
	make build-freebsd-amd64
	make build-freebsd-arm
	make build-darwin-386
	make build-darwin-amd64
	make build-openbsd-386
	make build-openbsd-amd64
	make build-windows-386
	make build-windows-amd64

.PHONY: build-linux-amd64
build-linux-amd64:
	env GOOS=linux GOARCH=amd64 go build -o builds/anonip-$(VERSION)-linux-amd64 github.com/open-dynaMIX/anonip-go

.PHONY: build-linux-386
build-linux-386:
	env GOOS=linux GOARCH=386 go build -o builds/anonip-$(VERSION)-linux-386 github.com/open-dynaMIX/anonip-go

.PHONY: build-linux-arm
build-linux-arm:
	env GOOS=linux GOARCH=arm go build -o builds/anonip-$(VERSION)-linux-arm github.com/open-dynaMIX/anonip-go

.PHONY: build-netbsd-386
build-netbsd-386:
	env GOOS=linux GOARCH=arm go build -o builds/anonip-$(VERSION)-netbsd-386 github.com/open-dynaMIX/anonip-go

.PHONY: build-netbsd-amd64
build-netbsd-amd64:
	env GOOS=linux GOARCH=arm go build -o builds/anonip-$(VERSION)-netbsd-amd64 github.com/open-dynaMIX/anonip-go

.PHONY: build-netbsd-arm
build-netbsd-arm:
	env GOOS=linux GOARCH=arm go build -o builds/anonip-$(VERSION)-netbsd-arm github.com/open-dynaMIX/anonip-go

.PHONY: build-freebsd-386
build-freebsd-386:
	env GOOS=linux GOARCH=arm go build -o builds/anonip-$(VERSION)-freebsd-386 github.com/open-dynaMIX/anonip-go

.PHONY: build-freebsd-amd64
build-freebsd-amd64:
	env GOOS=linux GOARCH=arm go build -o builds/anonip-$(VERSION)-freebsd-amd64 github.com/open-dynaMIX/anonip-go

.PHONY: build-freebsd-arm
build-freebsd-arm:
	env GOOS=linux GOARCH=arm go build -o builds/anonip-$(VERSION)-freebsd-arm github.com/open-dynaMIX/anonip-go

.PHONY: build-darwin-386
build-darwin-386:
	env GOOS=linux GOARCH=arm go build -o builds/anonip-$(VERSION)-darwin-386 github.com/open-dynaMIX/anonip-go

.PHONY: build-darwin-amd64
build-darwin-amd64:
	env GOOS=linux GOARCH=arm go build -o builds/anonip-$(VERSION)-darwin-amd64 github.com/open-dynaMIX/anonip-go

.PHONY: build-openbsd-386
build-openbsd-386:
	env GOOS=linux GOARCH=arm go build -o builds/anonip-$(VERSION)-openbsd-386 github.com/open-dynaMIX/anonip-go

.PHONY: build-openbsd-amd64
build-openbsd-amd64:
	env GOOS=linux GOARCH=arm go build -o builds/anonip-$(VERSION)-openbsd-amd64 github.com/open-dynaMIX/anonip-go

.PHONY: build-windows-386
build-windows-386:
	env GOOS=linux GOARCH=arm go build -o builds/anonip-$(VERSION)-windows-386 github.com/open-dynaMIX/anonip-go

.PHONY: build-windows-amd64
build-windows-amd64:
	env GOOS=linux GOARCH=arm go build -o builds/anonip-$(VERSION)-windows-amd64 github.com/open-dynaMIX/anonip-go

.PHONY: test
test:
	@go test -cover -v

.PHONY: lint
lint:
	@golangci-lint run --exclude-use-default=false --enable=golint
