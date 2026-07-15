//go:build linux

package iso

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const sectorSize = 512

// blockWriteBytes returns lifetime bytes written to a block device (from sysfs).
func blockWriteBytes(device string) (int64, error) {
	name := strings.TrimPrefix(BlockDevicePath(device), "/dev/")
	data, err := os.ReadFile(filepath.Join("/sys/block", name, "stat"))
	if err != nil {
		return 0, err
	}
	fields := strings.Fields(string(data))
	if len(fields) < 7 {
		return 0, fmt.Errorf("short /sys/block/%s/stat", name)
	}
	sectors, err := strconv.ParseUint(fields[6], 10, 64)
	if err != nil {
		return 0, err
	}
	return int64(sectors) * sectorSize, nil
}
