#!/usr/bin/env bash

# Setup script for Go + Templ + Tailwind CSS + DaisyUI (standalone, no npm required)
# Downloads latest versions from GitHub releases
#
# Usage: ./scripts/setup.sh [static_dir]
#
# This script:
# - Downloads latest Tailwind CSS standalone binary
# - Downloads latest DaisyUI plugin files
# - Downloads latest Datastar JS
# - Installs templ CLI (if not present)
# - Creates input.css with templ source paths
# - Compiles CSS with minification

set -euo pipefail

# Configuration
STATIC_DIR="${1:-static}"
CSS_DIR="$STATIC_DIR/css"
JS_DIR="$STATIC_DIR/js"

TAILWIND_BASE_URL="https://github.com/tailwindlabs/tailwindcss/releases/latest/download"
DAISYUI_BASE_URL="https://github.com/saadeghi/daisyui/releases/latest/download"
DATASTAR_URL="https://cdn.jsdelivr.net/gh/starfederation/datastar@latest/bundles/datastar.js"

INPUT_CSS_CONTENT='@import "tailwindcss";

@source "../../internal/views/**/*.templ";
@source not "./tailwindcss";
@source not "./daisyui{,*}.mjs";

@plugin "./daisyui.mjs";'

# Error handler
trap 'echo "  ❌ Installation failed" >&2; exit 1' ERR

# Detect OS
get_os() {
    case "$(uname -s)" in
        Linux*) echo "linux";;
        Darwin*) echo "macos";;
        *) echo "unknown";;
    esac
}

# Detect architecture
get_arch() {
    case "$(uname -m)" in
        x86_64|amd64) echo "x64";;
        aarch64|arm64) echo "arm64";;
        *) echo "unknown";;
    esac
}

# Check for musl libc on Linux
get_musl_suffix() {
    [ "$(uname -s)" = "Linux" ] && ldd --version 2>&1 | grep -q musl && echo "-musl" || echo ""
}

# Format OS name for display
format_os() {
    case "$1" in
        linux) echo "Linux";;
        macos) echo "macOS";;
        *) echo "$1";;
    esac
}

# Build Tailwind filename for this platform
build_tailwind_filename() {
    echo "tailwindcss-$1-$2$3"
}

# Install templ CLI if not present
install_templ() {
    echo "  📦 Setting up templ (via go tool directive)..."
    go mod tidy
    echo "  ✅ templ available via 'go tool templ'"
}

# Main installation
main() {
    local os=$(get_os)
    local arch=$(get_arch)

    [ "$os" = "unknown" ] || [ "$arch" = "unknown" ] && echo "❌ Unsupported system" >&2 && exit 1

    echo "🚀 Setting up Go + Templ + Tailwind CSS + DaisyUI for $(format_os "$os") $arch"
    echo ""

    # Create directories
    mkdir -p "$CSS_DIR"
    mkdir -p "$JS_DIR"

    # Install templ
    install_templ

    # Download Tailwind CSS binary (latest)
    echo "  📦 Downloading Tailwind CSS (latest)..."
    local filename=$(build_tailwind_filename "$os" "$arch" "$(get_musl_suffix)")
    curl -fsSLo "$CSS_DIR/tailwindcss" "$TAILWIND_BASE_URL/$filename"
    chmod +x "$CSS_DIR/tailwindcss"
    echo "  ✅ Tailwind CSS downloaded"

    # Download DaisyUI (latest)
    echo "  📦 Downloading DaisyUI (latest)..."
    curl -fsSLo "$CSS_DIR/daisyui.mjs" "$DAISYUI_BASE_URL/daisyui.mjs"
    curl -fsSLo "$CSS_DIR/daisyui-theme.mjs" "$DAISYUI_BASE_URL/daisyui-theme.mjs"
    echo "  ✅ DaisyUI downloaded"

    # Download Datastar (latest)
    echo "  📦 Downloading Datastar (latest)..."
    curl -fsSLo "$JS_DIR/datastar.js" "$DATASTAR_URL"
    echo "  ✅ Datastar downloaded"

    # Create input.css
    echo "$INPUT_CSS_CONTENT" > "$CSS_DIR/input.css"
    echo "  ✅ Created $CSS_DIR/input.css"

    # Generate templ files
    echo "  🔨 Generating templ files..."
    go tool templ generate || echo "  ⚠️  templ generate skipped (this is normal for first setup)"

    # Build CSS
    echo "  🔨 Building CSS..."
    cd "$CSS_DIR"
    ./tailwindcss -i input.css -o output.css --minify
    cd - > /dev/null
    echo "  ✅ CSS built"

    echo ""
    echo "✅ Setup complete!"
    echo ""
    echo "Files created:"
    echo "  - $CSS_DIR/tailwindcss (binary)"
    echo "  - $CSS_DIR/daisyui.mjs"
    echo "  - $CSS_DIR/daisyui-theme.mjs"
    echo "  - $CSS_DIR/input.css"
    echo "  - $CSS_DIR/output.css"
    echo "  - $JS_DIR/datastar.js"
    echo ""
    echo "Next steps:"
    echo "  make build   - Build the server"
    echo "  make run     - Build and run the server"
    echo "  make dev     - Run in development mode with watchers"
}

main
