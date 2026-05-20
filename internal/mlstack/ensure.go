// Package mlstack ensures the ML compose stack is running.
package mlstack

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/bartrosa/homelab-cli/internal/executil"
)

// EnsureUp runs podman-compose up -d in ml-stackDir and prints service URLs.
func EnsureUp(ctx context.Context, mlStackDir string, serverIP string, dryRun bool, stdout, stderr io.Writer) error {
	dir, err := filepath.Abs(mlStackDir)
	if err != nil {
		return err
	}
	if _, err := os.Stat(filepath.Join(dir, "podman-compose.yml")); err != nil {
		return fmt.Errorf("podman-compose.yml not found in %s", dir)
	}
	if !executil.CommandExists("podman-compose") {
		return fmt.Errorf("podman-compose not found; run: lab bootstrap server")
	}
	ex := executil.NewRunner(stdout, stderr)
	ex.DryRun = dryRun
	ex.WorkDir = dir
	fmt.Fprintln(stdout, "=== Podnoszenie ml-stack (podman-compose up -d) ===")
	if err := ex.Run(ctx, "podman-compose", "up", "-d"); err != nil {
		return err
	}
	ip := serverIP
	if ip == "" {
		ip = "127.0.0.1"
	}
	printURLs(stdout, ip)
	return nil
}

func printURLs(w io.Writer, ip string) {
	fmt.Fprintf(w, "\nGotowe. Używaj portu 8080:\n")
	fmt.Fprintf(w, "  MLflow:        http://%s:8080/mlflow\n", ip)
	fmt.Fprintf(w, "  Registry UI:   http://%s:8080/registry/\n", ip)
	fmt.Fprintf(w, "  Redis Insight: http://%s:8080/redis\n", ip)
	fmt.Fprintf(w, "  Langfuse:      http://%s:3001\n", ip)
	fmt.Fprintf(w, "  Langflow:      http://%s:7860\n", ip)
	fmt.Fprintf(w, "  Traefik:       http://%s:8081/dashboard/\n", ip)
}
