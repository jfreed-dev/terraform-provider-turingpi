.PHONY: build test lint clean install fmt vet release release-prep

BINARY_NAME=terraform-provider-turingpi
VERSION?=1.0.0
# Extract current version from README (more reliable than git tags)
CURRENT_VERSION=$(shell grep -oP 'version = "\K[0-9]+\.[0-9]+\.[0-9]+' README.md | head -1)
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
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@latest run

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

# Release preparation - updates version in all docs/examples
# Usage: make release-prep VERSION=1.0.9
release-prep:
	@if [ "$(VERSION)" = "1.0.0" ]; then \
		echo "ERROR: Please specify VERSION, e.g., make release-prep VERSION=1.0.9"; \
		exit 1; \
	fi
	@echo "Updating version references from $(CURRENT_VERSION) to $(VERSION)..."
	@# Update README.md
	@sed -i 's/version = "$(CURRENT_VERSION)"/version = "$(VERSION)"/g' README.md
	@# Update docs
	@find docs -name "*.md" -exec sed -i 's/version = "$(CURRENT_VERSION)"/version = "$(VERSION)"/g' {} \;
	@# Update examples
	@find examples -name "*.tf" -exec sed -i 's/version = "$(CURRENT_VERSION)"/version = "$(VERSION)"/g' {} \;
	@echo "Version references updated to $(VERSION)"

# Full release - updates docs, commits, tags, and pushes
# Usage: make release VERSION=1.0.9
release: release-prep test
	@echo "Committing version updates..."
	@git add README.md docs/ examples/
	@git commit -S -m "Update documentation to v$(VERSION)" || true
	@echo "Creating signed tag v$(VERSION)..."
	@git tag -s -a v$(VERSION) -m "v$(VERSION)"
	@echo "Pushing to origin..."
	@git push origin main
	@git push origin v$(VERSION)
	@echo ""
	@echo "✓ Release v$(VERSION) complete!"
	@echo "  - GitHub Actions will build and publish the release"
	@echo "  - Don't forget to update CHANGELOG.md after the release"
