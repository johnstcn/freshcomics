CONTAINERTOOL := $(shell which docker)
GOBIN := $(shell which go)
GOLANGCI_LINT := $(shell which golangci-lint)
OUTPUT_BINARY_FE ?= "freshcomics"
OUTPUT_BINARY_BE ?= "crawld"

all: help

.PHONY: help
help: Makefile
	@echo
	@echo " Available targets:"
	@sed -n 's/^##//p' $< | column -t -s ':' | sed -e 's/^/ /'
	@echo

## docker: build docker image
.PHONY: docker
docker:
	$(CONTAINERTOOL) build . -f Dockerfile.$(OUTPUT_BINARY_BE) -t $(OUTPUT_BINARY_BE):latest
	$(CONTAINERTOOL) build . -f Dockerfile.$(OUTPUT_BINARY_FE) -t $(OUTPUT_BINARY_FE):latest

## clean: clean built binary
.PHONY: clean
clean: go-clean

## lint: run linters
.PHONY: lint
lint: golangci-lint

## test: run unit tests
.PHONY: test
test: go-test

## build: build binary
.PHONY: go-build
build: go-build

.PHONY: go-clean
go-clean:
	$(GOBIN) clean
	rm -rf ./build

.PHONY: golangci-lint
golangci-lint:
	$(GOLANGCI_LINT) run ./...

.PHONY: go-test
go-test:
	$(GOBIN) test -v -race ./...

.PHONY: go-build
go-build:
	mkdir -p ./build
	$(GOBIN) build -o build/$(OUTPUT_BINARY_FE) cmd/freshcomics/main.go
	$(GOBIN) build -o build/$(OUTPUT_BINARY_BE) cmd/crawld/main.go
