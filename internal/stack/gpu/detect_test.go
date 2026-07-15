package gpu_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bartrosa/homelab-cli/internal/stack/gpu"
	"github.com/stretchr/testify/require"
)

func readFixture(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	require.NoError(t, err)
	return string(data)
}

func TestParseLSPCI_nvidia(t *testing.T) {
	gpus := gpu.ParseLSPCI(readFixture(t, "lspci_nvidia.txt"))
	require.Len(t, gpus, 2)
	require.Equal(t, gpu.VendorNvidia, gpus[1].Vendor)
	require.Contains(t, gpus[1].Model, "RTX 4090")
}

func TestParseLSPCI_amd(t *testing.T) {
	gpus := gpu.ParseLSPCI(readFixture(t, "lspci_amd.txt"))
	require.Len(t, gpus, 1)
	require.Equal(t, gpu.VendorAmd, gpus[0].Vendor)
}

func TestParseLSPCI_none(t *testing.T) {
	require.Empty(t, gpu.ParseLSPCI(readFixture(t, "lspci_none.txt")))
}
