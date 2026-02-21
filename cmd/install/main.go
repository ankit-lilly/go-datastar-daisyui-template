package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
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

	// Run independent setup tasks concurrently to reduce total install time.
	if err := runParallel(
		task{
			name: "go dependencies",
			fn:   downloadDeps,
		},
		task{
			name: "tailwind",
			fn: func() error {
				return downloadTailwind(cssDir)
			},
		},
		task{
			name: "daisyui",
			fn: func() error {
				return downloadDaisyUI(cssDir)
			},
		},
		task{
			name: "datastar",
			fn: func() error {
				return downloadDatastar(jsDir)
			},
		},
		task{
			name: "input.css",
			fn: func() error {
				return createInputCSS(cssDir)
			},
		},
	); err != nil {
		fatal("Setup failed: %v", err)
	}

	// Generate templ files
	generateTempl()

	// Build CSS
	if err := buildCSS(cssDir); err != nil {
		fatal("Failed to build CSS: %v", err)
	}

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

func downloadDeps() error {
	fmt.Println("  📦 Downloading Go dependencies...")

	cmd := exec.Command("go", "mod", "tidy")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	fmt.Println("  ✅ Go dependencies ready (templ available via 'go tool templ')")
	return nil
}

func downloadTailwind(cssDir string) error {
	fmt.Println("  📦 Downloading Tailwind CSS (latest)...")

	filename := buildTailwindFilename()
	url := fmt.Sprintf("%s/%s", tailwindBaseURL, filename)
	destPath := filepath.Join(cssDir, "tailwindcss")

	if err := downloadFile(url, destPath); err != nil {
		return err
	}

	if err := os.Chmod(destPath, 0755); err != nil {
		return err
	}

	fmt.Println("  ✅ Tailwind CSS downloaded")
	return nil
}

func downloadDaisyUI(cssDir string) error {
	fmt.Println("  📦 Downloading DaisyUI (latest)...")

	if err := runParallel(
		task{
			name: "daisyui.mjs",
			fn: func() error {
				url := fmt.Sprintf("%s/daisyui.mjs", daisyUIBaseURL)
				destPath := filepath.Join(cssDir, "daisyui.mjs")
				return downloadFile(url, destPath)
			},
		},
		task{
			name: "daisyui-theme.mjs",
			fn: func() error {
				url := fmt.Sprintf("%s/daisyui-theme.mjs", daisyUIBaseURL)
				destPath := filepath.Join(cssDir, "daisyui-theme.mjs")
				return downloadFile(url, destPath)
			},
		},
	); err != nil {
		return err
	}

	fmt.Println("  ✅ DaisyUI downloaded")
	return nil
}

func downloadDatastar(jsDir string) error {
	fmt.Println("  📦 Downloading Datastar v" + datastarVersion + "...")

	destPath := filepath.Join(jsDir, "datastar.js")
	if err := downloadFile(datastarURL, destPath); err != nil {
		return err
	}

	fmt.Println("  ✅ Datastar v" + datastarVersion + " downloaded")
	return nil
}

func createInputCSS(cssDir string) error {
	content := `@import "tailwindcss";

@source "../../internal/views/**/*.templ";
@source not "./tailwindcss";
@source not "./daisyui{,*}.mjs";

@plugin "./daisyui.mjs";`

	destPath := filepath.Join(cssDir, "input.css")
	if err := os.WriteFile(destPath, []byte(content), 0644); err != nil {
		return err
	}
	fmt.Println("  ✅ Created input.css")
	return nil
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

func buildCSS(cssDir string) error {
	fmt.Println("  🔨 Building CSS...")

	// Run from cssDir, so use relative paths
	cmd := exec.Command("./tailwindcss", "-i", "input.css", "-o", "output.css", "--minify")
	cmd.Dir = cssDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	fmt.Println("  ✅ CSS built")
	return nil
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

type task struct {
	name string
	fn   func() error
}

func runParallel(tasks ...task) error {
	var wg sync.WaitGroup
	errCh := make(chan error, len(tasks))

	for _, t := range tasks {
		t := t
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := t.fn(); err != nil {
				errCh <- fmt.Errorf("%s: %w", t.name, err)
			}
		}()
	}

	wg.Wait()
	close(errCh)

	var errs []error
	for err := range errCh {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}
