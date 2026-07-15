// Package iso provides bootable USB tooling.
package iso

import (
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"text/tabwriter"
)

// ErrNotImplemented indicates a stub resolver.
var ErrNotImplemented = errors.New("not implemented yet")

// Release describes a resolved ISO download.
type Release struct {
	Version     string
	ISOURL      string
	ChecksumURL string
	GPGKeyURL   string
	ISOFilename string
}

// Distro describes a supported distribution.
type Distro struct {
	ID            string
	DisplayName   string
	Architectures []string
	ApproxSize    string
	Resolve       func(arch string, version string) (Release, error)
}

// Catalog returns all known distros.
func Catalog() []Distro {
	return []Distro{
		{
			ID:            "ubuntu-desktop",
			DisplayName:   "Ubuntu Desktop LTS",
			Architectures: []string{"amd64"},
			ApproxSize:    "~6 GB",
			Resolve:       resolveUbuntuDesktop,
		},
		{
			ID:            "ubuntu-server",
			DisplayName:   "Ubuntu Server",
			Architectures: []string{"amd64", "arm64"},
			ApproxSize:    "~3 GB",
			Resolve:       stubResolver("ubuntu-server"),
		},
		{
			ID:            "fedora-silverblue",
			DisplayName:   "Fedora Silverblue",
			Architectures: []string{"amd64"},
			ApproxSize:    "~2.5 GB",
			Resolve:       resolveFedoraSilverblue,
		},
		{
			ID:            "fedora-workstation",
			DisplayName:   "Fedora Workstation",
			Architectures: []string{"amd64", "arm64"},
			ApproxSize:    "~2.5 GB",
			Resolve:       stubResolver("fedora-workstation"),
		},
		{
			ID:            "debian",
			DisplayName:   "Debian",
			Architectures: []string{"amd64", "arm64"},
			ApproxSize:    "~700 MB",
			Resolve:       stubResolver("debian"),
		},
		{
			ID:            "arch",
			DisplayName:   "Arch Linux",
			Architectures: []string{"amd64"},
			ApproxSize:    "~1 GB",
			Resolve:       stubResolver("arch"),
		},
		{
			ID:            "opensuse-tumbleweed",
			DisplayName:   "openSUSE Tumbleweed",
			Architectures: []string{"amd64"},
			ApproxSize:    "~4 GB",
			Resolve:       stubResolver("opensuse-tumbleweed"),
		},
		{
			ID:            "nixos",
			DisplayName:   "NixOS",
			Architectures: []string{"amd64"},
			ApproxSize:    "~1.5 GB",
			Resolve:       stubResolver("nixos"),
		},
	}
}

func stubResolver(id string) func(string, string) (Release, error) {
	return func(string, string) (Release, error) {
		return Release{}, fmt.Errorf("%s: %w", id, ErrNotImplemented)
	}
}

// LookupDistro finds a distro by ID.
func LookupDistro(id string) (Distro, bool) {
	id = strings.ToLower(strings.TrimSpace(id))
	for _, d := range Catalog() {
		if d.ID == id {
			return d, true
		}
	}
	return Distro{}, false
}

// ListEntry is a row for lab iso list.
type ListEntry struct {
	ID            string
	Version       string
	ApproxSize    string
	Architectures string
}

// ListDistros resolves latest versions for display.
func ListDistros() ([]ListEntry, error) {
	var out []ListEntry
	for _, d := range Catalog() {
		arch := "amd64"
		if len(d.Architectures) > 0 {
			arch = d.Architectures[0]
		}
		ver := "latest"
		if d.Resolve != nil {
			if rel, err := d.Resolve(arch, ""); err == nil {
				ver = rel.Version
			} else if errors.Is(err, ErrNotImplemented) {
				ver = "planned"
			}
		}
		out = append(out, ListEntry{
			ID:            d.ID,
			Version:       ver,
			ApproxSize:    d.ApproxSize,
			Architectures: strings.Join(d.Architectures, ", "),
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

// WriteList prints the catalog table.
func WriteList(w io.Writer, entries []ListEntry) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "DISTRO\tVERSION\tSIZE\tARCHITECTURES")
	for _, e := range entries {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", e.ID, e.Version, e.ApproxSize, e.Architectures)
	}
	return tw.Flush()
}

// DiskType classifies block devices.
type DiskType string

// Disk classification labels.
const (
	DiskSystem DiskType = "SYSTEM"
	DiskUSB    DiskType = "USB"
)

// Disk describes a block device for lab iso disks.
type Disk struct {
	Device string
	Size   string
	Model  string
	Tran   string
	Type   DiskType
}

// ClassifyDisk determines SYSTEM vs USB from lsblk fields.
func ClassifyDisk(rm bool, tran string) DiskType {
	tr := strings.ToLower(strings.TrimSpace(tran))
	if rm || tr == "usb" {
		return DiskUSB
	}
	return DiskSystem
}

// FormatDiskLine returns a display line for a disk.
func FormatDiskLine(d Disk) string {
	tag := string(d.Type)
	switch d.Type {
	case DiskUSB:
		tag += " ✅"
	case DiskSystem:
		tag += " ⚠️"
	}
	return fmt.Sprintf("%-12s %-8s %-24s %-6s %s", d.Device, d.Size, d.Model, d.Tran, tag)
}
