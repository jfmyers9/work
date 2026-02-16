.PHONY: build clean test release lint fmt help

.DEFAULT_GOAL := help

## build: Build binaries for all platforms
build:
	@./scripts/build.sh

## clean: Remove build artifacts
clean:
	@rm -rf ./dist
	@echo "✓ Build artifacts removed"

## test: Run all tests
test:
	@go test -v -race -coverprofile=coverage.out ./...
	@echo ""
	@echo "Test coverage:"
	@go tool cover -func=coverage.out | tail -1

## install: Install work locally
install: build
	@cp ./dist/work /usr/local/bin/work
	@echo "✓ work installed to /usr/local/bin/work"

## release: Create a release (requires VERSION)
release:
	@if [ -z "$(VERSION)" ]; then \
		echo "ERROR: VERSION not set. Usage: make release VERSION=v1.0.0"; \
		exit 1; \
	fi
	@echo "Creating release $(VERSION)..."
	@git tag -a $(VERSION) -m "Release $(VERSION)"
	@git push origin $(VERSION)
	@./scripts/build.sh
	@echo ""
	@echo "✓ Release $(VERSION) tagged and built"
	@echo ""
	@echo "Next steps:"
	@echo "  1. Create GitHub release at: https://github.com/jfmyers9/work/releases/new?tag=$(VERSION)"
	@echo "  2. Upload artifacts from ./dist/"

## lint: Run linters
lint:
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed. Install from: https://golangci-lint.run/welcome/install/" && exit 1)
	@golangci-lint run

## fmt: Format code
fmt:
	@go fmt ./...
	@echo "✓ Code formatted"

## help: Show this help message
help:
	@echo "work - filesystem-based issue tracker"
	@echo ""
	@echo "Usage:"
	@echo "  make <target>"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^## //p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/  /'
