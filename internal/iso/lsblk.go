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
	Model      string      `json:"model"`
	Tran       string      `json:"tran"`
	RM         bool        `json:"rm"`
	Type       string      `json:"type"`
	Vendor     string      `json:"vendor"`
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
		if dev.Type != "disk" {
			continue
		}
		model := strings.TrimSpace(dev.Model)
		if model == "" {
			model = strings.TrimSpace(dev.Vendor)
		}
		diskType := ClassifyDisk(dev.RM, dev.Tran)
		disks = append(disks, Disk{
			Device: "/dev/" + dev.Name,
			Size:   dev.Size,
			Model:  model,
			Tran:   dev.Tran,
			Type:   diskType,
		})
	}
	return disks, nil
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
