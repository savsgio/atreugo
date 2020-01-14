BIN_DIR := .tools/bin
GOLANGCI_LINT_VERSION := 1.22.2
GOLANGCI_LINT := $(BIN_DIR)/golangci-lint

all: build test lint

tidy:
	go mod tidy -v

build:
	go build ./...

test:
	go test -cover -race -v ./...

lint: $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) run --fast --enable-all

$(GOLANGCI_LINT):
	curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(BIN_DIR) v$(GOLANGCI_LINT_VERSION)
