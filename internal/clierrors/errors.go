// Package clierrors defines stable sentinel errors shared across CLI packages.
package clierrors

import "errors"

// ErrNotImplemented is returned by scaffolded commands until their adapters exist.
var ErrNotImplemented = errors.New("not implemented yet")
