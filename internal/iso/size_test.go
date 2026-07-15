package iso

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseSizeBytes(t *testing.T) {
	tests := []struct {
		in   string
		want float64
	}{
		{"58G", 58 * 1024 * 1024 * 1024},
		{"58,6G", 58.6 * 1024 * 1024 * 1024},
		{"1,8T", 1.8 * 1024 * 1024 * 1024 * 1024},
		{"500M", 500 * 1024 * 1024},
		{"326,9M", 326.9 * 1024 * 1024},
		{"4096", 4096},
	}
	for _, tc := range tests {
		got, err := ParseSizeBytes(tc.in)
		require.NoError(t, err, tc.in)
		require.InDelta(t, tc.want, float64(got), 1024*1024, tc.in)
	}
}

func TestValidateDeviceCapacity_polishLocale(t *testing.T) {
	require.NoError(t, ValidateDeviceCapacity("58,6G"))
	require.Error(t, ValidateDeviceCapacity("2,0G"))
}
