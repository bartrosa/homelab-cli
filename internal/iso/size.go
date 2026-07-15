package iso

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseSizeBytes converts lsblk size strings (e.g. 58G, 500M) to bytes.
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
	if u, ok := units[last]; ok {
		numStr := strings.TrimSpace(size[:len(size)-1])
		f, err := strconv.ParseFloat(numStr, 64)
		if err != nil {
			return 0, err
		}
		return int64(f * float64(u)), nil
	}
	return strconv.ParseInt(size, 10, 64)
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
