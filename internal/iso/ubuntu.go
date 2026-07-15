package iso

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const ubuntuBase = "https://releases.ubuntu.com/"

var ubuntuDirRe = regexp.MustCompile(`(?m)<a href="([0-9]+\.[0-9]+(?:\.[0-9]+)?)/">`)

func resolveUbuntuDesktop(arch, version string) (Release, error) {
	if arch == "" {
		arch = "amd64"
	}
	if arch != "amd64" {
		return Release{}, fmt.Errorf("ubuntu-desktop: only amd64 supported")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	ver := strings.TrimSpace(version)
	if ver == "" {
		var err error
		ver, err = latestUbuntuLTSVersion(ctx)
		if err != nil {
			return Release{}, err
		}
	}

	base := ubuntuBase + ver + "/"
	isoFile := fmt.Sprintf("ubuntu-%s-desktop-%s.iso", ver, arch)
	return Release{
		Version:     ver + " LTS",
		ISOURL:      base + isoFile,
		ChecksumURL: base + "SHA256SUMS",
		ChecksumSigURL: base + "SHA256SUMS.gpg",
		SigningKeyIDs: []string{
			"843938DF228D22F7B3742BC0D94AA3F0EFE21092", // Ubuntu CD Image Automatic Signing Key (2012)
			"46181433FBB75451",                         // Ubuntu CD Image Signing Key (legacy)
		},
		ISOFilename: isoFile,
	}, nil
}

func latestUbuntuLTSVersion(ctx context.Context) (string, error) {
	client := &http.Client{Timeout: 2 * time.Minute}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ubuntuBase, nil)
	if err != nil {
		return "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer closeBody(resp)
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var versions []string
	for _, m := range ubuntuDirRe.FindAllStringSubmatch(string(data), -1) {
		v := m[1]
		if strings.HasPrefix(v, "24.") || strings.HasPrefix(v, "22.") {
			versions = append(versions, v)
		}
	}
	if len(versions) == 0 {
		return "", fmt.Errorf("no ubuntu LTS version found in index")
	}
	// Pick highest semver-ish
	best := versions[0]
	for _, v := range versions[1:] {
		if compareUbuntuVer(v, best) > 0 {
			best = v
		}
	}
	return best, nil
}

func compareUbuntuVer(a, b string) int {
	ap := strings.Split(a, ".")
	bp := strings.Split(b, ".")
	for i := 0; i < 3; i++ {
		ai, bi := 0, 0
		if i < len(ap) {
			ai, _ = strconv.Atoi(ap[i])
		}
		if i < len(bp) {
			bi, _ = strconv.Atoi(bp[i])
		}
		if ai != bi {
			return ai - bi
		}
	}
	return 0
}
