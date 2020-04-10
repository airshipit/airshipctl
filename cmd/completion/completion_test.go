/*
 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     https://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

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
