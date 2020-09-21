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
	"opendev.org/airship/airshipctl/testutil"
)

func TestGetEncryptionConfigCmd(t *testing.T) {
	settings := func() (*config.Config, error) {
		return &config.Config{
			EncryptionConfigs: map[string]*config.EncryptionConfig{
				config.AirshipDefaultContext: testutil.DummyEncryptionConfig(),
			},
		}, nil
	}

	emptySettings := func() (*config.Config, error) {
		return &config.Config{}, nil
	}

	cmdTests := []*testutil.CmdTest{
		{
			Name:    "get-encryption-config-with-help",
			CmdLine: "--help",
			Cmd:     cmd.NewGetEncryptionConfigCommand(nil),
		},
		{
			Name:    "get-encryption-config-not-found",
			CmdLine: "foo",
			Cmd:     cmd.NewGetEncryptionConfigCommand(emptySettings),
			Error:   config.ErrEncryptionConfigurationNotFound{Name: "foo"},
		},
		{
			Name:    "get-encryption-config-all",
			CmdLine: "",
			Cmd:     cmd.NewGetEncryptionConfigCommand(settings),
		},
		{
			Name:    "get-empty-encryption-config",
			CmdLine: config.AirshipDefaultContext,
			Cmd:     cmd.NewGetEncryptionConfigCommand(settings),
		},
	}

	for _, tt := range cmdTests {
		testutil.RunTest(t, tt)
	}
}
