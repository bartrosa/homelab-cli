SHELL := /bin/bash
.DEFAULT_GOAL := help

MODULE_PATH := github.com/bartrosa/homelab-cli
BIN_DIR := bin
BINARY := $(BIN_DIR)/lab

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
GOVERSION ?= $(shell go env GOVERSION)

LDFLAGS := -s -w \
	-X '$(MODULE_PATH)/internal/buildinfo.Version=$(VERSION)' \
	-X '$(MODULE_PATH)/internal/buildinfo.Commit=$(COMMIT)' \
	-X '$(MODULE_PATH)/internal/buildinfo.Date=$(DATE)' \
	-X '$(MODULE_PATH)/internal/buildinfo.GoVersion=$(GOVERSION)'

.PHONY: help build install test lint fmt vet tidy clean ci

help: ## List targets
	@grep -E '^[a-zA-Z0-9_-]+:.*?##' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-18s\033[0m %s\n", $$1, $$2}'

build: ## Build bin/lab
	@mkdir -p $(BIN_DIR)
	go build -trimpath -ldflags "$(LDFLAGS)" -o $(BINARY) ./cmd/lab

install: ## go install with ldflags
	go install -trimpath -ldflags "$(LDFLAGS)" ./cmd/lab

test: ## Run tests
	go test ./... -race -cover

lint: ## Run golangci-lint
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.6 run

fmt: ## gofmt + golangci formatters (gofumpt/goimports)
	gofmt -s -w .
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.6 fmt

vet: ## go vet
	go vet ./...

tidy: ## go mod tidy
	go mod tidy

clean: ## Remove build artifacts
	rm -rf $(BIN_DIR) dist

ci: fmt vet lint test build ## Full local CI pipeline
