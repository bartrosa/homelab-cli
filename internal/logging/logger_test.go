package logging

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNew_JSONDebugLine(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := slog.New(handlerFor(&buf, "debug", "json"))
	logger.Debug("hello", slog.String("k", "v"))

	line := strings.TrimSpace(buf.String())
	var payload map[string]any
	require.NoError(t, json.Unmarshal([]byte(line), &payload))
	require.Contains(t, payload, "msg")
}
