package updater

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// UpdateOptions configures a self-update run.
type UpdateOptions struct {
	ForceVersion      string
	IncludePrerelease bool
	Yes               bool
	CheckOnly         bool
	Stdout            io.Writer
	Stderr            io.Writer
	Client            ReleaseClient
}

// PerformUpdate checks for updates and optionally installs.
func PerformUpdate(ctx context.Context, currentVersion string, opts UpdateOptions) (int, error) {
	if opts.Client == nil {
		opts.Client = NewHTTPClient()
	}
	if opts.Stdout == nil {
		opts.Stdout = os.Stdout
	}
	if opts.Stderr == nil {
		opts.Stderr = os.Stderr
	}

	var release *Release
	var err error
	if opts.ForceVersion != "" {
		release, err = opts.Client.ReleaseByTag(ctx, opts.ForceVersion)
	} else {
		release, err = opts.Client.LatestRelease(ctx, opts.IncludePrerelease)
	}
	if err != nil {
		return 1, fmt.Errorf("fetch release: %w", err)
	}

	remote := release.TagName
	cmp := CompareVersions(currentVersion, remote)
	if opts.CheckOnly {
		if cmp >= 0 {
			fmt.Fprintf(opts.Stdout, "already up to date (%s)\n", currentVersion)
			return 0, nil
		}
		fmt.Fprintf(opts.Stdout, "update available: %s → %s\n", currentVersion, remote)
		return 3, nil
	}

	if opts.ForceVersion == "" && cmp >= 0 {
		fmt.Fprintf(opts.Stdout, "already up to date (%s)\n", currentVersion)
		return 0, nil
	}

	fmt.Fprintf(opts.Stdout, "Updating %s → %s\n", currentVersion, remote)

	exe, err := CurrentExecutable()
	if err != nil {
		return 1, err
	}
	if NeedsSudo(exe) {
		return 1, fmt.Errorf("cannot write to %s: re-run with sudo", exe)
	}

	goos, goarch := Platform()
	asset, err := SelectAsset(release, goos, goarch)
	if err != nil {
		return 1, err
	}

	client := NewHTTPClient().Client
	tmpdir, err := os.MkdirTemp("", "lab-update-*")
	if err != nil {
		return 1, err
	}
	defer func() { _ = os.RemoveAll(tmpdir) }()

	archive := filepath.Join(tmpdir, asset.Name)
	fmt.Fprintf(opts.Stdout, "Downloading %s\n", asset.Name)
	if err := Download(ctx, client, asset.BrowserDownloadURL, archive, opts.Stdout); err != nil {
		return 1, err
	}
	fmt.Fprintln(opts.Stdout)

	// Verify checksums
	var checksumAsset *Asset
	for i := range release.Assets {
		if release.Assets[i].Name == "checksums.txt" {
			checksumAsset = &release.Assets[i]
			break
		}
	}
	if checksumAsset != nil {
		checksumsPath := filepath.Join(tmpdir, "checksums.txt")
		if err := Download(ctx, client, checksumAsset.BrowserDownloadURL, checksumsPath, nil); err != nil {
			return 1, fmt.Errorf("download checksums: %w", err)
		}
		data, err := os.ReadFile(checksumsPath)
		if err != nil {
			return 1, err
		}
		if err := VerifyChecksum(string(data), asset.Name, archive); err != nil {
			return 1, err
		}
		fmt.Fprintln(opts.Stdout, "Checksum verified")
	}

	newBin := filepath.Join(tmpdir, "lab")
	if err := ExtractLabBinary(archive, newBin); err != nil {
		return 1, err
	}

	if err := Replace(exe, newBin); err != nil {
		return 1, fmt.Errorf("install: %w", err)
	}

	fmt.Fprintf(opts.Stdout, "Updated to %s\n", remote)
	return 0, nil
}
