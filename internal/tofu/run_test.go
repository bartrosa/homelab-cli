package tofu_test

import (
	"context"
	"testing"

	"github.com/bartrosa/homelab-cli/internal/tofu"
	"github.com/stretchr/testify/require"
)

func TestCommandLine_chdir(t *testing.T) {
	args := tofu.CommandLine(tofu.Stack{Dir: "/stacks/rb5009-core"}, "plan")
	require.Equal(t, []string{"-chdir=/stacks/rb5009-core", "plan"}, args)
}

func TestPlan_buildsChdirArgs(t *testing.T) {
	rec := &recordingRunner{}
	err := tofu.Plan(context.Background(), rec, tofu.Stack{Dir: "/tmp/stack"}, tofu.RunOpts{})
	require.NoError(t, err)
	require.Equal(t, []string{"tofu -chdir=/tmp/stack plan"}, rec.calls)
}

func TestApply_withoutYes_noAutoApprove(t *testing.T) {
	rec := &recordingRunner{}
	err := tofu.Apply(context.Background(), rec, tofu.Stack{Dir: "/tmp/stack"}, tofu.RunOpts{AutoApprove: false})
	require.NoError(t, err)
	require.Equal(t, []string{"tofu -chdir=/tmp/stack apply"}, rec.calls)
	require.NotContains(t, rec.calls[0], "-auto-approve")
}

func TestApply_withYes_addsAutoApprove(t *testing.T) {
	rec := &recordingRunner{}
	err := tofu.Apply(context.Background(), rec, tofu.Stack{Dir: "/tmp/stack"}, tofu.RunOpts{AutoApprove: true})
	require.NoError(t, err)
	require.Equal(t, []string{"tofu -chdir=/tmp/stack apply -auto-approve"}, rec.calls)
}

func TestApply_dryRun_doesNotExecute(t *testing.T) {
	rec := &recordingRunner{}
	err := tofu.Apply(context.Background(), rec, tofu.Stack{Dir: "/tmp/stack"}, tofu.RunOpts{
		AutoApprove: true,
		DryRun:      true,
	})
	require.NoError(t, err)
	require.Empty(t, rec.calls)
}

func TestInit_Validate_Fmt(t *testing.T) {
	rec := &recordingRunner{}
	stack := tofu.Stack{Dir: "hq-infra/tofu/stacks/rb5009-core"}
	require.NoError(t, tofu.Init(context.Background(), rec, stack, tofu.RunOpts{}))
	require.NoError(t, tofu.Validate(context.Background(), rec, stack, tofu.RunOpts{}))
	require.NoError(t, tofu.Fmt(context.Background(), rec, stack, tofu.RunOpts{}))
	require.Equal(t, []string{
		"tofu -chdir=hq-infra/tofu/stacks/rb5009-core init",
		"tofu -chdir=hq-infra/tofu/stacks/rb5009-core validate",
		"tofu -chdir=hq-infra/tofu/stacks/rb5009-core fmt",
	}, rec.calls)
}

func TestRun_emptyDir(t *testing.T) {
	rec := &recordingRunner{}
	err := tofu.Plan(context.Background(), rec, tofu.Stack{Dir: "  "}, tofu.RunOpts{})
	require.Error(t, err)
	require.Empty(t, rec.calls)
}
