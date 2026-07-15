package iso_test

import (
	"testing"

	"github.com/bartrosa/homelab-cli/internal/iso"
	"github.com/stretchr/testify/require"
)

const sampleLSBLK = `{
  "blockdevices": [
    {"name":"sda","size":"500G","model":"Samsung SSD 860","tran":"sata","rm":false,"type":"disk","vendor":""},
    {"name":"nvme0n1","size":"1.0T","model":"WD_BLACK SN770","tran":"nvme","rm":false,"type":"disk","vendor":""},
    {"name":"sdb","size":"58G","model":"SanDisk Ultra USB 3.0","tran":"usb","rm":true,"type":"disk","vendor":""}
  ]
}`

func TestParseLSBLKJSON_classifiesUSB(t *testing.T) {
	disks, err := iso.ParseLSBLKJSON(sampleLSBLK)
	require.NoError(t, err)
	require.Len(t, disks, 3)

	require.Equal(t, iso.DiskSystem, disks[0].Type)
	require.Equal(t, "/dev/sda", disks[0].Device)

	require.Equal(t, iso.DiskSystem, disks[1].Type)

	require.Equal(t, iso.DiskUSB, disks[2].Type)
	require.Equal(t, "usb", disks[2].Tran)
}

func TestClassifyDisk(t *testing.T) {
	require.Equal(t, iso.DiskUSB, iso.ClassifyDisk(true, "sata"))
	require.Equal(t, iso.DiskUSB, iso.ClassifyDisk(false, "usb"))
	require.Equal(t, iso.DiskSystem, iso.ClassifyDisk(false, "nvme"))
}

func TestLookupDistro(t *testing.T) {
	d, ok := iso.LookupDistro("ubuntu-desktop")
	require.True(t, ok)
	require.Equal(t, "ubuntu-desktop", d.ID)
}

func TestParseSHA256SUMS(t *testing.T) {
	content := "abc123  ubuntu-24.04.3-desktop-amd64.iso\n"
	hash, ok := iso.ParseSHA256SUMS(content, "ubuntu-24.04.3-desktop-amd64.iso")
	require.True(t, ok)
	require.Equal(t, "abc123", hash)
}
