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
	currentContextFlag = "--" + config.FlagCurrentContext

	fooContext     = "ContextFoo"
	barContext     = "ContextBar"
	bazContext     = "ContextBaz"
	missingContext = "contextMissing"
)

func TestGetContextCmd(t *testing.T) {
	conf := &config.Config{
		Contexts: map[string]*config.Context{
			fooContext: getNamedTestContext(fooContext),
			barContext: getNamedTestContext(barContext),
			bazContext: getNamedTestContext(bazContext),
		},
		CurrentContext: bazContext,
	}

	settings := &environment.AirshipCTLSettings{}
	settings.SetConfig(conf)

	cmdTests := []*testutil.CmdTest{
		{
			Name:    "get-context",
			CmdLine: fmt.Sprintf("%s", fooContext),
			Cmd:     cmd.NewCmdConfigGetContext(settings),
		},
		{
			Name:    "get-all-contexts",
			CmdLine: fmt.Sprintf("%s %s", fooContext, barContext),
			Cmd:     cmd.NewCmdConfigGetContext(settings),
		},
		// This is not implemented yet
		{
			Name:    "get-multiple-contexts",
			CmdLine: fmt.Sprintf("%s %s", fooContext, barContext),
			Cmd:     cmd.NewCmdConfigGetContext(settings),
		},

		{
			Name:    "missing",
			CmdLine: fmt.Sprintf("%s", missingContext),
			Cmd:     cmd.NewCmdConfigGetContext(settings),
			Error: fmt.Errorf(`Context %s information was not
found in the configuration.`, missingContext),
		},
		{
			Name:    "get-current-context",
			CmdLine: fmt.Sprintf("%s", currentContextFlag),
			Cmd:     cmd.NewCmdConfigGetContext(settings),
		},
	}

	for _, tt := range cmdTests {
		testutil.RunTest(t, tt)
	}
}

func TestNoContextsGetContextCmd(t *testing.T) {
	settings := &environment.AirshipCTLSettings{}
	settings.SetConfig(&config.Config{})
	cmdTest := &testutil.CmdTest{
		Name:    "no-contexts",
		CmdLine: "",
		Cmd:     cmd.NewCmdConfigGetContext(settings),
	}
	testutil.RunTest(t, cmdTest)
}

func getNamedTestContext(contextName string) *config.Context {
	kContext := &kubeconfig.Context{
		Namespace: "dummy_namespace",
		AuthInfo:  "dummy_user",
		Cluster:   fmt.Sprintf("dummycluster_%s", config.Ephemeral),
	}

	newContext := &config.Context{
		NameInKubeconf: fmt.Sprintf("%s_%s", contextName, config.Ephemeral),
		Manifest:       fmt.Sprintf("Manifest_%s", contextName),
	}
	newContext.SetKubeContext(kContext)

	return newContext
}
