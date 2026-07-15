// Package clierrors defines stable sentinel errors shared across CLI packages.
package clierrors

import (
	"errors"
	"fmt"
)

// ErrNotImplemented is returned by scaffolded commands until their adapters exist.
var ErrNotImplemented = errors.New("not implemented yet")

// ExitError carries a specific process exit code.
type ExitError struct {
	Code int
	Msg  string
}

func (e *ExitError) Error() string {
	if e.Msg == "" {
		return fmt.Sprintf("exit %d", e.Code)
	}
	return e.Msg
}

// ExitCode returns the desired exit code.
func (e *ExitError) ExitCode() int {
	return e.Code
}
