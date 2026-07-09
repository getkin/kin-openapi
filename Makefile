.DEFAULT_GOAL := prepare

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
	go fmt ./...

test:
	go test ./...

prepare:
	$(MAKE) generate
	$(MAKE) format
	$(MAKE) test
