# For linux, relace darwin with linux in 'build'

EXECUTABLE=go-spss
LINUX=lib$(EXECUTABLE)_linux.so
DARWIN=lib$(EXECUTABLE)_darwin.so
VERSION=$(shell git describe --tags --always --long --dirty)

.PHONY: all test clean

all: test build ## Build and run tests

test: ## Run unit tests
	go test

build: linux  darwin ## Build binaries
	@echo version: $(VERSION)

linux: $(LINUX) ## Build for Linux

darwin: $(DARWIN) ## Build for Darwin (macOS)

$(LINUX):
	env GOOS=linux GOARCH=amd64 go build -i -v -o $(LINUX) -ldflags="-s -w -X main.version=$(VERSION) -lreadstat"

$(DARWIN):
	env GOOS=darwin GOARCH=amd64 go build -i -v -o $(DARWIN) -ldflags="-s -w -X main.version=$(VERSION) -lreadstat"

clean: ## Remove previous build
	rm -f $(LINUX) $(DARWIN)

help: ## Display available commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
