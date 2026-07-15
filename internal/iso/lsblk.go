package iso

import (
	"encoding/json"
	"fmt"
	"strings"
)

type lsblkRoot struct {
	BlockDevices []lsblkNode `json:"blockdevices"`
}

type lsblkNode struct {
	Name       string      `json:"name"`
	Size       string      `json:"size"`
	Model      *string     `json:"model"`
	Tran       *string     `json:"tran"`
	RM         bool        `json:"rm"`
	Type       string      `json:"type"`
	Vendor     *string     `json:"vendor"`
	Mountpoint *string     `json:"mountpoint"`
	Children   []lsblkNode `json:"children"`
}

// ParseLSBLKJSON parses lsblk JSON output into Disk entries.
func ParseLSBLKJSON(raw string) ([]Disk, error) {
	var payload lsblkRoot
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return nil, fmt.Errorf("parse lsblk JSON: %w", err)
	}
	var disks []Disk
	for _, dev := range payload.BlockDevices {
		if !isBlockDisk(dev) {
			continue
		}
		model := strVal(dev.Model)
		if model == "" {
			model = strVal(dev.Vendor)
		}
		tran := strVal(dev.Tran)
		diskType := ClassifyDisk(dev.RM, tran)
		disks = append(disks, Disk{
			Device: "/dev/" + dev.Name,
			Size:   dev.Size,
			Model:  model,
			Tran:   tran,
			Type:   diskType,
		})
	}
	return disks, nil
}

func isBlockDisk(dev lsblkNode) bool {
	if dev.Type == "disk" {
		return true
	}
	// Older lsblk without TYPE column: skip known non-disk prefixes.
	if dev.Type != "" {
		return false
	}
	name := dev.Name
	if strings.HasPrefix(name, "loop") || strings.HasPrefix(name, "dm-") || strings.HasPrefix(name, "ram") {
		return false
	}
	return name != ""
}

func strVal(s *string) string {
	if s == nil {
		return ""
	}
	return strings.TrimSpace(*s)
}

func collectMountpoints(dev lsblkNode, out *[]string) {
	if dev.Mountpoint != nil && *dev.Mountpoint != "" {
		*out = append(*out, *dev.Mountpoint)
	}
	for _, c := range dev.Children {
		collectMountpoints(c, out)
	}
}

func collectPartNames(dev lsblkNode, out *[]string) {
	*out = append(*out, dev.Name)
	for _, c := range dev.Children {
		collectPartNames(c, out)
	}
}

func collectUmountTargets(dev lsblkNode, out *[]string) {
	if dev.Mountpoint != nil {
		mp := strings.TrimSpace(*dev.Mountpoint)
		if mp != "" {
			*out = append(*out, mp)
			*out = append(*out, "/dev/"+dev.Name)
		}
	}
	for _, c := range dev.Children {
		collectUmountTargets(c, out)
	}
}
