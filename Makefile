.DEFAULT_GOAL := all
.PHONY: help generate format test test-fast fuzz all

# How long to spend on each fuzz target. Override: make fuzz FUZZTIME=10m
FUZZTIME ?= 60s
# How long to spend shrinking each new input. Go's default, 60s, lets workers
# sit minimizing large inputs instead of fuzzing: 1s finds many times more
# coverage per run. Override: make fuzz FUZZMINIMIZETIME=10s
FUZZMINIMIZETIME ?= 1s

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

## fuzz: Fuzz each target the fuzz workflow runs, for FUZZTIME (default 60s)
fuzz:
	go test -run='^$$' -fuzz='^FuzzLoadFromData$$' -fuzztime=$(FUZZTIME) -fuzzminimizetime=$(FUZZMINIMIZETIME) ./openapi3
	go test -run='^$$' -fuzz='^FuzzLoadWithOrigins$$' -fuzztime=$(FUZZTIME) -fuzzminimizetime=$(FUZZMINIMIZETIME) ./openapi3
	go test -run='^$$' -fuzz='^FuzzYAMLMatchesJSON$$' -fuzztime=$(FUZZTIME) -fuzzminimizetime=$(FUZZMINIMIZETIME) ./openapi3
	go test -run='^$$' -fuzz='^FuzzSchemaVisitJSON$$' -fuzztime=$(FUZZTIME) -fuzzminimizetime=$(FUZZMINIMIZETIME) ./openapi3
	go test -run='^$$' -fuzz='^FuzzInternalizeRefs$$' -fuzztime=$(FUZZTIME) -fuzzminimizetime=$(FUZZMINIMIZETIME) ./openapi3
	go test -run='^$$' -fuzz='^FuzzInternalizeRefsKeepsMeaning$$' -fuzztime=$(FUZZTIME) -fuzzminimizetime=$(FUZZMINIMIZETIME) ./openapi3
	go test -run='^$$' -fuzz='^FuzzValidateOptions$$' -fuzztime=$(FUZZTIME) -fuzzminimizetime=$(FUZZMINIMIZETIME) ./openapi3
	go test -run='^$$' -fuzz='^FuzzInternalizeRefsCustomResolver$$' -fuzztime=$(FUZZTIME) -fuzzminimizetime=$(FUZZMINIMIZETIME) ./openapi3
	go test -run='^$$' -fuzz='^FuzzMarshalYAML$$' -fuzztime=$(FUZZTIME) -fuzzminimizetime=$(FUZZMINIMIZETIME) ./openapi3
	go test -run='^$$' -fuzz='^FuzzUnmarshalFromData$$' -fuzztime=$(FUZZTIME) -fuzzminimizetime=$(FUZZMINIMIZETIME) ./openapi2
	go test -run='^$$' -fuzz='^FuzzNew$$' -fuzztime=$(FUZZTIME) -fuzzminimizetime=$(FUZZMINIMIZETIME) ./internal/nodeindex
	go test -run='^$$' -fuzz='^FuzzToV3$$' -fuzztime=$(FUZZTIME) -fuzzminimizetime=$(FUZZMINIMIZETIME) ./openapi2conv
	go test -run='^$$' -fuzz='^FuzzValidateRequest$$' -fuzztime=$(FUZZTIME) -fuzzminimizetime=$(FUZZMINIMIZETIME) ./openapi3filter
	go test -run='^$$' -fuzz='^FuzzValidateResponse$$' -fuzztime=$(FUZZTIME) -fuzzminimizetime=$(FUZZMINIMIZETIME) ./openapi3filter
	go test -run='^$$' -fuzz='^FuzzBodyDecoder$$' -fuzztime=$(FUZZTIME) -fuzzminimizetime=$(FUZZMINIMIZETIME) ./openapi3filter
	go test -run='^$$' -fuzz='^FuzzRoutersAgree$$' -fuzztime=$(FUZZTIME) -fuzzminimizetime=$(FUZZMINIMIZETIME) ./routers
	go test -run='^$$' -fuzz='^FuzzServerMatchRawURL$$' -fuzztime=$(FUZZTIME) -fuzzminimizetime=$(FUZZMINIMIZETIME) ./openapi3
	go test -run='^$$' -fuzz='^FuzzPathsFind$$' -fuzztime=$(FUZZTIME) -fuzzminimizetime=$(FUZZMINIMIZETIME) ./openapi3
	go test -run='^$$' -fuzz='^FuzzFromV3$$' -fuzztime=$(FUZZTIME) -fuzzminimizetime=$(FUZZMINIMIZETIME) ./openapi2conv
	go test -run='^$$' -fuzz='^FuzzPathPattern$$' -fuzztime=$(FUZZTIME) -fuzzminimizetime=$(FUZZMINIMIZETIME) ./routers/legacy/pathpattern
	go test -run='^$$' -fuzz='^FuzzValidateParameter$$' -fuzztime=$(FUZZTIME) -fuzzminimizetime=$(FUZZMINIMIZETIME) ./openapi3filter

## help: list targets
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  make /'
