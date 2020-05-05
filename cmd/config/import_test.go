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
	"errors"
	"testing"

	cmd "opendev.org/airship/airshipctl/cmd/config"
	"opendev.org/airship/airshipctl/pkg/environment"
	"opendev.org/airship/airshipctl/testutil"
)

func TestConfigImport(t *testing.T) {
	settings := &environment.AirshipCTLSettings{Config: testutil.DummyConfig()}

	cmdTests := []*testutil.CmdTest{
		{
			Name:    "config-import-with-help",
			CmdLine: "--help",
			Cmd:     cmd.NewImportCommand(nil),
		},
		{
			Name:    "config-import-no-args",
			CmdLine: "",
			Cmd:     cmd.NewImportCommand(settings),
			Error:   errors.New("accepts 1 arg(s), received 0"),
		},
		{
			Name:    "config-import-file-does-not-exist",
			CmdLine: "foo",
			Cmd:     cmd.NewImportCommand(settings),
			Error:   errors.New("stat foo: no such file or directory"),
		},
	}

	for _, tt := range cmdTests {
		testutil.RunTest(t, tt)
	}
}
