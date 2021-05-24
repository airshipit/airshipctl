/*
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package config_test

import (
	"fmt"
	"testing"

	cmd "opendev.org/airship/airshipctl/cmd/config"
	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/testutil"
)

const (
	fooContext     = "ContextFoo"
	barContext     = "ContextBar"
	bazContext     = "ContextBaz"
	missingContext = "contextMissing"
)

func TestGetContextCmd(t *testing.T) {
	settings := func() (*config.Config, error) {
		return &config.Config{
			Contexts: map[string]*config.Context{
				fooContext: getNamedTestContext(fooContext),
				barContext: getNamedTestContext(barContext),
				bazContext: getNamedTestContext(bazContext),
			},
			CurrentContext: bazContext,
		}, nil
	}

	cmdTests := []*testutil.CmdTest{
		{
			Name:    "get-context",
			CmdLine: fmt.Sprintf("%s", fooContext),
			Cmd:     cmd.NewGetContextCommand(settings),
		},
		{
			Name:    "get-all-contexts",
			CmdLine: fmt.Sprintf("%s", barContext),
			Cmd:     cmd.NewGetContextCommand(settings),
		},
		// This is not implemented yet
		{
			Name:    "get-multiple-contexts",
			CmdLine: fmt.Sprintf("%s %s", fooContext, barContext),
			Cmd:     cmd.NewGetContextCommand(settings),
			Error:   fmt.Errorf("accepts at most 1 arg(s), received 2"),
		},

		{
			Name:    "missing",
			CmdLine: fmt.Sprintf("%s", missingContext),
			Cmd:     cmd.NewGetContextCommand(settings),
			Error: fmt.Errorf("missing configuration: context with name '%s'",
				missingContext),
		},
		{
			Name:    "get-current-context",
			CmdLine: "--current",
			Cmd:     cmd.NewGetContextCommand(settings),
		},
	}

	for _, tt := range cmdTests {
		testutil.RunTest(t, tt)
	}
}

func TestNoContextsGetContextCmd(t *testing.T) {
	settings := func() (*config.Config, error) { return new(config.Config), nil }
	cmdTest := &testutil.CmdTest{
		Name:    "no-contexts",
		CmdLine: "",
		Cmd:     cmd.NewGetContextCommand(settings),
	}
	testutil.RunTest(t, cmdTest)
}

func getNamedTestContext(contextName string) *config.Context {
	newContext := &config.Context{
		Manifest: fmt.Sprintf("Manifest_%s", contextName),
	}
	return newContext
}
