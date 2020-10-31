VERSION = $(shell git describe --tags --abbrev=0)

.PHONY: build-all
build-all:
	make build-linux-amd64
	make build-linux-386
	make build-linux-arm

.PHONY: build-linux-amd64
build-linux-amd64:
	env GOOS=linux GOARCH=amd64 go build -o builds/anonip-$(VERSION)-linux-amd64 github.com/open-dynaMIX/anonip-go

.PHONY: build-linux-386
build-linux-386:
	env GOOS=linux GOARCH=386 go build -o builds/anonip-$(VERSION)-linux-386 github.com/open-dynaMIX/anonip-go

.PHONY: build-linux-arm
build-linux-arm:
	env GOOS=linux GOARCH=arm go build -o builds/anonip-$(VERSION)-linux-arm github.com/open-dynaMIX/anonip-go
