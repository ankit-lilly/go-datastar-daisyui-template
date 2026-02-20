package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	tailwindBaseURL = "https://github.com/tailwindlabs/tailwindcss/releases/latest/download"
	daisyUIBaseURL  = "https://github.com/saadeghi/daisyui/releases/latest/download"
	datastarVersion = "1.0.0-RC.7"
	datastarURL     = "https://cdn.jsdelivr.net/gh/starfederation/datastar@" + datastarVersion + "/bundles/datastar.js"
)

func main() {
	staticDir := "static"
	if len(os.Args) > 1 {
		staticDir = os.Args[1]
	}

	cssDir := filepath.Join(staticDir, "css")
	jsDir := filepath.Join(staticDir, "js")

	// Create directories
	if err := os.MkdirAll(cssDir, 0755); err != nil {
		fatal("Failed to create css directory: %v", err)
	}
	if err := os.MkdirAll(jsDir, 0755); err != nil {
		fatal("Failed to create js directory: %v", err)
	}

	fmt.Printf("🚀 Setting up Go + Templ + Datastar + DaisyUI template for %s/%s\n\n", runtime.GOOS, runtime.GOARCH)

	// Download Go dependencies (including templ tool)
	downloadDeps()

	// Download Tailwind CSS (latest)
	downloadTailwind(cssDir)

	// Download DaisyUI (latest)
	downloadDaisyUI(cssDir)

	// Download Datastar (latest)
	downloadDatastar(jsDir)

	// Create input.css
	createInputCSS(cssDir)

	// Generate templ files
	generateTempl()

	// Build CSS
	buildCSS(cssDir)

	fmt.Println("\n✅ Setup complete!")
	fmt.Println("\nFiles created:")
	fmt.Printf("  - %s/tailwindcss (binary)\n", cssDir)
	fmt.Printf("  - %s/daisyui.mjs\n", cssDir)
	fmt.Printf("  - %s/daisyui-theme.mjs\n", cssDir)
	fmt.Printf("  - %s/input.css\n", cssDir)
	fmt.Printf("  - %s/output.css\n", cssDir)
	fmt.Printf("  - %s/datastar.js\n", jsDir)
	fmt.Println("\nNext steps:")
	fmt.Println("  make build   - Build the server")
	fmt.Println("  make run     - Build and run the server")
	fmt.Println("  make dev     - Run in development mode with watchers")
}

func downloadDeps() {
	fmt.Println("  📦 Downloading Go dependencies...")

	cmd := exec.Command("go", "mod", "tidy")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fatal("Failed to download dependencies: %v", err)
	}
	fmt.Println("  ✅ Go dependencies ready (templ available via 'go tool templ')")
}

func downloadTailwind(cssDir string) {
	fmt.Println("  📦 Downloading Tailwind CSS (latest)...")

	filename := buildTailwindFilename()
	url := fmt.Sprintf("%s/%s", tailwindBaseURL, filename)
	destPath := filepath.Join(cssDir, "tailwindcss")

	if err := downloadFile(url, destPath); err != nil {
		fatal("Failed to download Tailwind CSS: %v", err)
	}

	if err := os.Chmod(destPath, 0755); err != nil {
		fatal("Failed to make Tailwind executable: %v", err)
	}

	fmt.Println("  ✅ Tailwind CSS downloaded")
}

func downloadDaisyUI(cssDir string) {
	fmt.Println("  📦 Downloading DaisyUI (latest)...")

	// Download daisyui.mjs
	url := fmt.Sprintf("%s/daisyui.mjs", daisyUIBaseURL)
	destPath := filepath.Join(cssDir, "daisyui.mjs")
	if err := downloadFile(url, destPath); err != nil {
		fatal("Failed to download daisyui.mjs: %v", err)
	}

	// Download daisyui-theme.mjs
	url = fmt.Sprintf("%s/daisyui-theme.mjs", daisyUIBaseURL)
	destPath = filepath.Join(cssDir, "daisyui-theme.mjs")
	if err := downloadFile(url, destPath); err != nil {
		fatal("Failed to download daisyui-theme.mjs: %v", err)
	}

	fmt.Println("  ✅ DaisyUI downloaded")
}

func downloadDatastar(jsDir string) {
	fmt.Println("  📦 Downloading Datastar v" + datastarVersion + "...")

	destPath := filepath.Join(jsDir, "datastar.js")
	if err := downloadFile(datastarURL, destPath); err != nil {
		fatal("Failed to download Datastar: %v", err)
	}

	fmt.Println("  ✅ Datastar v" + datastarVersion + " downloaded")
}

func createInputCSS(cssDir string) {
	content := `@import "tailwindcss";

@source "../../internal/views/**/*.templ";
@source not "./tailwindcss";
@source not "./daisyui{,*}.mjs";

@plugin "./daisyui.mjs";`

	destPath := filepath.Join(cssDir, "input.css")
	if err := os.WriteFile(destPath, []byte(content), 0644); err != nil {
		fatal("Failed to create input.css: %v", err)
	}
	fmt.Println("  ✅ Created input.css")
}

func generateTempl() {
	fmt.Println("  🔨 Generating templ files...")

	cmd := exec.Command("go", "tool", "templ", "generate")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Println("  ⚠️  templ generate failed (this is normal for first run)")
		return
	}
	fmt.Println("  ✅ templ files generated")
}

func buildCSS(cssDir string) {
	fmt.Println("  🔨 Building CSS...")

	// Run from cssDir, so use relative paths
	cmd := exec.Command("./tailwindcss", "-i", "input.css", "-o", "output.css", "--minify")
	cmd.Dir = cssDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fatal("Failed to build CSS: %v", err)
	}
	fmt.Println("  ✅ CSS built")
}

func downloadFile(url, destPath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download returned status %d for %s", resp.StatusCode, url)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func buildTailwindFilename() string {
	osName := runtime.GOOS
	arch := runtime.GOARCH

	// Map Go OS to Tailwind naming convention
	if osName == "darwin" {
		osName = "macos"
	}

	// Map Go arch to Tailwind naming convention
	archName := arch
	switch arch {
	case "amd64":
		archName = "x64"
	case "arm64":
		archName = "arm64"
	}

	// Check for musl on Linux
	muslSuffix := ""
	if runtime.GOOS == "linux" && isMusl() {
		muslSuffix = "-musl"
	}

	return fmt.Sprintf("tailwindcss-%s-%s%s", osName, archName, muslSuffix)
}

func isMusl() bool {
	cmd := exec.Command("ldd", "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}
	return strings.Contains(string(output), "musl")
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "❌ "+format+"\n", args...)
	os.Exit(1)
}
