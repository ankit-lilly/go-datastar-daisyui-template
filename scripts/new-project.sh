#!/usr/bin/env bash

# Create a new project from this template
# Usage: ./scripts/new-project.sh <project-path> <module-name>
# Example: ./scripts/new-project.sh ~/projects/my-app github.com/myuser/my-app

set -euo pipefail

TEMPLATE_DIR="$(cd "$(dirname "$0")/.." && pwd)"
PROJECT_PATH="${1:-}"
MODULE_NAME="${2:-}"

if [ -z "$PROJECT_PATH" ] || [ -z "$MODULE_NAME" ]; then
    echo "Usage: $0 <project-path> <module-name>"
    echo ""
    echo "Example:"
    echo "  $0 ~/projects/my-app github.com/myuser/my-app"
    echo "  $0 ./my-app mycompany.com/apps/my-app"
    exit 1
fi

if [ -d "$PROJECT_PATH" ]; then
    echo "❌ Error: Directory $PROJECT_PATH already exists"
    exit 1
fi

OLD_MODULE="github.com/ankit-lilly/go-datastar-daisyui-template"

patch_generated_makefile() {
    local file="$1"
    [ -f "$file" ] || return 0

    local tmp_file
    tmp_file="$(mktemp)"

    awk '
    BEGIN { skip_install = 0 }
    /^install:$/ { skip_install = 1; next }
    skip_install == 1 {
        if ($0 ~ /^[^[:space:]].*:/) {
            skip_install = 0
        } else {
            next
        }
    }
    /^setup: install$/ { next }
    /^\t@echo "  install/ { next }
    /^\t@echo "  setup/ { next }
    /Run '\''make install'\'' first to download dependencies/ {
        print "\t\techo \"Frontend assets are missing in static/css\"; \\"
        next
    }
    /^\.PHONY:/ {
        print ".PHONY: build run dev clean templ fmt css deps test help clean-all"
        next
    }
    { print }
    ' "$file" > "$tmp_file"

    mv "$tmp_file" "$file"
}

patch_generated_readme() {
    local file="$1"
    [ -f "$file" ] || return 0

    local tmp_file
    tmp_file="$(mktemp)"

    awk '
    /│   └── install\// { next }
    /│       └── main.go[[:space:]]+# Install script/ { next }
    /^The install script/ {
        print "During project creation, the template downloads:"
        next
    }
    /^make install/ { next }
    /go run \.\/cmd\/install/ { next }
    /^make setup/ { next }
    /install[[:space:]]+# Download Tailwind/ { next }
    /setup[[:space:]]+# Alias for install/ { next }
    /├── scripts\// { next }
    /│   ├── create\.sh/ { next }
    /│   ├── setup\.sh/ { next }
    /│   └── new-project\.sh/ { next }
    /^# Install dependencies and run$/ { print "# Run"; next }
    { print }
    ' "$file" > "$tmp_file"

    mv "$tmp_file" "$file"
}

echo "🚀 Creating new project: $MODULE_NAME"
echo "   Location: $PROJECT_PATH"
echo ""

# Copy template (excluding generated files and git)
echo "  📁 Copying template files..."
mkdir -p "$PROJECT_PATH"
rsync -a \
    --exclude='.git' \
    --exclude='bin/' \
    --exclude='/scripts/' \
    --exclude='cmd/install/' \
    --exclude='static/css/tailwindcss' \
    --exclude='static/css/daisyui*.mjs' \
    --exclude='static/css/input.css' \
    --exclude='static/css/output.css' \
    --exclude='static/js/datastar.js' \
    --exclude='internal/views/*_templ.go' \
    --exclude='docs/' \
    --exclude='daisuidocs.txt' \
    --exclude='llms.md' \
    --exclude='go.sum' \
    "$TEMPLATE_DIR/" "$PROJECT_PATH/"

# Sanity check: ensure core scaffold files were copied
if [ ! -f "$PROJECT_PATH/go.mod" ] || [ ! -f "$PROJECT_PATH/cmd/server/main.go" ]; then
    echo "❌ Error: Template copy incomplete (missing Go scaffold files)"
    exit 1
fi

# Update module name in go.mod
echo "  📝 Updating go.mod..."
sed -i '' "s|$OLD_MODULE|$MODULE_NAME|g" "$PROJECT_PATH/go.mod"

# Update import paths in all Go files
echo "  📝 Updating import paths..."
find "$PROJECT_PATH" -type f -name "*.go" -exec sed -i '' "s|$OLD_MODULE|$MODULE_NAME|g" {} \;

# Update import paths in templ files
find "$PROJECT_PATH" -type f -name "*.templ" -exec sed -i '' "s|$OLD_MODULE|$MODULE_NAME|g" {} \;

# Remove template-only install workflow from generated project docs/build file
patch_generated_makefile "$PROJECT_PATH/Makefile"
patch_generated_readme "$PROJECT_PATH/README.md"

# Create directories
mkdir -p "$PROJECT_PATH/static/css" "$PROJECT_PATH/static/js" "$PROJECT_PATH/bin"

# Initialize git
echo "  📦 Initializing git repository..."
cd "$PROJECT_PATH"
git init -q

# Download dependencies and run setup
echo "  📦 Running setup..."
bash "$TEMPLATE_DIR/scripts/setup.sh" "static"

echo ""
echo "✅ Project created successfully!"
echo ""
echo "Next steps:"
echo "  cd $PROJECT_PATH"
echo "  make run      - Build and run the server"
echo "  make dev      - Run in development mode with watchers"
echo ""
echo "Then open http://localhost:8080"
