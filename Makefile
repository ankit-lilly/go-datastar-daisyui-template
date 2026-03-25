.PHONY: setup install build run dev clean templ fmt

install:
	@go run ./cmd/install

setup: install

templ:
	@go tool templ generate

fmt:
	@find . -type f -name '*.go' -not -path './vendor/*' -exec gofmt -w {} +

build: templ
	@go build -o bin/server ./cmd/server

run: build
	@./bin/server

dev:
	@if [ ! -f static/css/tailwindcss ]; then \
		echo "Run 'make install' first to download dependencies"; \
		exit 1; \
	fi
	@echo "Starting templ watcher, CSS watcher and server..."
	@go tool templ generate --watch &
	@cd static/css && ./tailwindcss -i input.css -o output.css --watch &
	@sleep 1 && go run ./cmd/server

css:
	@cd static/css && ./tailwindcss -i input.css -o output.css --minify

clean:
	@rm -rf bin/
	@rm -f static/css/output.css
	@rm -f internal/views/*_templ.go

clean-all: clean
	@rm -f static/css/tailwindcss
	@rm -f static/css/daisyui.mjs
	@rm -f static/css/daisyui-theme.mjs
	@rm -f static/css/input.css
	@rm -f static/js/datastar.js

deps:
	@go mod tidy

test:
	@go test -v ./...

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
