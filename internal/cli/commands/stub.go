package commands

import (
	"fmt"

	"github.com/bartrosa/homelab-cli/internal/clierrors"

	"github.com/spf13/cobra"
)

// StubRunE returns a Cobra RunE that always fails with ErrNotImplemented.
func StubRunE() func(*cobra.Command, []string) error {
	return func(*cobra.Command, []string) error {
		return fmt.Errorf("%w", clierrors.ErrNotImplemented)
	}
}
