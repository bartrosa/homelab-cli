// Package gpu detects graphics hardware from lspci output.
package gpu

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/bartrosa/homelab-cli/internal/exec"
)

// Vendor identifies a GPU vendor.
type Vendor string

// Known GPU vendors from PCI vendor IDs.
const (
	VendorNvidia Vendor = "nvidia"
	VendorAmd    Vendor = "amd"
	VendorIntel  Vendor = "intel"
	VendorNone   Vendor = "none"
)

// Info describes one GPU device.
type Info struct {
	Vendor Vendor
	Model  string
	Driver string
}

var bracketRe = regexp.MustCompile(`\[([0-9a-f]{4}):([0-9a-f]{4})\]`)

// Detect parses lspci -nn output.
func Detect(ctx context.Context, runner exec.Runner) ([]Info, error) {
	out, err := runner.RunWithOutput(ctx, "lspci", "-nn")
	if err != nil {
		return nil, fmt.Errorf("lspci: %w", err)
	}
	return ParseLSPCI(out), nil
}

// ParseLSPCI parses lspci lines (testable).
func ParseLSPCI(output string) []Info {
	var out []Info
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		lower := strings.ToLower(line)
		if !strings.Contains(lower, "vga") && !strings.Contains(lower, "3d") && !strings.Contains(lower, "display") {
			continue
		}
		matches := bracketRe.FindAllStringSubmatch(line, -1)
		if len(matches) == 0 {
			continue
		}
		last := matches[len(matches)-1]
		vendorID := strings.ToLower(last[1])
		model := extractModel(line)
		v := VendorNone
		switch vendorID {
		case "10de":
			v = VendorNvidia
		case "1002":
			v = VendorAmd
		case "8086":
			v = VendorIntel
		}
		if v == VendorNone {
			continue
		}
		out = append(out, Info{Vendor: v, Model: model})
	}
	return out
}

func extractModel(line string) string {
	idx := strings.Index(line, ": ")
	if idx < 0 {
		return strings.TrimSpace(line)
	}
	rest := strings.TrimSpace(line[idx+2:])
	if bracket := strings.LastIndex(rest, "["); bracket > 0 {
		rest = strings.TrimSpace(rest[:bracket])
	}
	return rest
}

// DetectNvidia reports whether an NVIDIA GPU was found.
func DetectNvidia(ctx context.Context, runner exec.Runner) (bool, error) {
	gpus, err := Detect(ctx, runner)
	if err != nil {
		return false, err
	}
	for _, g := range gpus {
		if g.Vendor == VendorNvidia {
			return true, nil
		}
	}
	return false, nil
}

// DetectAmd reports whether an AMD GPU was found.
func DetectAmd(ctx context.Context, runner exec.Runner) (bool, error) {
	gpus, err := Detect(ctx, runner)
	if err != nil {
		return false, err
	}
	for _, g := range gpus {
		if g.Vendor == VendorAmd {
			return true, nil
		}
	}
	return false, nil
}
