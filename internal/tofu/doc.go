// Package tofu installs the OpenTofu binary and orchestrates tofu commands
// against IaC stack directories. Install delegates to OS-native methods;
// orchestration wraps the tofu CLI via the shared exec runner.
package tofu
