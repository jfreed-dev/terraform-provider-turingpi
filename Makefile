.PHONY: build test lint clean install fmt vet

BINARY_NAME=terraform-provider-turingpi
VERSION?=1.0.0
GOOS?=$(shell go env GOOS)
GOARCH?=$(shell go env GOARCH)
PLUGIN_DIR=~/.terraform.d/plugins/local/turingpi/turingpi/$(VERSION)/$(GOOS)_$(GOARCH)

default: build

build:
	go build -o $(BINARY_NAME)

test:
	go test -v ./...

test-race:
	go test -v -race ./...

test-cover:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

lint:
	golangci-lint run

fmt:
	go fmt ./...

vet:
	go vet ./...

clean:
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html

install: build
	mkdir -p $(PLUGIN_DIR)
	cp $(BINARY_NAME) $(PLUGIN_DIR)/

uninstall:
	rm -rf ~/.terraform.d/plugins/local/turingpi

tidy:
	go mod tidy

deps:
	go mod download

all: fmt vet lint test build
