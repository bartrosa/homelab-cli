//go:build linux

package iso

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseDDBytes(t *testing.T) {
	b, ok := parseDDBytes("597688464 bytes (598 MB, 570 MiB) copied, 17.1 s, 34.9 MB/s")
	require.True(t, ok)
	require.Equal(t, int64(597688464), b)

	b, ok = parseDDBytes("skopiowane 578813952 bajtów (579 MB, 552 MiB), 16 s, 35,2 MB/s")
	require.True(t, ok)
	require.Equal(t, int64(578813952), b)

	_, ok = parseDDBytes("579+0 records in")
	require.False(t, ok)
}

func TestReadDDStatus_carriageReturn(t *testing.T) {
	err := readDDStatus(strings.NewReader("skopiowane 1000 bajtów\rskopiowane 2000 bajtów\n"))
	require.NoError(t, err)
}
