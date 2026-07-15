package iso

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseSizeBytes converts lsblk size strings (e.g. 58G, 58,6G, 500M) to bytes.
func ParseSizeBytes(size string) (int64, error) {
	size = strings.TrimSpace(size)
	if size == "" {
		return 0, fmt.Errorf("empty size")
	}
	units := map[byte]int64{
		'K': 1024,
		'M': 1024 * 1024,
		'G': 1024 * 1024 * 1024,
		'T': 1024 * 1024 * 1024 * 1024,
	}
	last := size[len(size)-1]
	unitKey := strings.ToUpper(string(last))[0]
	if u, ok := units[unitKey]; ok {
		numStr := normalizeSizeNumber(size[:len(size)-1])
		f, err := strconv.ParseFloat(numStr, 64)
		if err != nil {
			return 0, fmt.Errorf("parse size %q: %w", size, err)
		}
		return int64(f * float64(u)), nil
	}
	// Plain integer bytes (no unit suffix).
	plain := normalizeSizeNumber(size)
	if strings.Contains(plain, ".") {
		f, err := strconv.ParseFloat(plain, 64)
		if err != nil {
			return 0, fmt.Errorf("parse size %q: %w", size, err)
		}
		return int64(f), nil
	}
	n, err := strconv.ParseInt(plain, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse size %q: %w", size, err)
	}
	return n, nil
}

// normalizeSizeNumber accepts European decimal commas from locale-aware lsblk output.
func normalizeSizeNumber(s string) string {
	s = strings.TrimSpace(s)
	if !strings.Contains(s, ",") {
		return s
	}
	if strings.Contains(s, ".") {
		// e.g. 1.234,5 → 1234.5
		s = strings.ReplaceAll(s, ".", "")
	}
	return strings.ReplaceAll(s, ",", ".")
}

const minUSBCapacity = 4 * 1024 * 1024 * 1024 // 4 GiB

// ValidateDeviceCapacity rejects devices smaller than min USB size.
func ValidateDeviceCapacity(size string) error {
	bytes, err := ParseSizeBytes(size)
	if err != nil {
		return err
	}
	if bytes < minUSBCapacity {
		return fmt.Errorf("device size %s is below 4 GiB minimum for ISO writes", size)
	}
	return nil
}
