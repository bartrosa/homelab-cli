package pkgmgr

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/bartrosa/homelab-cli/internal/exec"
)

// OSType identifies bootstrap targets.
type OSType string

// Bootstrap OS targets.
const (
	OSUbuntu     OSType = "ubuntu"
	OSSilverblue OSType = "silverblue"
)

// Detect reads /etc/os-release and returns a Manager.
func Detect(runner exec.Runner) (Manager, OSType, error) {
	data, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return nil, "", ErrUnavailable
	}
	osType, mgrName := parseOSRelease(data)
	switch mgrName {
	case "apt":
		return &APT{Runner: runner, Sudo: true}, osType, nil
	case "rpm-ostree":
		return &RPMOstree{Runner: runner}, osType, nil
	default:
		return nil, osType, ErrUnavailable
	}
}

// ParseOSRelease classifies os-release content (exported for tests).
func ParseOSRelease(content []byte) (OSType, string) {
	return parseOSRelease(content)
}

func parseOSRelease(data []byte) (OSType, string) {
	vals := map[string]string{}
	sc := bufio.NewScanner(bytes.NewReader(data))
	for sc.Scan() {
		line := sc.Text()
		if i := strings.IndexByte(line, '='); i > 0 {
			key := line[:i]
			val := strings.Trim(strings.TrimSpace(line[i+1:]), `"`)
			vals[key] = val
		}
	}

	id := strings.ToLower(vals["ID"])
	variant := strings.ToLower(vals["VARIANT_ID"])
	name := strings.ToLower(vals["NAME"])

	if variant == "silverblue" || variant == "sericea" || strings.Contains(name, "silverblue") {
		return OSSilverblue, "rpm-ostree"
	}
	if id == "ubuntu" {
		return OSUbuntu, "apt"
	}
	if id == "debian" {
		return OSType("debian"), "apt"
	}
	if id == "fedora" {
		return OSSilverblue, "rpm-ostree"
	}
	return OSType(id), ""
}

// DetectTarget resolves --target auto|ubuntu|silverblue.
func DetectTarget(target string, runner exec.Runner) (Manager, OSType, error) {
	switch strings.ToLower(strings.TrimSpace(target)) {
	case "", "auto":
		return Detect(runner)
	case "ubuntu":
		return &APT{Runner: runner, Sudo: true}, OSUbuntu, nil
	case "silverblue":
		return &RPMOstree{Runner: runner}, OSSilverblue, nil
	default:
		return nil, "", fmt.Errorf("unsupported target %q", target)
	}
}
