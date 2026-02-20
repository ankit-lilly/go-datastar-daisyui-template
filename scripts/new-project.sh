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

echo "🚀 Creating new project: $MODULE_NAME"
echo "   Location: $PROJECT_PATH"
echo ""

# Copy template (excluding generated files and git)
echo "  📁 Copying template files..."
mkdir -p "$PROJECT_PATH"
rsync -a \
    --exclude='.git' \
    --exclude='bin/' \
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

# Update module name in go.mod
echo "  📝 Updating go.mod..."
sed -i '' "s|$OLD_MODULE|$MODULE_NAME|g" "$PROJECT_PATH/go.mod"

# Update import paths in all Go files
echo "  📝 Updating import paths..."
find "$PROJECT_PATH" -type f -name "*.go" -exec sed -i '' "s|$OLD_MODULE|$MODULE_NAME|g" {} \;

# Update import paths in templ files
find "$PROJECT_PATH" -type f -name "*.templ" -exec sed -i '' "s|$OLD_MODULE|$MODULE_NAME|g" {} \;

# Create directories
mkdir -p "$PROJECT_PATH/static/css" "$PROJECT_PATH/static/js" "$PROJECT_PATH/bin"

# Initialize git
echo "  📦 Initializing git repository..."
cd "$PROJECT_PATH"
git init -q

# Download dependencies and run setup
echo "  📦 Downloading Go dependencies..."
go mod tidy

echo "  📦 Running install script..."
go run ./cmd/install

echo ""
echo "✅ Project created successfully!"
echo ""
echo "Next steps:"
echo "  cd $PROJECT_PATH"
echo "  make run      - Build and run the server"
echo "  make dev      - Run in development mode with watchers"
echo ""
echo "Then open http://localhost:8080"
