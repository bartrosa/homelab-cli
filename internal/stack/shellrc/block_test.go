package shellrc_test

import (
	"strings"
	"testing"

	"github.com/bartrosa/homelab-cli/internal/stack/shellrc"
	"github.com/stretchr/testify/require"
)

func TestGenerateBlock_bashIdempotent(t *testing.T) {
	entries := []shellrc.Entry{
		{Shell: "all", Marker: "mise", Content: `eval "$(mise activate bash)"`},
		{Shell: "all", Marker: "path", Content: `export PATH="$HOME/.local/bin:$PATH"`},
	}
	block := shellrc.GenerateBlock(shellrc.Bash, entries)
	require.Contains(t, block, "BEGIN homelab-cli managed block")
	require.Contains(t, block, "mise")
	require.Contains(t, block, "END homelab-cli managed block")

	content := "# user config\n" + block + "\n# tail\n"
	replaced := replaceBlockForTest(content, block)
	require.Equal(t, content, replaced)
}

func TestGenerateBlock_fishSyntax(t *testing.T) {
	entries := []shellrc.Entry{
		{Shell: "all", Marker: "path", Content: `export PATH="$HOME/.local/bin:$PATH"`},
	}
	block := shellrc.GenerateBlock(shellrc.Fish, entries)
	require.Contains(t, block, "set -gx PATH")
	require.NotContains(t, block, "export PATH")
}

func replaceBlockForTest(content, block string) string {
	start := strings.Index(content, "# BEGIN homelab-cli managed block")
	if start < 0 {
		return content + "\n" + block + "\n"
	}
	endRel := strings.Index(content[start:], "# END homelab-cli managed block")
	end := start + endRel + len("# END homelab-cli managed block")
	return content[:start] + block + content[end:]
}
