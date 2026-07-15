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

func TestParseLSBLKJSON_withoutTypeColumn(t *testing.T) {
	const noType = `{
  "blockdevices": [
    {"name":"nvme0n1","size":"1T","model":"WD_BLACK","tran":"nvme","rm":false},
    {"name":"sdb","size":"58G","model":"SanDisk","tran":"usb","rm":true},
    {"name":"loop0","size":"4K","model":null,"tran":null,"rm":false}
  ]
}`
	disks, err := iso.ParseLSBLKJSON(noType)
	require.NoError(t, err)
	require.Len(t, disks, 2)
	require.Equal(t, "/dev/nvme0n1", disks[0].Device)
	require.Equal(t, "/dev/sdb", disks[1].Device)
	require.Equal(t, iso.DiskUSB, disks[1].Type)
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
