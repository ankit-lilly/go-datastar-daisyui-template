# Go + Templ + Datastar + DaisyUI Template

A modern, server-rendered Go web application template featuring:

- **Go 1.24+** with new HTTP handler patterns and `go tool` directive
- **Templ** for type-safe, compiled HTML templates
- **Structured logging** with `log/slog`
- **Graceful shutdown** handling
- **Datastar** for reactive frontend without npm
- **DaisyUI 5** for beautiful UI components (always latest version)
- **Tailwind CSS 4** (standalone binary, no Node.js required)
- **Background job hub** for long-running tasks with SSE progress updates

## Create a New Project

### One-liner (Recommended)

```bash
curl -fsSL https://raw.githubusercontent.com/ankit-lilly/go-datastar-daisyui-template/main/scripts/create.sh | bash
```

You'll be prompted for:
1. **Project directory** - Where to create the project (default: `./my-app`)
2. **Go module name** - Your module path (default: `github.com/<your-username>/<project-name>`)

The script will automatically:
- Download the template
- Update module name and import paths
- Initialize git
- Install dependencies (Tailwind, DaisyUI, Datastar)
- Generate templ files and build CSS

### Non-interactive

```bash
curl -fsSL https://raw.githubusercontent.com/ankit-lilly/go-datastar-daisyui-template/main/scripts/create.sh | bash -s -- ./my-app github.com/myuser/my-app
```

### Manual Setup

If you prefer to set things up manually:

```bash
# Clone the template
git clone https://github.com/ankit-lilly/go-datastar-daisyui-template my-app
cd my-app

# Remove template git history
rm -rf .git && git init

# Update module name in go.mod and all import paths
# (find/replace github.com/ankit-lilly/go-datastar-daisyui-template with your module)

# Install dependencies and run
make install
make run
```

## Quick Start (for development on the template itself)

```bash
make install   # Download Tailwind, DaisyUI, Datastar, setup templ
make run       # Build and run the server

# Open http://localhost:8080
```

## Project Structure

```
.
├── cmd/
│   ├── server/
│   │   └── main.go           # Application entry point
│   └── install/
│       └── main.go           # Install script (downloads dependencies)
├── internal/
│   ├── config/
│   │   └── config.go         # Configuration management
│   ├── handlers/
│   │   └── handlers.go       # HTTP handlers
│   ├── jobs/
│   │   └── hub.go            # Background job hub
│   ├── util/
│   │   └── id.go             # Utility functions
│   └── views/
│       ├── components.templ  # Shared components (navbar, footer, etc.)
│       ├── index.templ       # Home page
│       └── demo.templ        # Demo page with examples
├── static/
│   ├── css/
│   │   ├── tailwindcss       # Tailwind binary (downloaded)
│   │   ├── daisyui.mjs       # DaisyUI plugin (downloaded)
│   │   ├── input.css         # CSS input file (generated)
│   │   └── output.css        # Compiled CSS (generated)
│   └── js/
│       └── datastar.js       # Datastar library (downloaded)
├── scripts/
│   ├── create.sh             # curl-able project creator
│   ├── setup.sh              # Setup script (bash version)
│   └── new-project.sh        # Create new project (local use)
├── Makefile
├── go.mod                    # Includes templ as a tool dependency
└── README.md
```

## Development

### Development Mode

Run with hot-reload for templ and CSS:

```bash
make dev
```

This starts:
- templ file watcher (regenerates Go code on .templ changes)
- CSS watcher (rebuilds output.css on changes)
- Go server

### Manual Development

```bash
# Terminal 1: Watch templ files
go tool templ generate --watch

# Terminal 2: Watch CSS
cd static/css && ./tailwindcss -i input.css -o output.css --watch

# Terminal 3: Run server
go run ./cmd/server
```

### Available Make Targets

```bash
make install    # Download Tailwind, DaisyUI, Datastar and setup templ
make setup      # Alias for install
make templ      # Generate Go code from templ files
make build      # Generate templ and build the Go binary
make run        # Build and run the server
make dev        # Run in development mode with watchers
make css        # Rebuild CSS only
make deps       # Download Go dependencies
make test       # Run tests
make clean      # Remove build artifacts and generated templ files
make clean-all  # Remove all generated files
```

## How It Works

### Dependencies (No npm required!)

The install script (`make install` or `go run ./cmd/install`) downloads:

| Dependency | Source | Purpose |
|------------|--------|---------|
| Tailwind CSS | GitHub releases (latest) | CSS utility framework (standalone binary) |
| DaisyUI | GitHub releases (latest) | UI component library (CSS plugin) |
| Datastar | jsDelivr CDN (latest) | Reactive frontend via SSE |
| Templ | go.mod tool directive | Type-safe HTML templates |

### Templ (via `go tool`)

Templ is managed as a tool dependency in `go.mod`:

```go
tool github.com/a-h/templ/cmd/templ
```

This means:
- Version is pinned to your project
- No global installation needed
- Run with `go tool templ generate`
- Automatically available after `go mod tidy`

### CSS Build Pipeline

```
internal/views/*.templ  →  Tailwind scans for classes  →  output.css
```

The `input.css` configures Tailwind to scan templ files:

```css
@import "tailwindcss";
@source "../../internal/views/**/*.templ";
@plugin "./daisyui.mjs";
```

## Templ Components

This template uses [templ](https://templ.guide) for type-safe HTML templates:

### Component Definition

```go
// internal/views/components.templ
templ Card(title, description string) {
    <div class="card bg-base-200">
        <div class="card-body">
            <h2 class="card-title">{ title }</h2>
            <p>{ description }</p>
        </div>
    </div>
}
```

### Using Components

```go
// internal/views/index.templ
templ IndexPage() {
    @Base("Home") {
        @Navbar("home")
        @Card("Hello", "World")
        @Footer()
    }
}
```

### Rendering in Handlers

```go
func (h *Handlers) Index(w http.ResponseWriter, r *http.Request) {
    views.IndexPage().Render(r.Context(), w)
}
```

## Datastar Usage

Datastar provides reactive frontend capabilities through HTML attributes:

### Two-way Binding

```html
<input data-bind:name>
<span data-text="$name"></span>
```

### SSE Requests

```html
<button data-on:click="@get('/api/data')">Load</button>
<button data-on:click="@post('/api/submit')">Submit</button>
```

### Conditional Display

```html
<div data-show="$isVisible">Shown when isVisible is true</div>
```

### Server-Side (Go)

```go
sse := datastar.NewSSE(w, r)

// Patch DOM elements
sse.PatchElements(`<div id="content">Updated!</div>`)

// Patch signals
sse.PatchSignals([]byte(`{"status": "done"}`))
```

## DaisyUI Components

This template includes DaisyUI 5 (always downloads latest version) with all its components. See the demo page for examples.

Common components:

- `btn`, `btn-primary`, `btn-outline`
- `card`, `card-body`, `card-title`
- `input`, `select`, `checkbox`, `toggle`
- `alert`, `badge`, `progress`
- `navbar`, `footer`, `menu`
- `modal`, `drawer`, `dropdown`

## Background Jobs

The template includes a job hub for running background tasks:

```go
// Create a job
job := jobHub.NewJob("my-task", func(j *jobs.Job) error {
    for i := 0; i <= 100; i += 10 {
        j.SetProgress(i)
        time.Sleep(time.Second)
    }
    return nil
})

// Submit for execution
jobHub.Submit(job)

// Stream progress to client
for update := range job.Updates() {
    sse.PatchElements(fmt.Sprintf(`<progress value="%d" max="100">`, update.Progress))
}
```

## Configuration

Environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `ADDR`   | `:8080` | Server address |
| `ENV`    | `development` | Environment name |

## License

MIT
