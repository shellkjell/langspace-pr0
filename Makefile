.PHONY: all build test lint clean coverage benchmark docs local-ci setup-local-ci verify

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=langspace

# Linting
GOLINT=golangci-lint

all: lint test build

build:
	$(GOBUILD) -v ./...

test:
	$(GOTEST) -v -race ./...

lint:
	$(GOLINT) run

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f coverage.out

coverage:
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out

benchmark:
	$(GOTEST) -bench=. -benchmem ./...

deps:
	$(GOMOD) download
	$(GOMOD) tidy

docs:
	godoc -http=:6060

# Security scan
security:
	gosec ./...

# Generate mocks
mocks:
	mockgen -source=pkg/ast/entity.go -destination=pkg/ast/mock_entity.go -package=ast

build-main:
	$(GOBUILD) -o $(BINARY_NAME) cmd/langspace/main.go

# Cross compilation
build-all:
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_NAME)-linux-amd64
	GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BINARY_NAME)-windows-amd64.exe
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(BINARY_NAME)-darwin-amd64

# Docker
docker-build:
	docker build -t $(BINARY_NAME) .

# Development helpers
run:
	$(GOBUILD) -o $(BINARY_NAME)
	./$(BINARY_NAME)

watch:
	reflex -r '\.go$$' -s -- sh -c '$(GOTEST) ./...'

# Local CI support
verify: lint test

local-ci:
	act -j test -j build

setup-local-ci:
	git config core.hooksPath .githooks
	chmod +x .githooks/pre-push
	@echo "Local CI setup complete. Git hooks configured to use .githooks directory."
