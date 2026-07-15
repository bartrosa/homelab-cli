package stack_test

import (
	"testing"

	"github.com/bartrosa/homelab-cli/internal/stack"
	_ "github.com/bartrosa/homelab-cli/internal/stack/components"
	"github.com/stretchr/testify/require"
)

func TestMergePresets_customOverride(t *testing.T) {
	custom := map[string][]string{"ml": {"python", "go"}}
	merged := stack.MergePresets(custom)
	require.Equal(t, []string{"python", "go"}, merged["ml"])
	require.Contains(t, merged["backend"], "python")
}

func TestResolvePreset_unknown(t *testing.T) {
	_, err := stack.ResolvePreset("nope", nil)
	require.Error(t, err)
}

func TestPresetNames_includesFull(t *testing.T) {
	names := stack.PresetNames(nil)
	require.Contains(t, names, "full")
	require.Contains(t, names, "ml")
}
