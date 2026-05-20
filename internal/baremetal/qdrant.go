package baremetal

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/bartrosa/homelab-cli/internal/executil"
)

// InstallQdrant installs Qdrant from GitHub release (Linux bare metal).
func InstallQdrant(ctx context.Context, dryRun bool, stdout, stderr io.Writer) error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("qdrant bare-metal install supports linux only (got %s)", runtime.GOOS)
	}
	for _, bin := range []string{"curl", "tar", "sudo"} {
		if !executil.CommandExists(bin) {
			return fmt.Errorf("%s not found in PATH", bin)
		}
	}
	ex := executil.NewRunner(stdout, stderr)
	ex.DryRun = dryRun

	installDir := envOr("QDRANT_INSTALL_DIR", "/opt/qdrant")
	dataDir := "/var/lib/qdrant"
	if os.Getenv("QDRANT_USE_HOT_STORAGE") != "" {
		if _, err := os.Stat("/mnt/data-hot"); err == nil {
			dataDir = "/mnt/data-hot/qdrant"
		}
	}

	arch, asset, err := qdrantAsset()
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "=== Qdrant bare metal (%s) ===\n", arch)

	tag := os.Getenv("QDRANT_TAG")
	if tag == "" {
		tag, err = latestQdrantTag(ctx, ex)
		if err != nil {
			return err
		}
	}
	binPath := filepath.Join(installDir, "qdrant")
	if _, err := os.Stat(binPath); os.IsNotExist(err) {
		url := fmt.Sprintf("https://github.com/qdrant/qdrant/releases/download/%s/%s", tag, asset)
		fmt.Fprintf(stdout, "Pobieranie %s...\n", tag)
		if err := ex.Run(ctx, "bash", "-c",
			fmt.Sprintf(`set -euo pipefail
tmpdir=$(mktemp -d)
curl -fsSL -o "$tmpdir/qdrant.tgz" %q
tar -xzf "$tmpdir/qdrant.tgz" -C "$tmpdir"
sudo mkdir -p %q
sudo install -m 0755 "$(find "$tmpdir" -name qdrant -type f | head -1)" %q
rm -rf "$tmpdir"`, url, installDir, binPath)); err != nil {
			return fmt.Errorf("install qdrant binary: %w", err)
		}
	} else {
		fmt.Fprintln(stdout, "Binarka Qdrant już jest w", installDir)
	}

	configYAML := fmt.Sprintf(`log_level: INFO
storage:
  storage_path: %s/storage
  snapshots_path: %s/snapshots
service:
  host: 0.0.0.0
  http_port: 6333
  grpc_port: 6334
`, dataDir, dataDir)

	script := fmt.Sprintf(`set -euo pipefail
if ! getent group qdrant >/dev/null; then sudo groupadd --system qdrant; fi
if ! getent passwd qdrant >/dev/null; then sudo useradd --system --home-dir %q --no-create-home --gid qdrant qdrant; fi
sudo mkdir -p %q/storage %q/snapshots
sudo chown -R qdrant:qdrant %q
cat <<'QEOF' | sudo tee %q/config.yaml >/dev/null
%s
QEOF
`, installDir, dataDir, dataDir, dataDir, installDir, configYAML)

	if err := ex.Run(ctx, "bash", "-c", script); err != nil {
		return fmt.Errorf("qdrant config: %w", err)
	}

	unit := `[Unit]
Description=Qdrant vector search
After=network.target

[Service]
Type=simple
User=qdrant
Group=qdrant
ExecStart=` + binPath + ` --config-path ` + filepath.Join(installDir, "config.yaml") + `
Restart=on-failure

[Install]
WantedBy=multi-user.target
`
	if err := ex.Run(ctx, "bash", "-c", fmt.Sprintf(
		`printf %%s %q | sudo tee /etc/systemd/system/qdrant.service >/dev/null
sudo systemctl daemon-reload
sudo systemctl enable --now qdrant.service`, unit)); err != nil {
		return fmt.Errorf("qdrant systemd: %w", err)
	}
	fmt.Fprintln(stdout, "Qdrant: http://<host>:6333/dashboard")
	return nil
}

func qdrantAsset() (arch, asset string, err error) {
	switch runtime.GOARCH {
	case "amd64":
		return "x86_64", "qdrant-x86_64-unknown-linux-musl.tar.gz", nil
	case "arm64":
		return "aarch64", "qdrant-aarch64-unknown-linux-musl.tar.gz", nil
	default:
		return "", "", fmt.Errorf("unsupported arch %s", runtime.GOARCH)
	}
}

func latestQdrantTag(ctx context.Context, ex *executil.Runner) (string, error) {
	out, err := ex.Output(ctx, "curl", "-sSf", "https://api.github.com/repos/qdrant/qdrant/releases/latest")
	if err != nil {
		return "", err
	}
	var rel struct {
		TagName string `json:"tag_name"`
	}
	if err := json.Unmarshal(out, &rel); err != nil {
		return "", err
	}
	if rel.TagName == "" {
		return "", fmt.Errorf("empty tag from GitHub API")
	}
	return rel.TagName, nil
}

func envOr(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}
