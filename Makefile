.DEFAULT_GOAL := all
.PHONY: help generate format test test-fast all

## all: Run generation, formatting, and tests in sequence
all: generate format test

## generate: Regenerate Go code, map helpers, and API documentation
generate:
	go generate ./...
	./maps.sh
	./docs.sh

## format: Apply the repository formatters and import cleanup
format:
	go fmt ./...

## test: Run the full Go package test suite
test:
	go test ./...

## test-fast: Run tests skipping the slow ones
test-fast:
	go test -short ./...

## help: list targets
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  make /'
