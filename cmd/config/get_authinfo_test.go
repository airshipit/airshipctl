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

	"github.com/stretchr/testify/assert"

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
				fooAuthInfo: getTestAuthInfo(fooAuthInfo),
			},
		},
	}
	settingsWithMultipleAuth := &environment.AirshipCTLSettings{
		Config: &config.Config{
			AuthInfos: map[string]*config.AuthInfo{
				barAuthInfo: getTestAuthInfo(barAuthInfo),
				bazAuthInfo: getTestAuthInfo(bazAuthInfo),
			},
		},
	}

	cmdTests := []*testutil.CmdTest{
		{
			Name:    "get-specific-credentials",
			CmdLine: fooAuthInfo,
			Cmd:     cmd.NewGetAuthInfoCommand(settings),
		},
		{
			Name:    "get-all-credentials",
			CmdLine: "",
			Cmd:     cmd.NewGetAuthInfoCommand(settingsWithMultipleAuth),
		},
		{
			Name:    "missing",
			CmdLine: missingAuthInfo,
			Cmd:     cmd.NewGetAuthInfoCommand(settings),
			Error: fmt.Errorf("user %s information was not "+
				"found in the configuration", missingAuthInfo),
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
		Cmd:     cmd.NewGetAuthInfoCommand(settings),
	}
	testutil.RunTest(t, cmdTest)
}

func TestDecodeAuthInfo(t *testing.T) {
	_, err := config.DecodeAuthInfo(&kubeconfig.AuthInfo{Password: "dummy_password"})
	assert.Error(t, err, config.ErrDecodingCredentials{Given: "dummy_password"})

	_, err = config.DecodeAuthInfo(&kubeconfig.AuthInfo{ClientCertificate: "dummy_certificate"})
	assert.Error(t, err, config.ErrDecodingCredentials{Given: "dummy_certificate"})

	_, err = config.DecodeAuthInfo(&kubeconfig.AuthInfo{ClientKey: "dummy_key"})
	assert.Error(t, err, config.ErrDecodingCredentials{Given: "dummy_key"})

	_, err = config.DecodeAuthInfo(&kubeconfig.AuthInfo{Token: "dummy_token"})
	assert.Error(t, err, config.ErrDecodingCredentials{Given: "dummy_token"})
}

func getTestAuthInfo(authName string) *config.AuthInfo {
	kAuthInfo := &kubeconfig.AuthInfo{
		Username:          authName + "_user",
		Password:          authName + "_password",
		ClientCertificate: authName + "_certificate",
		ClientKey:         authName + "_key",
		Token:             authName + "_token",
	}

	newAuthInfo := &config.AuthInfo{}
	encodedKAuthInfo := config.EncodeAuthInfo(kAuthInfo)
	newAuthInfo.SetKubeAuthInfo(encodedKAuthInfo)
	return newAuthInfo
}
