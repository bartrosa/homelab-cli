package baremetal

import (
	"context"
	"fmt"
	"io"
	"os"
	"runtime"

	"github.com/bartrosa/homelab-cli/internal/executil"
)

// InstallClickHouse installs ClickHouse from official apt repo (Debian/Ubuntu).
func InstallClickHouse(ctx context.Context, dryRun bool, stdout, stderr io.Writer) error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("clickhouse bare-metal install supports linux only (got %s)", runtime.GOOS)
	}
	for _, bin := range []string{"curl", "sudo", "apt-get"} {
		if !executil.CommandExists(bin) {
			return fmt.Errorf("%s not found in PATH", bin)
		}
	}
	ex := executil.NewRunner(stdout, stderr)
	ex.DryRun = dryRun
	hot := os.Getenv("CLICKHOUSE_USE_HOT_STORAGE")
	dataPath := envOr("CLICKHOUSE_DATA_PATH", "/mnt/data-hot/clickhouse")
	script := fmt.Sprintf(`set -euo pipefail
echo "=== ClickHouse bare metal ==="
if [ ! -f /usr/share/keyrings/clickhouse-keyring.gpg ]; then
  sudo apt-get update -qq
  sudo apt-get install -y apt-transport-https ca-certificates curl gnupg
  curl -fsSL 'https://packages.clickhouse.com/rpm/lts/repodata/repomd.xml.key' | sudo gpg --dearmor -o /usr/share/keyrings/clickhouse-keyring.gpg
  ARCH=$(dpkg --print-architecture)
  echo "deb [signed-by=/usr/share/keyrings/clickhouse-keyring.gpg arch=${ARCH}] https://packages.clickhouse.com/deb stable main" | sudo tee /etc/apt/sources.list.d/clickhouse.list
  sudo apt-get update -qq
fi
if ! dpkg -l clickhouse-server &>/dev/null; then
  sudo apt-get install -y clickhouse-server clickhouse-client
fi
if [ -n %q ] && [ -d /mnt/data-hot ]; then
  sudo mkdir -p %q
  sudo chown clickhouse:clickhouse %q || true
  if [ -f /etc/clickhouse-server/config.xml ]; then
    sudo sed -i 's|<path>.*</path>|<path>%s/</path>|' /etc/clickhouse-server/config.xml || true
  fi
fi
sudo systemctl enable --now clickhouse-server || true
echo "ClickHouse: http://<host>:8123"
`, hot, dataPath, dataPath, dataPath)
	return ex.Run(ctx, "bash", "-c", script)
}
