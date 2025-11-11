.PHONY: setup githooks install test format

install:
	@go mod download

githooks:
	@git config core.hooksPath .githooks

setup: install githooks

test:
	go test -v ./...