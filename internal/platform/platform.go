// Package platform detects OS and package-manager backend for lab.
package platform

import (
	"bytes"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// OS family constants.
const (
	OSDarwin = "darwin"
	OSLinux  = "linux"
)

// Packager backend identifiers.
const (
	PackagerBrew    = "brew"
	PackagerAPT     = "apt"
	PackagerDNF     = "dnf"
	PackagerUnknown = "unknown"
)

// Info describes the current machine.
type Info struct {
	GOOS     string
	Family   string
	Packager string
	// IsSilverblue when rpm-ostree is the primary package interface.
	IsSilverblue bool
}

// Detect inspects the environment and picks a packager.
func Detect() Info {
	goos := runtime.GOOS
	info := Info{GOOS: goos, Family: goos}

	switch goos {
	case OSDarwin:
		if hasCmd("brew") {
			info.Packager = PackagerBrew
		} else {
			info.Packager = PackagerUnknown
		}
	case OSLinux:
		switch {
		case isSilverblue():
			info.IsSilverblue = true
			if hasCmd("rpm-ostree") {
				info.Packager = PackagerDNF // rpm-ostree install wraps dnf-ish
			} else {
				info.Packager = PackagerUnknown
			}
		case hasCmd("apt-get"):
			info.Packager = PackagerAPT
		case hasCmd("dnf"):
			info.Packager = PackagerDNF
		default:
			info.Packager = PackagerUnknown
		}
	default:
		info.Packager = PackagerUnknown
	}

	return info
}

func hasCmd(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func isSilverblue() bool {
	data, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return false
	}
	s := string(data)
	return strings.Contains(s, "VARIANT_ID=sericea") ||
		strings.Contains(s, "VARIANT_ID=silverblue") ||
		strings.Contains(s, "Fedora Silverblue") ||
		strings.Contains(s, "Bluefin")
}

// SupportsMise reports whether mise-based toolchains are practical on this host.
func (i Info) SupportsMise() bool {
	return i.GOOS == OSDarwin || i.GOOS == OSLinux
}

// PackagerLabel returns a human-readable backend name.
func (i Info) PackagerLabel() string {
	if i.IsSilverblue {
		return "rpm-ostree (Silverblue)"
	}
	switch i.Packager {
	case PackagerBrew:
		return "Homebrew"
	case PackagerAPT:
		return "apt"
	case PackagerDNF:
		return "dnf"
	default:
		return "unknown"
	}
}

// ReadOSReleaseKey returns a value from /etc/os-release (linux only).
func ReadOSReleaseKey(key string) (string, bool) {
	data, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return "", false
	}
	prefix := key + "="
	for _, line := range bytes.Split(data, []byte("\n")) {
		if strings.HasPrefix(string(line), prefix) {
			v := strings.TrimPrefix(string(line), prefix)
			return strings.Trim(v, `"`), true
		}
	}
	return "", false
}
