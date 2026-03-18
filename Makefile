BIN_DIR := ./bin
BIN_NAME := structify

.PHONY: build test lint install clean

build:
	go build -o $(BIN_DIR)/$(BIN_NAME) ./cmd/structify

test:
	go test ./...

lint:
	@if command -v golangci-lint >/dev/null 2>&1; then golangci-lint run; else echo "golangci-lint not installed, skipping lint"; fi

install:
	go install ./...

clean:
	rm -rf $(BIN_DIR)

