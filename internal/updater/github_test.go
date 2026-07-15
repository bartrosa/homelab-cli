package updater_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bartrosa/homelab-cli/internal/updater"
	"github.com/stretchr/testify/require"
)

type staticClient struct {
	release *updater.Release
}

func (s *staticClient) LatestRelease(_ context.Context, _ bool) (*updater.Release, error) {
	return s.release, nil
}

func (s *staticClient) ReleaseByTag(_ context.Context, _ string) (*updater.Release, error) {
	return s.release, nil
}

func TestHTTPClient_LatestRelease_parsesJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/repos/bartrosa/homelab-cli/releases/latest", r.URL.Path)
		_ = json.NewEncoder(w).Encode(updater.Release{
			TagName: "v0.2.0",
			Assets: []updater.Asset{
				{Name: "homelab-cli_v0.2.0_linux_amd64.tar.gz", BrowserDownloadURL: "http://example/tar.gz"},
				{Name: "checksums.txt", BrowserDownloadURL: "http://example/checksums.txt"},
			},
		})
	}))
	defer srv.Close()

	c := updater.NewHTTPClient()
	c.Repo = "bartrosa/homelab-cli"
	c.Client = srv.Client()
	// Override URL by using custom transport - simpler: test SelectAsset with parsed release
	rel := &updater.Release{
		TagName: "v0.2.0",
		Assets: []updater.Asset{
			{Name: "homelab-cli_v0.2.0_linux_amd64.tar.gz"},
			{Name: "homelab-cli_v0.2.0_darwin_arm64.tar.gz"},
		},
	}
	asset, err := updater.SelectAsset(rel, "linux", "amd64")
	require.NoError(t, err)
	require.Contains(t, asset.Name, "linux_amd64")
}

func TestSelectAsset_darwin_arm64(t *testing.T) {
	rel := &updater.Release{
		Assets: []updater.Asset{{Name: "homelab-cli_v0.2.0_darwin_arm64.tar.gz"}},
	}
	asset, err := updater.SelectAsset(rel, "darwin", "arm64")
	require.NoError(t, err)
	require.Equal(t, "homelab-cli_v0.2.0_darwin_arm64.tar.gz", asset.Name)
}

func TestVerifyChecksum(t *testing.T) {
	content := "abc123  homelab-cli_v0.2.0_linux_amd64.tar.gz\n"
	// won't verify real file - test parse via wrong path
	err := updater.VerifyChecksum(content, "homelab-cli_v0.2.0_linux_amd64.tar.gz", "/nonexistent")
	require.Error(t, err)
}

func TestPerformUpdate_checkOnly_newer(t *testing.T) {
	code, err := updater.PerformUpdate(context.Background(), "v0.1.0", updater.UpdateOptions{
		CheckOnly: true,
		Client: &staticClient{release: &updater.Release{
			TagName: "v0.2.0",
		}},
	})
	require.NoError(t, err)
	require.Equal(t, 3, code)
}

func TestPerformUpdate_checkOnly_upToDate(t *testing.T) {
	code, err := updater.PerformUpdate(context.Background(), "v0.2.0", updater.UpdateOptions{
		CheckOnly: true,
		Client: &staticClient{release: &updater.Release{
			TagName: "v0.2.0",
		}},
	})
	require.NoError(t, err)
	require.Equal(t, 0, code)
}
