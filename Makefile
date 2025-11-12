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

coverage:
	go test -coverprofile=.coverage -coverpkg=./... ./...
	go tool cover -html .coverage -o cover.html
	rm .coverage