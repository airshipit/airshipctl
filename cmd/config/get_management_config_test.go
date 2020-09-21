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
	"testing"

	cmd "opendev.org/airship/airshipctl/cmd/config"
	"opendev.org/airship/airshipctl/pkg/config"
	redfishdell "opendev.org/airship/airshipctl/pkg/remote/redfish/vendors/dell"
	"opendev.org/airship/airshipctl/testutil"
)

func TestGetManagementConfigCmd(t *testing.T) {
	settings := func() (*config.Config, error) {
		return &config.Config{
			ManagementConfiguration: map[string]*config.ManagementConfiguration{
				config.AirshipDefaultContext: testutil.DummyManagementConfiguration(),
				"test": {
					Type: redfishdell.ClientType,
				},
			},
		}, nil
	}
	emptySettings := func() (*config.Config, error) {
		return &config.Config{}, nil
	}

	cmdTests := []*testutil.CmdTest{
		{
			Name:    "get-management-config-with-help",
			CmdLine: "--help",
			Cmd:     cmd.NewGetManagementConfigCommand(nil),
		},
		{
			Name:    "get-management-config-not-found",
			CmdLine: "foo",
			Cmd:     cmd.NewGetManagementConfigCommand(settings),
			Error:   config.ErrManagementConfigurationNotFound{Name: "foo"},
		},
		{
			Name:    "get-management-config-all",
			CmdLine: "",
			Cmd:     cmd.NewGetManagementConfigCommand(settings),
		},
		{
			Name:    "get-management-config-default",
			CmdLine: config.AirshipDefaultContext,
			Cmd:     cmd.NewGetManagementConfigCommand(settings),
		},
		{
			Name:    "get-empty-management-config",
			CmdLine: "",
			Cmd:     cmd.NewGetManagementConfigCommand(emptySettings),
		},
	}

	for _, tt := range cmdTests {
		testutil.RunTest(t, tt)
	}
}
