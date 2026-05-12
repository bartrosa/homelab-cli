// Command lab is the homelab automation CLI entrypoint.
package main

import (
	"__MODULE_PATH__/internal/cli"
	"context"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	err := cli.Execute(ctx)
	stop()

	if err != nil {
		cli.PrintCommandError(err)
		os.Exit(1)
	}
}
