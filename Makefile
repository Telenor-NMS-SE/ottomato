.PHONY: setup githooks install test format

install:
	@echo "installing modules"
	@go mod download

githooks:
	@echo "adding git hooks"
	@git config core.hooksPath .githooks

setup: install githooks

test:
	go test -v ./...