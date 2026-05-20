package baremetal

import (
	"context"
	"fmt"
	"io"
	"os"
	"runtime"

	"github.com/bartrosa/homelab-cli/internal/executil"
)

// InstallMilvus installs Milvus standalone from GitHub DEB (Debian/Ubuntu).
func InstallMilvus(ctx context.Context, dryRun bool, stdout, stderr io.Writer) error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("milvus bare-metal install supports linux only (got %s)", runtime.GOOS)
	}
	for _, bin := range []string{"curl", "sudo", "dpkg"} {
		if !executil.CommandExists(bin) {
			return fmt.Errorf("%s not found in PATH", bin)
		}
	}
	ex := executil.NewRunner(stdout, stderr)
	ex.DryRun = dryRun
	version := envOr("MILVUS_VERSION", "2.6.9")
	hot := os.Getenv("MILVUS_USE_HOT_STORAGE")
	script := fmt.Sprintf(`set -euo pipefail
ARCH=$(dpkg --print-architecture)
case "$ARCH" in amd64) PKG_ARCH=amd64;; arm64) PKG_ARCH=arm64;; *) echo "unsupported arch"; exit 1;; esac
DEB_NAME="milvus_%s-1_${PKG_ARCH}.deb"
URL="https://github.com/milvus-io/milvus/releases/download/v%s/${DEB_NAME}"
echo "=== Milvus bare metal v%s ==="
if ! dpkg -l milvus &>/dev/null; then
  TMP_DEB=$(mktemp -t milvus_XXXXXX.deb)
  curl -fsSL -o "$TMP_DEB" "$URL"
  sudo apt-get install -y "$TMP_DEB"
  rm -f "$TMP_DEB"
else
  echo "Milvus już zainstalowany."
fi
if [ -n %q ] && [ -d /mnt/data-hot ]; then
  sudo mkdir -p /mnt/data-hot/milvus
  if [ -f /etc/milvus/config.yaml ]; then
    sudo sed -i 's|dataPath:.*|dataPath: /mnt/data-hot/milvus|' /etc/milvus/config.yaml || true
  fi
fi
sudo systemctl enable --now milvus || true
echo "Milvus: http://<host>:9091"
`, version, version, version, hot)
	return ex.Run(ctx, "bash", "-c", script)
}
