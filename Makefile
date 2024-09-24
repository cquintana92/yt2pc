ARTIFACT_NAME:=yt2pc

SHELL:=/bin/sh
MAKEFILE_PATH := $(abspath $(lastword $(MAKEFILE_LIST)))
PROJECT_ROOT := $(dir $(MAKEFILE_PATH))
BIN_DIRECTORY := ${PROJECT_ROOT}/bin

.PHONY: default
default: help

.PHONY: deps
deps: ## Setup dependencies
	@go get ./...

.PHONY: build
build: deps ## Build
	@go build

.PHONY: fmt
fmt: ## Apply linting and formatting
	@go fmt ./...

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
