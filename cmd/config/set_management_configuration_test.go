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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	cmd "opendev.org/airship/airshipctl/cmd/config"
	"opendev.org/airship/airshipctl/pkg/config"
	redfishdell "opendev.org/airship/airshipctl/pkg/remote/redfish/vendors/dell"
	"opendev.org/airship/airshipctl/testutil"
)

func TestConfigSetManagementConfigurationCmd(t *testing.T) {
	conf, cleanup := testutil.InitConfig(t)
	defer cleanup(t)

	settings := func() (*config.Config, error) {
		return conf, nil
	}

	cmdTests := []*testutil.CmdTest{
		{
			Name:    "config-cmd-set-management-config-with-help",
			CmdLine: "--help",
			Cmd:     cmd.NewSetManagementConfigCommand(nil),
		},
		{
			Name:    "config-cmd-set-management-config-no-args",
			CmdLine: "",
			Cmd:     cmd.NewSetManagementConfigCommand(nil),
			Error:   fmt.Errorf("accepts %d arg(s), received %d", 1, 0),
		},
		{
			Name:    "config-cmd-set-management-config-excess-args",
			CmdLine: "arg1 arg2",
			Cmd:     cmd.NewSetManagementConfigCommand(nil),
			Error:   fmt.Errorf("accepts %d arg(s), received %d", 1, 2),
		},
		{
			Name:    "config-cmd-set-management-config-not-found",
			CmdLine: fmt.Sprintf("%s-test", config.AirshipDefaultContext),
			Cmd:     cmd.NewSetManagementConfigCommand(settings),
			Error: config.ErrManagementConfigurationNotFound{
				Name: fmt.Sprintf("%s-test", config.AirshipDefaultContext),
			},
		},
	}

	for _, tt := range cmdTests {
		testutil.RunTest(t, tt)
	}
}

func TestConfigSetManagementConfigurationInsecure(t *testing.T) {
	conf, cleanup := testutil.InitConfig(t)
	defer cleanup(t)

	conf.ManagementConfiguration[config.AirshipDefaultManagementConfiguration] = config.NewManagementConfiguration()

	settings := func() (*config.Config, error) {
		return conf, nil
	}

	require.False(t, conf.ManagementConfiguration[config.AirshipDefaultContext].Insecure)

	testutil.RunTest(t, &testutil.CmdTest{
		Name:    "config-cmd-set-management-config-change-insecure",
		CmdLine: fmt.Sprintf("%s --insecure=true", config.AirshipDefaultContext),
		Cmd:     cmd.NewSetManagementConfigCommand(settings),
	})
	testutil.RunTest(t, &testutil.CmdTest{
		Name:    "config-cmd-set-management-config-no-change",
		CmdLine: fmt.Sprintf("%s --insecure=true", config.AirshipDefaultContext),
		Cmd:     cmd.NewSetManagementConfigCommand(settings),
	})

	assert.True(t, conf.ManagementConfiguration[config.AirshipDefaultContext].Insecure)
}

func TestConfigSetManagementConfigurationType(t *testing.T) {
	conf, cleanup := testutil.InitConfig(t)
	defer cleanup(t)

	conf.ManagementConfiguration[config.AirshipDefaultManagementConfiguration] = config.NewManagementConfiguration()

	settings := func() (*config.Config, error) {
		return conf, nil
	}

	require.NotEqual(t, redfishdell.ClientType,
		conf.ManagementConfiguration[config.AirshipDefaultContext].Type)

	cmdTests := []*testutil.CmdTest{
		{
			Name:    "config-cmd-set-management-config-unknown-type",
			CmdLine: fmt.Sprintf("%s --management-type=foo", config.AirshipDefaultContext),
			Cmd:     cmd.NewSetManagementConfigCommand(settings),
			Error:   config.ErrUnknownManagementType{Type: "foo"},
		},
		{
			Name: "config-cmd-set-management-config-change-type",
			CmdLine: fmt.Sprintf("%s --management-type=%s", config.AirshipDefaultContext,
				redfishdell.ClientType),
			Cmd: cmd.NewSetManagementConfigCommand(settings),
		},
	}

	for _, tt := range cmdTests {
		testutil.RunTest(t, tt)
	}

	assert.Equal(t, redfishdell.ClientType,
		conf.ManagementConfiguration[config.AirshipDefaultContext].Type)
}

func TestConfigSetManagementConfigurationUseProxy(t *testing.T) {
	conf, cleanup := testutil.InitConfig(t)
	defer cleanup(t)

	conf.ManagementConfiguration[config.AirshipDefaultManagementConfiguration] = config.NewManagementConfiguration()

	settings := func() (*config.Config, error) {
		return conf, nil
	}

	require.False(t, conf.ManagementConfiguration[config.AirshipDefaultContext].UseProxy)

	testutil.RunTest(t, &testutil.CmdTest{
		Name:    "config-cmd-set-management-config-change-use-proxy",
		CmdLine: fmt.Sprintf("%s --use-proxy=true", config.AirshipDefaultContext),
		Cmd:     cmd.NewSetManagementConfigCommand(settings),
	})

	assert.True(t, conf.ManagementConfiguration[config.AirshipDefaultContext].UseProxy)
}

func TestConfigSetManagementConfigurationMultipleOptions(t *testing.T) {
	conf, cleanup := testutil.InitConfig(t)
	defer cleanup(t)

	conf.ManagementConfiguration[config.AirshipDefaultManagementConfiguration] = config.NewManagementConfiguration()

	settings := func() (*config.Config, error) {
		return conf, nil
	}

	require.NotEqual(t, redfishdell.ClientType,
		conf.ManagementConfiguration[config.AirshipDefaultContext].Type)
	require.False(t, conf.ManagementConfiguration[config.AirshipDefaultContext].UseProxy)

	testutil.RunTest(t, &testutil.CmdTest{
		Name: "config-cmd-set-management-config-change-type",
		CmdLine: fmt.Sprintf("%s --management-type=%s --use-proxy=true", config.AirshipDefaultContext,
			redfishdell.ClientType),
		Cmd: cmd.NewSetManagementConfigCommand(settings),
	})

	assert.Equal(t, redfishdell.ClientType,
		conf.ManagementConfiguration[config.AirshipDefaultContext].Type)
	assert.True(t, conf.ManagementConfiguration[config.AirshipDefaultContext].UseProxy)
}
