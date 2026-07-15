// Package updater handles self-update from GitHub releases.
package updater

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// GitHubRepo is the owner/repo slug used for release downloads.
const GitHubRepo = "bartrosa/homelab-cli"

// ReleaseClient fetches release metadata and assets.
type ReleaseClient interface {
	LatestRelease(ctx context.Context, includePrerelease bool) (*Release, error)
	ReleaseByTag(ctx context.Context, tag string) (*Release, error)
}

// HTTPClient implements ReleaseClient against the GitHub REST API.
type HTTPClient struct {
	Repo   string
	Client *http.Client
}

// Release is a GitHub release with downloadable assets.
type Release struct {
	TagName    string  `json:"tag_name"`
	Prerelease bool    `json:"prerelease"`
	Assets     []Asset `json:"assets"`
}

// Asset is a release attachment.
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// NewHTTPClient returns a client with sensible defaults.
func NewHTTPClient() *HTTPClient {
	return &HTTPClient{
		Repo: GitHubRepo,
		Client: &http.Client{
			Timeout: 5 * time.Minute,
		},
	}
}

// LatestRelease returns the newest matching release.
func (c *HTTPClient) LatestRelease(ctx context.Context, includePrerelease bool) (*Release, error) {
	if includePrerelease {
		return c.latestFromList(ctx, true)
	}
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", c.Repo)
	return c.fetchRelease(ctx, url)
}

// ReleaseByTag fetches a specific release tag.
func (c *HTTPClient) ReleaseByTag(ctx context.Context, tag string) (*Release, error) {
	tag = strings.TrimSpace(tag)
	if tag == "" {
		return nil, fmt.Errorf("empty version tag")
	}
	if !strings.HasPrefix(tag, "v") {
		tag = "v" + tag
	}
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/tags/%s", c.Repo, tag)
	return c.fetchRelease(ctx, url)
}

func (c *HTTPClient) latestFromList(ctx context.Context, includePrerelease bool) (*Release, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases", c.Repo)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github releases: HTTP %d", resp.StatusCode)
	}

	var releases []Release
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, err
	}
	for i := range releases {
		if !includePrerelease && releases[i].Prerelease {
			continue
		}
		return &releases[i], nil
	}
	return nil, fmt.Errorf("no releases found")
}

func (c *HTTPClient) fetchRelease(ctx context.Context, url string) (*Release, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github release: HTTP %d", resp.StatusCode)
	}

	var rel Release
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, err
	}
	return &rel, nil
}

// SelectAsset picks the tar.gz asset for the current GOOS/GOARCH.
func SelectAsset(release *Release, goos, goarch string) (*Asset, error) {
	if release == nil {
		return nil, fmt.Errorf("nil release")
	}
	want := fmt.Sprintf("_%s_%s.tar.gz", goos, goarch)
	for i := range release.Assets {
		if strings.HasSuffix(release.Assets[i].Name, want) {
			return &release.Assets[i], nil
		}
	}
	return nil, fmt.Errorf("no asset matching %s for release %s", want, release.TagName)
}

// Download fetches url to dest with optional progress.
func Download(ctx context.Context, client *http.Client, url, dest string, progress io.Writer) error {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Minute}
	}
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
		return fmt.Errorf("download: HTTP %d", resp.StatusCode)
	}

	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}
	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer closeFile(f)

	var r io.Reader = resp.Body
	if progress != nil {
		r = &progressReader{r: resp.Body, w: progress, total: resp.ContentLength}
	}
	_, err = io.Copy(f, r)
	return err
}

type progressReader struct {
	r       io.Reader
	w       io.Writer
	total   int64
	read    int64
	lastPct int
}

func (p *progressReader) Read(b []byte) (int, error) {
	n, err := p.r.Read(b)
	p.read += int64(n)
	if p.total > 0 {
		pct := int(p.read * 100 / p.total)
		if pct != p.lastPct && pct%5 == 0 {
			fmt.Fprintf(p.w, "\rDownloading... %d%%", pct)
			p.lastPct = pct
		}
	}
	return n, err
}

// VerifyChecksum compares file hash against checksums.txt content.
func VerifyChecksum(checksumsContent, filename, filePath string) error {
	want, ok := parseChecksumLine(checksumsContent, filename)
	if !ok {
		return fmt.Errorf("checksum for %q not found in checksums file", filename)
	}
	f, err := os.Open(filePath)
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
		return fmt.Errorf("checksum mismatch for %s: got %s want %s", filename, got, want)
	}
	return nil
}

func parseChecksumLine(content, filename string) (string, bool) {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		name := strings.TrimPrefix(parts[1], "*")
		if name == filename || strings.HasSuffix(name, "/"+filename) {
			return parts[0], true
		}
	}
	return "", false
}

// ExtractLabBinary extracts the lab binary from a goreleaser tar.gz to destPath.
func ExtractLabBinary(archivePath, destPath string) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer closeFile(f)

	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer closeFile(gz)

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if hdr.Typeflag != tar.TypeReg || hdr.Name != "lab" && !strings.HasSuffix(hdr.Name, "/lab") {
			continue
		}
		if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
			return err
		}
		out, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
		if err != nil {
			return err
		}
		if _, err := io.Copy(out, tr); err != nil {
			closeFile(out)
			return err
		}
		return out.Close()
	}
	return fmt.Errorf("lab binary not found in archive")
}

// Replace atomically replaces target with newBinary (same filesystem when possible).
func Replace(target, newBinary string) error {
	info, err := os.Stat(target)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	mode := os.FileMode(0o755)
	if info != nil {
		mode = info.Mode()
	}

	dir := filepath.Dir(target)
	tmp := filepath.Join(dir, ".lab-update-"+fmt.Sprintf("%d", os.Getpid()))
	if err := copyFile(newBinary, tmp, mode); err != nil {
		return err
	}
	if err := os.Rename(tmp, target); err != nil {
		_ = os.Remove(tmp)
		// Fallback: write beside as .new then rename
		alt := target + ".new"
		if err2 := copyFile(newBinary, alt, mode); err2 != nil {
			return fmt.Errorf("replace %s: %w (fallback: %v)", target, err, err2)
		}
		if err3 := os.Rename(alt, target); err3 != nil {
			return fmt.Errorf("replace %s: %w", target, err3)
		}
	}
	return nil
}

func copyFile(src, dst string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer closeFile(in)

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		closeFile(out)
		return err
	}
	return out.Close()
}

// CurrentExecutable returns the path to the running lab binary.
func CurrentExecutable() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(exe)
}

// NeedsSudo reports whether path is in a system directory and not writable.
func NeedsSudo(path string) bool {
	if !strings.HasPrefix(path, "/usr/local/") && !strings.HasPrefix(path, "/usr/bin/") {
		return false
	}
	f, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		return true
	}
	_ = f.Close()
	return false
}

// Platform returns GOOS/GOARCH for asset selection.
func Platform() (goos, goarch string) {
	return runtime.GOOS, runtime.GOARCH
}
