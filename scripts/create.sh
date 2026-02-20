#!/usr/bin/env bash

# Go + Templ + Datastar + DaisyUI Project Creator
# Usage: curl -fsSL https://raw.githubusercontent.com/ankit-lilly/go-datastar-daisyui-template/main/scripts/create.sh | bash
#
# Or with arguments:
# curl -fsSL ... | bash -s -- ./my-app github.com/myuser/my-app

set -euo pipefail

REPO_URL="https://github.com/ankit-lilly/go-datastar-daisyui-template"
TARBALL_URL="$REPO_URL/archive/refs/heads/main.tar.gz"
OLD_MODULE="github.com/ankit-lilly/go-datastar-daisyui-template"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

print_banner() {
    echo ""
    echo -e "${BLUE}╔═══════════════════════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║${NC}     ${GREEN}Go + Templ + Datastar + DaisyUI${NC} Project Creator      ${BLUE}║${NC}"
    echo -e "${BLUE}╚═══════════════════════════════════════════════════════════╝${NC}"
    echo ""
}

print_step() {
    echo -e "${GREEN}▸${NC} $1"
}

print_error() {
    echo -e "${RED}✖ Error:${NC} $1" >&2
}

print_success() {
    echo -e "${GREEN}✔${NC} $1"
}

# Check for required tools
check_requirements() {
    local missing=()

    for cmd in go curl tar sed rsync git; do
        if ! command -v "$cmd" &> /dev/null; then
            missing+=("$cmd")
        fi
    done

    if [ ${#missing[@]} -ne 0 ]; then
        print_error "Missing required tools: ${missing[*]}"
        echo ""
        echo "Please install them and try again:"
        echo ""
        if [[ " ${missing[*]} " =~ " rsync " ]]; then
            echo "  macOS:  brew install rsync"
            echo "  Ubuntu: sudo apt install rsync"
            echo "  Fedora: sudo dnf install rsync"
        fi
        if [[ " ${missing[*]} " =~ " go " ]]; then
            echo "  Go:     https://go.dev/dl/"
        fi
        echo ""
        exit 1
    fi

    # Check Go version (need 1.24+ for tool directive)
    local go_version=$(go version | grep -oE 'go[0-9]+\.[0-9]+' | sed 's/go//')
    local major=$(echo "$go_version" | cut -d. -f1)
    local minor=$(echo "$go_version" | cut -d. -f2)

    if [ "$major" -lt 1 ] || ([ "$major" -eq 1 ] && [ "$minor" -lt 24 ]); then
        print_error "Go 1.24 or higher is required (found go$go_version)"
        echo ""
        echo "  Download: https://go.dev/dl/"
        echo ""
        exit 1
    fi
}

# Prompt for input with default value
prompt() {
    local prompt_text="$1"
    local default="$2"
    local var_name="$3"

    if [ -n "$default" ]; then
        echo -ne "${BLUE}?${NC} $prompt_text ${YELLOW}[$default]${NC}: "
    else
        echo -ne "${BLUE}?${NC} $prompt_text: "
    fi

    # Read from /dev/tty to handle curl | bash
    read -r input < /dev/tty

    if [ -z "$input" ]; then
        eval "$var_name='$default'"
    else
        eval "$var_name='$input'"
    fi
}

# Prompt for yes/no
prompt_yn() {
    local prompt_text="$1"
    local default="$2"

    if [ "$default" = "y" ]; then
        echo -ne "${BLUE}?${NC} $prompt_text ${YELLOW}[Y/n]${NC}: "
    else
        echo -ne "${BLUE}?${NC} $prompt_text ${YELLOW}[y/N]${NC}: "
    fi

    # Read from /dev/tty to handle curl | bash
    read -r input < /dev/tty
    input="${input:-$default}"

    [[ "$input" =~ ^[Yy]$ ]]
}

# Main
main() {
    print_banner
    check_requirements

    # Get project directory
    local project_dir=""
    local module_name=""

    # Check if arguments were passed
    if [ $# -ge 2 ]; then
        project_dir="$1"
        module_name="$2"
    else
        # Interactive mode
        echo "Let's create your new Go web application!"
        echo ""

        # Get project directory
        prompt "Project directory" "./my-app" project_dir

        # Expand ~ to home directory
        project_dir="${project_dir/#\~/$HOME}"

        # Convert to absolute path
        if [[ "$project_dir" != /* ]]; then
            project_dir="$(pwd)/$project_dir"
        fi

        # Suggest module name based on directory
        local suggested_module=""
        local dir_name=$(basename "$project_dir")

        # Try to detect GitHub username from git config
        local git_user=$(git config --global user.name 2>/dev/null | tr '[:upper:]' '[:lower:]' | tr ' ' '-' || echo "myuser")
        suggested_module="github.com/$git_user/$dir_name"

        prompt "Go module name" "$suggested_module" module_name
    fi

    # Validate inputs
    if [ -z "$project_dir" ]; then
        print_error "Project directory is required"
        exit 1
    fi

    if [ -z "$module_name" ]; then
        print_error "Module name is required"
        exit 1
    fi

    if [ -d "$project_dir" ]; then
        print_error "Directory already exists: $project_dir"
        exit 1
    fi

    echo ""
    echo -e "${BLUE}Creating project:${NC}"
    echo "  Directory: $project_dir"
    echo "  Module:    $module_name"
    echo ""

    if [ $# -lt 2 ]; then
        if ! prompt_yn "Continue?" "y"; then
            echo "Aborted."
            exit 0
        fi
        echo ""
    fi

    # Create temp directory for download
    local tmp_dir=$(mktemp -d)
    trap "rm -rf $tmp_dir" EXIT

    # Download template
    print_step "Downloading template..."
    curl -fsSL "$TARBALL_URL" | tar -xz -C "$tmp_dir"

    # Find extracted directory (it includes the branch name)
    local extracted_dir=$(ls "$tmp_dir")
    local template_dir="$tmp_dir/$extracted_dir"

    # Create project directory
    print_step "Creating project directory..."
    mkdir -p "$project_dir"

    # Copy files (excluding unnecessary ones)
    print_step "Copying files..."
    rsync -a \
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
        --exclude='.git/' \
        "$template_dir/" "$project_dir/"

    # Update module name
    print_step "Updating module name..."

    # Update go.mod
    if [[ "$OSTYPE" == "darwin"* ]]; then
        sed -i '' "s|$OLD_MODULE|$module_name|g" "$project_dir/go.mod"
        find "$project_dir" -type f -name "*.go" -exec sed -i '' "s|$OLD_MODULE|$module_name|g" {} \;
        find "$project_dir" -type f -name "*.templ" -exec sed -i '' "s|$OLD_MODULE|$module_name|g" {} \;
    else
        sed -i "s|$OLD_MODULE|$module_name|g" "$project_dir/go.mod"
        find "$project_dir" -type f -name "*.go" -exec sed -i "s|$OLD_MODULE|$module_name|g" {} \;
        find "$project_dir" -type f -name "*.templ" -exec sed -i "s|$OLD_MODULE|$module_name|g" {} \;
    fi

    # Create directories
    mkdir -p "$project_dir/static/css" "$project_dir/static/js" "$project_dir/bin"

    # Initialize git
    print_step "Initializing git repository..."
    cd "$project_dir"
    git init -q

    # Generate templ files first (needed before go mod tidy)
    print_step "Generating templ files..."
    go run github.com/a-h/templ/cmd/templ@latest generate

    # Download dependencies
    print_step "Downloading Go dependencies..."
    go mod tidy

    # Run install
    print_step "Installing frontend dependencies (Tailwind, DaisyUI, Datastar)..."
    go run ./cmd/install

    echo ""
    echo -e "${GREEN}═══════════════════════════════════════════════════════════${NC}"
    print_success "Project created successfully!"
    echo -e "${GREEN}═══════════════════════════════════════════════════════════${NC}"
    echo ""
    echo "Next steps:"
    echo ""
    echo -e "  ${BLUE}cd${NC} $project_dir"
    echo -e "  ${BLUE}make run${NC}        # Start the server"
    echo -e "  ${BLUE}make dev${NC}        # Development mode with hot-reload"
    echo ""
    echo -e "Then open ${BLUE}http://localhost:8080${NC}"
    echo ""
}

main "$@"
