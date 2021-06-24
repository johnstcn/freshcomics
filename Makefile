CONTAINERTOOL := "docker"
GOBIN ?= "/usr/local/go/bin/go"
GOLANGCI_LINT ?= "/usr/local/go/bin/golangci-lint"
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
docker:
	$(CONTAINERTOOL) build . -f Dockerfile.$(OUTPUT_BINARY_BE) -t $(OUTPUT_BINARY_BE):latest
	$(CONTAINERTOOL) build . -f Dockerfile.$(OUTPUT_BINARY_FE) -t $(OUTPUT_BINARY_FE):latest

## clean: clean built binary
clean: go-clean

## lint: run linters
lint: golangci-lint

## test: run unit tests
test: go-test

## test-race: run unit tests with race detector enabled
test-race: go-test-race

## build: build binary
build: go-build

go-clean:
	$(GOBIN) clean
	rm -f $(OUTPUT_BINARY_FE)
	rm -f $(OUTPUT_BINARY_BE)

golangci-lint:
	$(GOLANGCI_LINT) run ./...

go-test:
	$(GOBIN) test -v ./...

go-test-race:
	$(GOBIN) test -v -race ./...

go-build:
	$(GOBIN) build -o $(OUTPUT_BINARY_FE) cmd/freshcomics/main.go
	$(GOBIN) build -o $(OUTPUT_BINARY_BE) cmd/crawld/main.go

