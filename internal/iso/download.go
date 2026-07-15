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
)

// DownloadOptions configures ISO download.
type DownloadOptions struct {
	Arch      string
	Version   string
	OutputDir string
	NoVerify  bool
	Force     bool
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

	fmt.Fprintf(opts.Stdout, "Downloading %s\n", rel.ISOURL)
	if err := downloadFile(ctx, opts.Client, rel.ISOURL, dest, opts.Stdout); err != nil {
		return DownloadResult{}, err
	}
	fmt.Fprintln(opts.Stdout)

	sumData, err := fetchBytes(ctx, opts.Client, rel.ChecksumURL)
	if err != nil {
		return DownloadResult{}, fmt.Errorf("checksum file: %w", err)
	}
	want, ok := findSHA256(string(sumData), rel.ISOFilename)
	if !ok {
		// Fedora CHECKSUM format
		want, ok = findFedoraChecksum(string(sumData), rel.ISOFilename)
	}
	if !ok {
		return DownloadResult{}, fmt.Errorf("checksum entry not found for %s", rel.ISOFilename)
	}
	if err := verifyFileSHA256(dest, want); err != nil {
		return DownloadResult{}, err
	}
	fmt.Fprintln(opts.Stdout, "SHA256 verified")

	if !opts.NoVerify && rel.GPGKeyURL != "" {
		if err := VerifyGPG(ctx, rel); err != nil {
			return DownloadResult{}, err
		}
		fmt.Fprintln(opts.Stdout, "GPG signature verified")
	}

	fmt.Fprintf(opts.Stdout, "Verified ISO: %s\n", dest)
	return DownloadResult{Path: dest}, nil
}

func downloadFile(ctx context.Context, client *http.Client, url, dest string, progress io.Writer) error {
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

	pr := &byteProgress{w: progress, total: resp.ContentLength}
	if _, err := io.Copy(f, io.TeeReader(resp.Body, pr)); err != nil {
		return err
	}
	pr.finish()
	return nil
}

type byteProgress struct {
	w       io.Writer
	total   int64
	read    int64
	last    time.Time
	lastPct int
}

func (p *byteProgress) Write(b []byte) (int, error) {
	n := len(b)
	p.read += int64(n)
	now := time.Now()
	if p.w != nil && (p.last.IsZero() || now.Sub(p.last) >= 500*time.Millisecond) {
		p.last = now
		if p.total > 0 {
			pct := int(p.read * 100 / p.total)
			if pct != p.lastPct {
				fmt.Fprintf(p.w, "\rProgress: %d%% (%d / %d bytes)", pct, p.read, p.total)
				p.lastPct = pct
			}
		} else {
			fmt.Fprintf(p.w, "\rDownloaded %d bytes", p.read)
		}
	}
	return n, nil
}

func (p *byteProgress) finish() {
	if p.w != nil {
		fmt.Fprintln(p.w)
	}
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
