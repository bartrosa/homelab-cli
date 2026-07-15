package prompt_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/bartrosa/homelab-cli/internal/prompt"
	"github.com/stretchr/testify/require"
)

func TestAskString_default(t *testing.T) {
	in := strings.NewReader("\n")
	out := &bytes.Buffer{}
	p := &prompt.StdinPrompter{In: in, Out: out}
	val, err := p.AskString("Port", "5432")
	require.NoError(t, err)
	require.Equal(t, "5432", val)
}

func TestValidateSchema_required(t *testing.T) {
	schema := prompt.Schema{Fields: []prompt.Field{{Name: "port", Required: true, Type: prompt.FieldTypeInt}}}
	err := prompt.ValidateSchema(schema, map[string]any{})
	require.Error(t, err)
}
