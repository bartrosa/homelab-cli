package iso

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const (
	fedoraReleaseAPI = "https://bodhi.fedoraproject.org/releases/?state=current"
	fedoraBase       = "https://download.fedoraproject.org/pub/fedora/linux/releases/"
)

var fedoraISORe = regexp.MustCompile(`Fedora-Silverblue-ostree-(x86_64|aarch64)-(\d+)-1\.(\d+)\.iso`)

func resolveFedoraSilverblue(arch, version string) (Release, error) {
	if arch == "" {
		arch = "amd64"
	}
	farch := "x86_64"
	if arch == "arm64" {
		farch = "aarch64"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	release := strings.TrimSpace(version)
	if release == "" {
		var err error
		release, err = currentFedoraRelease(ctx)
		if err != nil {
			release = "41" // fallback pin
		}
	}

	isoDir := fmt.Sprintf("%s%s/Silverblue/%s/iso/", fedoraBase, release, farch)
	client := &http.Client{Timeout: 2 * time.Minute}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, isoDir, nil)
	if err != nil {
		return Release{}, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return Release{}, err
	}
	defer closeBody(resp)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Release{}, err
	}

	var isoFile, subrelease string
	for _, m := range fedoraISORe.FindAllStringSubmatch(string(body), -1) {
		if m[1] != farch {
			continue
		}
		isoFile = m[0]
		subrelease = m[3]
		break
	}
	if isoFile == "" {
		return Release{}, fmt.Errorf("no Silverblue ISO found for Fedora %s %s", release, farch)
	}

	checksumFile := fmt.Sprintf("Fedora-Silverblue-%s-%s-%s-CHECKSUM", release, subrelease, farch)
	return Release{
		Version:         release,
		ISOURL:          isoDir + isoFile,
		ChecksumURL:     isoDir + checksumFile,
		ChecksumSigKind: SigClearsigned,
		SigningKeyURLs:  []string{"https://fedoraproject.org/fedora.gpg"},
		ISOFilename:     isoFile,
	}, nil
}

func currentFedoraRelease(ctx context.Context) (string, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fedoraReleaseAPI, nil)
	if err != nil {
		return "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer closeBody(resp)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bodhi API: HTTP %d", resp.StatusCode)
	}

	var releases []struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return "", err
	}
	for _, r := range releases {
		name := strings.TrimSpace(r.Name)
		if name != "" {
			return name, nil
		}
	}
	return "", fmt.Errorf("no current fedora release from bodhi")
}
