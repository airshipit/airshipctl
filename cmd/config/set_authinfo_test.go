/*
Copyright 2017 The Kubernetes Authors.

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

package config

import (
	"fmt"
	"strings"
	"testing"

	kubeconfig "k8s.io/client-go/tools/clientcmd/api"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/environment"
	"opendev.org/airship/airshipctl/testutil"
)

const (
	testUsername     = "admin@kubernetes"
	testPassword     = "adminPassword"
	newUserName      = "dummy"
	existingUserName = "def-user"
	pwdDelta         = "_changed"
)

type setAuthInfoTest struct {
	inputConfig  *config.Config
	cmdTest      *testutil.CmdTest
	userName     string
	userPassword string
}

func TestConfigSetAuthInfo(t *testing.T) {
	cmdTests := []*testutil.CmdTest{
		{
			Name:    "config-cmd-set-authinfo-with-help",
			CmdLine: "--help",
			Cmd:     NewCmdConfigSetAuthInfo(nil),
		},
		{
			Name:    "config-cmd-set-authinfo-too-many-args",
			CmdLine: "arg1 arg2",
			Cmd:     NewCmdConfigSetAuthInfo(nil),
			Error:   fmt.Errorf("accepts %d arg(s), received %d", 1, 2),
		},
		{
			Name:    "config-cmd-set-authinfo-too-few-args",
			CmdLine: "",
			Cmd:     NewCmdConfigSetAuthInfo(nil),
			Error:   fmt.Errorf("accepts %d arg(s), received %d", 1, 0),
		},
	}

	for _, tt := range cmdTests {
		testutil.RunTest(t, tt)
	}
}

// initInputConfig creates an input config
// Each of these config objects are associated with real files. Those files can be
// cleaned up by calling cleanup
func initInputConfig(t *testing.T) (given *config.Config, cleanup func(*testing.T)) {
	given, givenCleanup := config.InitConfig(t)
	kubeAuthInfo := kubeconfig.NewAuthInfo()
	kubeAuthInfo.Username = testUsername
	kubeAuthInfo.Password = testPassword
	given.KubeConfig().AuthInfos[existingUserName] = kubeAuthInfo
	given.AuthInfos[existingUserName].SetKubeAuthInfo(kubeAuthInfo)

	return given, givenCleanup
}

func (test setAuthInfoTest) run(t *testing.T) {
	settings := &environment.AirshipCTLSettings{}
	settings.SetConfig(test.inputConfig)
	test.cmdTest.Cmd = NewCmdConfigSetAuthInfo(settings)
	testutil.RunTest(t, test.cmdTest)

	afterRunConf := settings.Config()
	// Find the AuthInfo Created or Modified
	afterRunAuthInfo, err := afterRunConf.GetAuthInfo(test.userName)
	require.NoError(t, err)
	require.NotNil(t, afterRunAuthInfo)

	afterKauthinfo := afterRunAuthInfo.KubeAuthInfo()
	require.NotNil(t, afterKauthinfo)

	assert.EqualValues(t, afterKauthinfo.Username, testUsername)
	assert.EqualValues(t, afterKauthinfo.Password, test.userPassword)
}

func TestSetAuthInfo(t *testing.T) {
	given, cleanup := config.InitConfig(t)
	defer cleanup(t)

	input, cleanupInput := initInputConfig(t)
	defer cleanupInput(t)

	tests := []struct {
		testName     string
		flags        []string
		userName     string
		userPassword string
		inputConfig  *config.Config
	}{
		{
			testName: "set-auth-info",
			flags: []string{
				"--" + config.FlagUsername + "=" + testUsername,
				"--" + config.FlagPassword + "=" + testPassword,
			},
			userName:     newUserName,
			userPassword: testPassword,
			inputConfig:  given,
		},
		{
			testName: "modify-auth-info",
			flags: []string{
				"--" + config.FlagPassword + "=" + testPassword + pwdDelta,
			},
			userName:     existingUserName,
			userPassword: testPassword + pwdDelta,
			inputConfig:  input,
		},
	}
	for _, tt := range tests {
		tt := tt
		cmd := &testutil.CmdTest{
			Name:    tt.testName,
			CmdLine: fmt.Sprintf("%s %s", tt.userName, strings.Join(tt.flags, " ")),
		}
		test := setAuthInfoTest{
			inputConfig:  tt.inputConfig,
			cmdTest:      cmd,
			userName:     tt.userName,
			userPassword: tt.userPassword,
		}
		test.run(t)
	}
}
