package completion_test

import (
	"errors"
	"testing"

	"opendev.org/airship/airshipctl/cmd/completion"
	"opendev.org/airship/airshipctl/testutil"
)

func TestCompletion(t *testing.T) {
	cmd := completion.NewCompletionCommand()

	cmdTests := []*testutil.CmdTest{
		{
			Name:    "completion-bash",
			CmdLine: "bash",
			Cmd:     cmd,
		},
		{
			Name:    "completion-zsh",
			CmdLine: "zsh",
			Cmd:     cmd,
		},
		{
			Name:    "completion-unknown-shell",
			CmdLine: "fish",
			Cmd:     cmd,
			Error:   errors.New("unsupported shell type \"fish\""),
		},
	}

	for _, tt := range cmdTests {
		testutil.RunTest(t, tt)
	}
}
