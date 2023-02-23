SHELL := /bin/bash
PLATFORM := $(shell uname -s | tr '[:upper:]' '[:lower:]')
ARCH := $(shell uname -m | tr '[:upper:]' '[:lower:]')
ifeq ($(ARCH),x86_64)
	ARCH := amd64
endif

.PHONY: all
all: mod

.PHONY: mod
mod:
	@go mod tidy -compat=1.20
	@go mod verify
	@go mod vendor

.PHONY: update-dependencies
update-dependencies:
	@go get ./...
	@$(MAKE) mod

.PHONY: fmt
fmt:
	@go fmt ./...

.PHONY: vet
vet: fmt
	@go vet ./...

.PHONY: lint
lint: install-golangci-lint vet
	@clear
	@golangci-lint run .

.PHONY: install-golangci-lint
install-golangci-lint:
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.51.2
