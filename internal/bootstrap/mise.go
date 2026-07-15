package bootstrap

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/bartrosa/homelab-cli/internal/exec"
)

// InstallMiseScript is the official mise installer URL.
const InstallMiseScript = "https://mise.run"

// RunMiseInstall runs the upstream mise installer when mise is absent.
func RunMiseInstall(ctx context.Context, runner exec.Runner, stdout io.Writer) error {
	if _, err := runner.RunWithOutput(ctx, "mise", "--version"); err == nil {
		fmt.Fprintln(stdout, "mise already installed")
		return nil
	}
	home, _ := os.UserHomeDir()
	miseBin := home + "/.local/bin/mise"
	if st, err := os.Stat(miseBin); err == nil && !st.IsDir() {
		fmt.Fprintln(stdout, "mise binary present at", miseBin)
		return nil
	}
	return runner.Run(ctx, "bash", "-c", "curl https://mise.run | sh")
}
