// Command lab is the homelab automation CLI entrypoint.
package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"

	"github.com/bartrosa/homelab-cli/internal/cli"

	"github.com/bartrosa/homelab-cli/internal/clierrors"

	_ "github.com/bartrosa/homelab-cli/internal/services/register"
	_ "github.com/bartrosa/homelab-cli/internal/stack/components"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	err := cli.Execute(ctx)
	stop()

	if err != nil {
		var exitErr *clierrors.ExitError
		if errors.As(err, &exitErr) {
			if exitErr.Msg != "" {
				cli.PrintCommandError(err)
			}
			os.Exit(exitErr.ExitCode())
		}
		cli.PrintCommandError(err)
		os.Exit(1)
	}
}
