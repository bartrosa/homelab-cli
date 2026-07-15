package iso

import "strings"

// BlockDevicePath ensures a block device path has a /dev/ prefix.
func BlockDevicePath(device string) string {
	device = strings.TrimSpace(device)
	if device == "" {
		return ""
	}
	if !strings.HasPrefix(device, "/dev/") {
		return "/dev/" + strings.TrimPrefix(device, "/dev/")
	}
	return device
}
