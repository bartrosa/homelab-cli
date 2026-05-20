package system

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"
)

const (
	ubuntuMetaLTS   = "https://changelogs.ubuntu.com/meta-release-lts"
	ubuntuMeta      = "https://changelogs.ubuntu.com/meta-release"
	fedoraReleases = "https://dl.fedoraproject.org/pub/fedora/linux/releases/"
	fedoraISOBase  = "https://dl.fedoraproject.org/pub/fedora/linux/releases/%d/Silverblue/x86_64/iso/"
)

// BootImage is a downloadable bootable ISO discovered from upstream mirrors.
type BootImage struct {
	ID          string
	Label       string
	Series      string
	Kind        string // ubuntu-lts, ubuntu, fedora-silverblue
	spec        isoSpec
}

var ubuntuDesktopISO = regexp.MustCompile(`(?m)^([a-f0-9]{64})\s+\*?(ubuntu-[0-9]+\.[0-9]+(?:\.[0-9]+)?-desktop-amd64\.iso)\*?`)

// DiscoverBootImages queries Ubuntu and Fedora mirrors for current desktop/Silverblue ISOs.
func DiscoverBootImages(ctx context.Context) ([]BootImage, error) {
	client := &http.Client{Timeout: 2 * time.Minute}
	var out []BootImage

	ltsData, err := fetchURL(ctx, client, ubuntuMetaLTS)
	if err != nil {
		return nil, fmt.Errorf("ubuntu meta-release-lts: %w", err)
	}
	metaData, err := fetchURL(ctx, client, ubuntuMeta)
	if err != nil {
		return nil, fmt.Errorf("ubuntu meta-release: %w", err)
	}

	ltsSeries := filterSupported(parseMetaRelease(ltsData), true, false)
	// Two newest supported LTS releases that still publish desktop ISOs (skip EOL < 22.04).
	var ltsAdded int
	for _, series := range ltsSeries {
		if ltsAdded >= 2 {
			break
		}
		if !recentUbuntuSeries(series) {
			continue
		}
		img, err := resolveUbuntuDesktop(ctx, client, series, "ubuntu-lts")
		if err != nil {
			continue
		}
		img.Label = fmt.Sprintf("Ubuntu %s LTS (desktop)", series)
		out = append(out, img)
		ltsAdded++
	}

	interim := filterSupported(parseMetaRelease(metaData), false, true)
	for _, series := range interim {
		if !recentUbuntuSeries(series) {
			continue
		}
		img, err := resolveUbuntuDesktop(ctx, client, series, "ubuntu")
		if err != nil {
			continue
		}
		img.Label = fmt.Sprintf("Ubuntu %s (latest interim desktop)", series)
		out = append(out, img)
		break
	}

	fedoraImg, err := discoverFedoraSilverblue(ctx, client)
	if err != nil {
		return nil, err
	}
	out = append(out, fedoraImg)

	if len(out) == 0 {
		return nil, fmt.Errorf("no boot images discovered")
	}
	return out, nil
}

func resolveUbuntuDesktop(ctx context.Context, client *http.Client, series, kind string) (BootImage, error) {
	base := fmt.Sprintf("https://releases.ubuntu.com/%s/", series)
	sumsURL := base + "SHA256SUMS"
	data, err := fetchURL(ctx, client, sumsURL)
	if err != nil {
		return BootImage{}, err
	}
	m := ubuntuDesktopISO.FindStringSubmatch(string(data))
	if m == nil {
		return BootImage{}, fmt.Errorf("no desktop-amd64.iso in %s", sumsURL)
	}
	isoFile := m[2]
	return BootImage{
		ID:     fmt.Sprintf("%s-%s", kind, series),
		Series: series,
		Kind:   kind,
		spec: isoSpec{
			isoURL:       base + isoFile,
			checksumURL:  sumsURL,
			isoFile:      isoFile,
			checksumFile: "SHA256SUMS",
			fedoraStyle:  false,
		},
	}, nil
}

func discoverFedoraSilverblue(ctx context.Context, client *http.Client) (BootImage, error) {
	idx, err := fetchURL(ctx, client, fedoraReleases)
	if err != nil {
		return BootImage{}, fmt.Errorf("fedora releases index: %w", err)
	}
	release, err := latestNumericDir(idx, 30)
	if err != nil {
		return BootImage{}, err
	}
	isoDir := fmt.Sprintf(fedoraISOBase, release)
	dirData, err := fetchURL(ctx, client, isoDir)
	if err != nil {
		return BootImage{}, fmt.Errorf("fedora silverblue iso dir: %w", err)
	}
	names := parseApacheIndex(dirData)
	var isoFile, checksumFile string
	for _, n := range names {
		if strings.HasSuffix(n, ".iso") && strings.Contains(n, "Silverblue") {
			isoFile = n
		}
		if strings.Contains(n, "CHECKSUM") && strings.Contains(n, "Silverblue") {
			checksumFile = n
		}
	}
	if isoFile == "" {
		return BootImage{}, fmt.Errorf("no Silverblue iso in Fedora %d", release)
	}
	if checksumFile == "" {
		return BootImage{}, fmt.Errorf("no CHECKSUM file alongside %s in Fedora %d", isoFile, release)
	}
	return BootImage{
		ID:     fmt.Sprintf("fedora-silverblue-%d", release),
		Label:  fmt.Sprintf("Fedora %d Silverblue (latest)", release),
		Series: fmt.Sprintf("%d", release),
		Kind:   "fedora-silverblue",
		spec: isoSpec{
			isoURL:       isoDir + isoFile,
			checksumURL:  isoDir + checksumFile,
			isoFile:      isoFile,
			checksumFile: checksumFile,
			fedoraStyle:  true,
		},
	}, nil
}

// ResolveBootImage finds a discovered image by distro flag (e.g. ubuntu-latest, ubuntu-lts-24.04, fedora-silverblue).
func ResolveBootImage(ctx context.Context, distro string) (isoSpec, string, error) {
	distro = strings.TrimSpace(strings.ToLower(distro))
	if distro == "" {
		distro = "ubuntu-latest"
	}
	images, err := DiscoverBootImages(ctx)
	if err != nil {
		return isoSpec{}, "", err
	}
	switch distro {
	case "ubuntu-latest", "ubuntu":
		for _, img := range images {
			if img.Kind == "ubuntu" {
				return img.spec, img.Label, nil
			}
		}
	case "ubuntu-lts", "ubuntu-lts-latest":
		for _, img := range images {
			if img.Kind == "ubuntu-lts" {
				return img.spec, img.Label, nil
			}
		}
	case "fedora-silverblue", "fedora":
		for _, img := range images {
			if img.Kind == "fedora-silverblue" {
				return img.spec, img.Label, nil
			}
		}
	default:
		// ubuntu-lts-24.04, ubuntu-25.10, fedora-silverblue-43
		for _, img := range images {
			if img.ID == distro || img.ID == strings.ReplaceAll(distro, "ubuntu-", "ubuntu-lts-") {
				return img.spec, img.Label, nil
			}
		}
		// allow ubuntu-24.04 -> resolve series directly
		if strings.HasPrefix(distro, "ubuntu-") {
			series := strings.TrimPrefix(distro, "ubuntu-")
			series = strings.TrimPrefix(series, "lts-")
			kind := "ubuntu"
			if strings.Contains(distro, "lts") {
				kind = "ubuntu-lts"
			}
			client := &http.Client{Timeout: 2 * time.Minute}
			img, err := resolveUbuntuDesktop(ctx, client, series, kind)
			if err != nil {
				return isoSpec{}, "", err
			}
			return img.spec, fmt.Sprintf("Ubuntu %s", series), nil
		}
		if strings.HasPrefix(distro, "fedora-silverblue-") {
			client := &http.Client{Timeout: 2 * time.Minute}
			img, err := discoverFedoraSilverblue(ctx, client)
			if err != nil {
				return isoSpec{}, "", err
			}
			return img.spec, img.Label, nil
		}
	}
	ids := make([]string, len(images))
	for i, img := range images {
		ids[i] = img.ID
	}
	sort.Strings(ids)
	return isoSpec{}, "", fmt.Errorf("unknown distro %q (run: lab system usb list); known: %s", distro, strings.Join(ids, ", "))
}

func fetchURL(ctx context.Context, client *http.Client, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "homelab-cli/1.0 (+https://github.com/bartrosa/homelab-cli)")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: %s", url, resp.Status)
	}
	const maxBody = 32 << 20
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxBody))
	if err != nil {
		return nil, err
	}
	return data, nil
}

// PrintBootImageList writes discovered images to w.
func PrintBootImageList(ctx context.Context, w io.Writer) error {
	images, err := DiscoverBootImages(ctx)
	if err != nil {
		return err
	}
	fmt.Fprintln(w, "Odpytywanie Ubuntu (meta-release) i Fedora (dl.fedoraproject.org)…")
	fmt.Fprintln(w)
	for _, img := range images {
		fmt.Fprintf(w, "  %-28s %s\n", img.ID, img.Label)
		fmt.Fprintf(w, "    %s\n", img.spec.isoURL)
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Użycie: lab system usb --distro <ID> --device /dev/sdX")
	return nil
}
