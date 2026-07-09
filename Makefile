.DEFAULT_GOAL := help

.PHONY: help generate format test prepare

help:
	@printf '%s\n' \
		'Usage: make <target>' \
		'' \
		'  generate  Regenerate Go code, map helpers, and API documentation.' \
		'  format    Apply the repository formatters and import cleanup.' \
		'  test      Run the full Go package test suite.' \
		'  prepare   Run generation, formatting, and tests in sequence.'

generate:
	go generate ./...
	./maps.sh
	./docs.sh

format:
	go run golang.org/x/tools/go/analysis/passes/modernize/cmd/modernize@latest -test -diff -fix -omitzero=false ./...
	go fmt ./...
	go install github.com/incu6us/goimports-reviser/v3@latest
	find . -type f -iname '*.go' ! -iname '*.pb.go' -exec goimports-reviser {} \;

test:
	go test ./...

prepare:
	$(MAKE) generate
	$(MAKE) format
	$(MAKE) test
