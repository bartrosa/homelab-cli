package iso

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bartrosa/homelab-cli/internal/ui"
)

// DownloadOptions configures ISO download.
type DownloadOptions struct {
	Arch      string
	Version   string
	OutputDir string
	NoVerify  bool
	Force     bool
	NoColor   bool
	Stdout    io.Writer
	Client    *http.Client
}

// DownloadResult holds the verified ISO path.
type DownloadResult struct {
	Path string
}

// DownloadISO resolves, downloads, and verifies an ISO.
func DownloadISO(ctx context.Context, distro Distro, opts DownloadOptions) (DownloadResult, error) {
	if opts.Client == nil {
		opts.Client = &http.Client{Timeout: 30 * time.Minute}
	}
	if opts.Stdout == nil {
		opts.Stdout = os.Stdout
	}
	if opts.OutputDir == "" {
		home, _ := os.UserHomeDir()
		opts.OutputDir = filepath.Join(home, ".cache", "homelab-cli", "iso")
	}
	if err := os.MkdirAll(opts.OutputDir, 0o755); err != nil {
		return DownloadResult{}, err
	}

	rel, err := distro.Resolve(opts.Arch, opts.Version)
	if err != nil {
		return DownloadResult{}, err
	}

	dest := filepath.Join(opts.OutputDir, rel.ISOFilename)
	if !opts.Force {
		if st, err := os.Stat(dest); err == nil && st.Size() > 0 {
			if ok, _ := verifyLocalSHA256(dest, rel); ok {
				fmt.Fprintf(opts.Stdout, "already cached: %s\n", dest)
				return DownloadResult{Path: dest}, nil
			}
		}
	}

	fmt.Fprintf(opts.Stdout, "Downloading %s\n", rel.ISOFilename)
	if err := downloadFile(ctx, opts.Client, rel.ISOURL, dest, opts.Stdout, opts.NoColor); err != nil {
		return DownloadResult{}, err
	}

	if err := ui.RunWithSpinner(opts.Stdout, opts.NoColor, "Verifying SHA256", func() error {
		sumData, err := fetchBytes(ctx, opts.Client, rel.ChecksumURL)
		if err != nil {
			return fmt.Errorf("checksum file: %w", err)
		}
		want, ok := findSHA256(string(sumData), rel.ISOFilename)
		if !ok {
			want, ok = findFedoraChecksum(string(sumData), rel.ISOFilename)
		}
		if !ok {
			return fmt.Errorf("checksum entry not found for %s", rel.ISOFilename)
		}
		return verifyFileSHA256(dest, want)
	}); err != nil {
		return DownloadResult{}, err
	}
	fmt.Fprintln(opts.Stdout, "  SHA256 verified")

	needsGPG := rel.ChecksumSigKind == SigClearsigned || rel.ChecksumSigURL != ""
	if !opts.NoVerify && needsGPG {
		styles := ui.NewStyles(opts.Stdout, opts.NoColor)
		if err := ui.RunWithSpinner(opts.Stdout, opts.NoColor, "Verifying GPG signature", func() error {
			return VerifyGPG(ctx, rel)
		}); err != nil {
			ui.SecurityWarning(opts.Stdout, styles, opts.NoColor,
				"GPG VERIFICATION FAILED",
				"The checksum signature could NOT be verified with upstream signing keys.",
				"This may indicate a tampered download, compromised mirror, or man-in-the-middle attack.",
				"DO NOT install from this ISO: "+dest,
				"Delete the file and re-download, or verify manually (see ubuntu.com/tutorials/how-to-verify-ubuntu).",
				"",
				"Details: "+err.Error(),
			)
			return DownloadResult{}, fmt.Errorf("gpg verification failed: %w", err)
		}
		fmt.Fprintln(opts.Stdout, "  GPG signature verified")
	}

	fmt.Fprintf(opts.Stdout, "Verified ISO: %s\n", dest)
	return DownloadResult{Path: dest}, nil
}

func downloadFile(ctx context.Context, client *http.Client, url, dest string, out io.Writer, noColor bool) error {
	rep := ui.NewDownloadReporter(out, noColor)
	rep.BeginConnect("Connecting to server")
	defer rep.EndConnect()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer closeBody(resp)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer closeFile(f)

	rep.SetTotal(resp.ContentLength)
	if _, err := io.Copy(f, io.TeeReader(resp.Body, rep.Writer())); err != nil {
		return err
	}
	rep.Finish()
	return nil
}

func fetchBytes(ctx context.Context, client *http.Client, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

func verifyLocalSHA256(_ string, _ Release) (bool, error) {
	// Without network, can't verify - return false to trigger redownload
	return false, nil
}

func findSHA256(content, filename string) (string, bool) {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		name := strings.TrimPrefix(parts[1], "*")
		if name == filename {
			return parts[0], true
		}
	}
	return "", false
}

func findFedoraChecksum(content, filename string) (string, bool) {
	for _, line := range strings.Split(content, "\n") {
		if strings.Contains(line, filename) && strings.Contains(line, "SHA256") {
			// SHA256 (filename.iso) = abc...
			if i := strings.Index(line, "= "); i >= 0 {
				return strings.TrimSpace(line[i+2:]), true
			}
		}
	}
	return "", false
}

func verifyFileSHA256(path, want string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer closeFile(f)
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}
	got := hex.EncodeToString(h.Sum(nil))
	if !strings.EqualFold(got, want) {
		return fmt.Errorf("sha256 mismatch: got %s want %s", got, want)
	}
	return nil
}
