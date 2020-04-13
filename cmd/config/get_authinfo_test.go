/*
Copyright 2014 The Kubernetes Authors.

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

	kubeconfig "k8s.io/client-go/tools/clientcmd/api"

	cmd "opendev.org/airship/airshipctl/cmd/config"
	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/environment"
	"opendev.org/airship/airshipctl/testutil"
)

const (
	fooAuthInfo     = "AuthInfoFoo"
	barAuthInfo     = "AuthInfoBar"
	bazAuthInfo     = "AuthInfoBaz"
	missingAuthInfo = "authinfoMissing"
)

func TestGetAuthInfoCmd(t *testing.T) {
	settings := &environment.AirshipCTLSettings{
		Config: &config.Config{
			AuthInfos: map[string]*config.AuthInfo{
				fooAuthInfo: getTestAuthInfo(),
				barAuthInfo: getTestAuthInfo(),
				bazAuthInfo: getTestAuthInfo(),
			},
		},
	}

	cmdTests := []*testutil.CmdTest{
		{
			Name:    "get-credentials",
			CmdLine: fmt.Sprintf("%s", fooAuthInfo),
			Cmd:     cmd.NewCmdConfigGetAuthInfo(settings),
		},
		{
			Name:    "get-all-credentials",
			CmdLine: fmt.Sprintf("%s %s", fooAuthInfo, barAuthInfo),
			Cmd:     cmd.NewCmdConfigGetAuthInfo(settings),
		},
		// This is not implemented yet
		{
			Name:    "get-multiple-credentials",
			CmdLine: fmt.Sprintf("%s %s", fooAuthInfo, barAuthInfo),
			Cmd:     cmd.NewCmdConfigGetAuthInfo(settings),
		},

		{
			Name:    "missing",
			CmdLine: fmt.Sprintf("%s", missingAuthInfo),
			Cmd:     cmd.NewCmdConfigGetAuthInfo(settings),
			Error: fmt.Errorf("user %s information was not "+
				"found in the configuration.", missingAuthInfo),
		},
	}

	for _, tt := range cmdTests {
		testutil.RunTest(t, tt)
	}
}

func TestNoAuthInfosGetAuthInfoCmd(t *testing.T) {
	settings := &environment.AirshipCTLSettings{Config: new(config.Config)}
	cmdTest := &testutil.CmdTest{
		Name:    "no-credentials",
		CmdLine: "",
		Cmd:     cmd.NewCmdConfigGetAuthInfo(settings),
	}
	testutil.RunTest(t, cmdTest)
}

func getTestAuthInfo() *config.AuthInfo {
	kAuthInfo := &kubeconfig.AuthInfo{
		Username:          "dummy_user",
		Password:          "dummy_password",
		ClientCertificate: "dummy_certificate",
		ClientKey:         "dummy_key",
		Token:             "dummy_token",
	}

	newAuthInfo := &config.AuthInfo{}
	newAuthInfo.SetKubeAuthInfo(kAuthInfo)

	return newAuthInfo
}
