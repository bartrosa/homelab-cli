//go:build linux

package iso

import (
	"context"
	"fmt"
	"os"
	osexec "os/exec"

	iexec "github.com/bartrosa/homelab-cli/internal/exec"
)

func runPrivileged(ctx context.Context, runner iexec.Runner, name string, args ...string) error {
	if os.Geteuid() == 0 {
		return runner.Run(ctx, name, args...)
	}
	if _, err := osexec.LookPath("sudo"); err != nil {
		return fmt.Errorf("%s requires root privileges; install sudo or run: sudo lab iso write …", name)
	}
	sudoArgs := append([]string{name}, args...)
	return runner.Run(ctx, "sudo", sudoArgs...)
}

func tryUmount(ctx context.Context, runner iexec.Runner, target string) {
	_ = runner.Run(ctx, "umount", target)
	if os.Geteuid() != 0 {
		_ = runner.Run(ctx, "sudo", "umount", target)
	}
}
