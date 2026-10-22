.DEFAULT_GOAL := all
.PHONY: help generate format test test-fast fuzz all

# How long to spend on each fuzz target. Override: make fuzz FUZZTIME=10m
FUZZTIME ?= 60s

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

## fuzz: Fuzz each target for FUZZTIME (default 60s)
fuzz:
	go test -run='^$$' -fuzz='^FuzzLoadFromData$$' -fuzztime=$(FUZZTIME) ./openapi3
	go test -run='^$$' -fuzz='^FuzzSchemaVisitJSON$$' -fuzztime=$(FUZZTIME) ./openapi3
	go test -run='^$$' -fuzz='^FuzzToV3$$' -fuzztime=$(FUZZTIME) ./openapi2conv
	go test -run='^$$' -fuzz='^FuzzValidateRequest$$' -fuzztime=$(FUZZTIME) ./openapi3filter

## help: list targets
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  make /'
