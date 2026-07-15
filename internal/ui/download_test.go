package ui

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestFormatBytes(t *testing.T) {
	require.Equal(t, "512 B", FormatBytes(512))
	require.Equal(t, "1.0 KiB", FormatBytes(1024))
	require.Equal(t, "6.2 GiB", FormatBytes(6655619072))
}

func TestFormatETA(t *testing.T) {
	require.Equal(t, "45s", FormatETA(45*time.Second))
	require.Equal(t, "2m05s", FormatETA(125*time.Second))
	require.Equal(t, "1h05m", FormatETA(3900*time.Second))
}

func TestProgressBar(t *testing.T) {
	bar := progressBar(50, 20, true)
	require.True(t, strings.HasPrefix(bar, "["))
	require.Contains(t, bar, "=")
	require.Contains(t, bar, ">")
}

func TestDownloadReporterPlain(t *testing.T) {
	var buf bytes.Buffer
	rep := NewDownloadReporter(&buf, true)
	rep.SetTotal(1000)
	w := rep.Writer()
	_, err := w.Write(make([]byte, 250))
	require.NoError(t, err)
	rep.Finish()
	require.Contains(t, buf.String(), "250")
}
