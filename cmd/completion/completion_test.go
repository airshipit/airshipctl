package completion_test

import (
	"errors"
	"fmt"
	"testing"

	"opendev.org/airship/airshipctl/cmd/completion"
	"opendev.org/airship/airshipctl/testutil"
)

func TestCompletion(t *testing.T) {
	cmdTests := []*testutil.CmdTest{
		{
			Name:    "completion-bash",
			CmdLine: "bash",
			Cmd:     completion.NewCompletionCommand(),
		},
		{
			Name:    "completion-zsh",
			CmdLine: "zsh",
			Cmd:     completion.NewCompletionCommand(),
		},
		{
			Name:    "completion-unknown-shell",
			CmdLine: "fish",
			Cmd:     completion.NewCompletionCommand(),
			Error:   errors.New("unsupported shell type \"fish\""),
		},
		{
			Name:    "completion-cmd-too-many-args",
			CmdLine: "arg1 arg2",
			Cmd:     completion.NewCompletionCommand(),
			Error:   fmt.Errorf("accepts %d arg(s), received %d", 1, 2),
		},
		{
			Name:    "completion-cmd-too-few-args",
			CmdLine: "",
			Cmd:     completion.NewCompletionCommand(),
			Error:   fmt.Errorf("accepts %d arg(s), received %d", 1, 0),
		},
	}

	for _, tt := range cmdTests {
		testutil.RunTest(t, tt)
	}
}
