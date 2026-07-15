package iso

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/bartrosa/homelab-cli/internal/exec"
)

// VerifyGPG downloads checksums and verifies GPG signature using system gpg.
func VerifyGPG(ctx context.Context, rel Release) error {
	if rel.GPGKeyURL == "" {
		return nil
	}
	runner := exec.NewOSRunner(os.Stdout, os.Stderr)
	tmp, err := os.MkdirTemp("", "lab-iso-gpg-*")
	if err != nil {
		return err
	}
	defer func() { _ = os.RemoveAll(tmp) }()

	client := &http.Client{Timeout: 2 * time.Minute}
	sumPath := filepath.Join(tmp, "SHA256SUMS")
	gpgPath := filepath.Join(tmp, "SHA256SUMS.gpg")

	sumData, err := fetchBytes(ctx, client, rel.ChecksumURL)
	if err != nil {
		return err
	}
	if err := os.WriteFile(sumPath, sumData, 0o644); err != nil {
		return err
	}

	gpgData, err := fetchBytes(ctx, client, rel.GPGKeyURL)
	if err != nil {
		return err
	}
	if err := os.WriteFile(gpgPath, gpgData, 0o644); err != nil {
		return err
	}

	if err := runner.Run(ctx, "gpg", "--verify", gpgPath, sumPath); err != nil {
		return fmt.Errorf("gpg verify: %w", err)
	}
	return nil
}

// ParseSHA256SUMS extracts hash for filename from Ubuntu-style SHA256SUMS.
func ParseSHA256SUMS(content, filename string) (string, bool) {
	return findSHA256(content, filename)
}

// ParseFedoraCHECKSUM extracts hash from Fedora CHECKSUM file.
func ParseFedoraCHECKSUM(content, filename string) (string, bool) {
	return findFedoraChecksum(content, filename)
}

// ReadISOSize returns human-readable size.
func ReadISOSize(path string) (string, error) {
	st, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	return formatBytes(st.Size()), nil
}

func formatBytes(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for v := n / unit; v >= unit; v /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(n)/float64(div), "KMGTPE"[exp])
}

// CopyWithProgress is a test hook placeholder.
func CopyWithProgress(dst io.Writer, src io.Reader) (int64, error) {
	return io.Copy(dst, src)
}
