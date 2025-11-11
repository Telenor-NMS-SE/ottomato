.PHONY: setup githooks install test format

install:
	@go mod download

githoooks:
	@for hook in .githooks/*; do \
		name=$$(basename $$hook); \
		rm -f .git/hooks/$$name; \
		ln -s ../../$$name .git/hooks/$$name; \
	done

setup: install githooks

test:
	go test -v ./...