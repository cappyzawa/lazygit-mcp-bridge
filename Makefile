.PHONY: build clean install test

BINARY_NAME=lazygit-mcp-bridge
BUILD_DIR=build
CMD_DIR=cmd/lazygit-mcp-bridge

# Version information
VERSION ?= dev
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Build flags
LDFLAGS = -s -w \
	-X main.version=$(VERSION) \
	-X main.commit=$(COMMIT) \
	-X main.date=$(DATE)

build:
	go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) ./$(CMD_DIR)

clean:
	rm -rf $(BUILD_DIR)

install:
	go install -ldflags "$(LDFLAGS)" ./$(CMD_DIR)

test:
	go test ./...

run-server:
	go run $(CMD_DIR)/main.go server

.DEFAULT_GOAL := build