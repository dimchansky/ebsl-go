SHELL := bash
PACKAGE_NAME := github.com/dimchansky/ebsl-go
ARTIFACTS_DIR := $(if $(ARTIFACTS_DIR),$(ARTIFACTS_DIR),bin)

PKGS ?= $(shell go list ./...)
BENCH_FLAGS ?= -benchmem

export GO111MODULE := on

.PHONY: all
all: lint test

.PHONY: dependencies
dependencies:
	@echo "Installing dependencies..."
	go mod download
	@echo "Installing goimports..."
	go install golang.org/x/tools/cmd/goimports
	@echo "Installing golint..."
	go install golang.org/x/lint/golint
	@echo "Installing staticcheck..."
	go install honnef.co/go/tools/cmd/staticcheck

.PHONY: lint
lint:
	@echo "Checking formatting..."
	@gofiles=$$(go list -f {{.Dir}} $(PKGS) | grep -v mock) && [ -z "$$gofiles" ] || unformatted=$$(for d in $$gofiles; do goimports -l $$d/*.go; done) && [ -z "$$unformatted" ] || (echo >&2 "Go files must be formatted with goimports. Following files has problem:\n$$unformatted" && false)
	@echo "Checking vet..."
	@go vet ./...
	@echo "Checking staticcheck..."
	@staticcheck ./...
	@echo "Checking lint..."
	@golint ./...

.PHONY: test
test:
	go test -count=1 -tags=dev -timeout 60s -v ./...

.PHONY: bench
BENCH ?= .
bench:
	go test -bench=$(BENCH) -run="^$$" $(BENCH_FLAGS) ./...

.PHONY: cover
cover:
	mkdir -p ./${ARTIFACTS_DIR}/.cover
	go test -count=1 -coverprofile=./${ARTIFACTS_DIR}/.cover/cover.out -covermode=atomic -coverpkg=./...
	go tool cover -func=./${ARTIFACTS_DIR}/.cover/cover.out
	go tool cover -html=./${ARTIFACTS_DIR}/.cover/cover.out -o ./${ARTIFACTS_DIR}/cover.html

.PHONY: fmt
fmt:
	@echo "Formatting files..."
	@gofiles=$$(go list -f {{.Dir}} $(PKGS) | grep -v mock) && [ -z "$$gofiles" ] || for d in $$gofiles; do goimports -l -w $$d/*.go; done
