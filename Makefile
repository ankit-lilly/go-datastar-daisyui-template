.PHONY: setup install build run dev clean templ fmt

# Install downloads all dependencies using Go install script
install:
	@go run ./cmd/install

# Setup is an alias for install (backwards compatibility)
setup: install

# Generate templ files (using go tool directive)
templ:
	@go tool templ generate

# Format Go source files
fmt:
	@find . -type f -name '*.go' -not -path './vendor/*' -exec gofmt -w {} +

# Build the Go binary (generates templ first)
build: templ
	@go build -o bin/server ./cmd/server

# Run the server
run: build
	@./bin/server

# Development mode: watch templ, CSS and run server
dev:
	@if [ ! -f static/css/tailwindcss ]; then \
		echo "Run 'make install' first to download dependencies"; \
		exit 1; \
	fi
	@echo "Starting templ watcher, CSS watcher and server..."
	@go tool templ generate --watch &
	@cd static/css && ./tailwindcss -i input.css -o output.css --watch &
	@sleep 1 && go run ./cmd/server

# Rebuild CSS only
css:
	@cd static/css && ./tailwindcss -i input.css -o output.css --minify

# Clean build artifacts
clean:
	@rm -rf bin/
	@rm -f static/css/output.css
	@rm -f internal/views/*_templ.go

# Clean everything including downloaded files
clean-all: clean
	@rm -f static/css/tailwindcss
	@rm -f static/css/daisyui.mjs
	@rm -f static/css/daisyui-theme.mjs
	@rm -f static/css/input.css
	@rm -f static/js/datastar.js

# Download Go dependencies
deps:
	@go mod tidy

# Run tests
test:
	@go test -v ./...

# Help
help:
	@echo "Available targets:"
	@echo "  install    - Download Tailwind CSS, DaisyUI, Datastar and setup templ"
	@echo "  setup      - Alias for install"
	@echo "  templ      - Generate Go code from templ files"
	@echo "  fmt        - Format Go source files"
	@echo "  build      - Generate templ and build the Go binary"
	@echo "  run        - Build and run the server"
	@echo "  dev        - Run in development mode with templ and CSS watchers"
	@echo "  css        - Rebuild CSS only"
	@echo "  deps       - Download Go dependencies"
	@echo "  test       - Run tests"
	@echo "  clean      - Remove build artifacts and generated templ files"
	@echo "  clean-all  - Remove all generated files"
	@echo "  help       - Show this help"
