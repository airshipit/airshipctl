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
	"bytes"
	"fmt"
	"testing"

	kubeconfig "k8s.io/client-go/tools/clientcmd/api"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/environment"
	"opendev.org/airship/airshipctl/testutil"
)

const (
	testUser = "admin@kubernetes"
)

type setContextTest struct {
	description    string
	givenConfig    *config.Config
	args           []string
	flags          []string
	expectedOutput string
	expectedConfig *config.Config
}

func TestConfigSetContext(t *testing.T) {
	cmdTests := []*testutil.CmdTest{
		{
			Name:    "config-cmd-set-context-with-help",
			CmdLine: "--help",
			Cmd:     NewCmdConfigSetContext(nil),
		},
		{
			Name:    "config-cmd-set-context-no-flags",
			CmdLine: "context",
			Cmd:     NewCmdConfigSetContext(nil),
		},
		{
			Name:    "config-cmd-set-context-too-many-args",
			CmdLine: "arg1 arg2",
			Cmd:     NewCmdConfigSetContext(nil),
			Error:   fmt.Errorf("accepts at most %d arg(s), received %d", 1, 2),
		},
	}

	for _, tt := range cmdTests {
		testutil.RunTest(t, tt)
	}
}

func TestSetContext(t *testing.T) {
	given, cleanupGiven := config.InitConfig(t)
	defer cleanupGiven(t)

	tname := "dummycontext"
	tctype := config.Ephemeral

	expected, cleanupExpected := config.InitConfig(t)
	defer cleanupExpected(t)

	expected.Contexts[tname] = config.NewContext()
	clusterName := config.NewClusterComplexName()
	clusterName.WithType(tname, tctype)
	expected.Contexts[tname].NameInKubeconf = clusterName.Name()
	expected.Contexts[tname].Manifest = "edge_cloud"

	expkContext := kubeconfig.NewContext()
	expkContext.AuthInfo = testUser
	expkContext.Namespace = "kube-system"
	expected.KubeConfig().Contexts[expected.Contexts[tname].NameInKubeconf] = expkContext

	test := setContextTest{
		description: "Testing 'airshipctl config set-context' with a new context",
		givenConfig: given,
		args:        []string{tname},
		flags: []string{
			"--" + config.FlagClusterType + "=" + config.Target,
			"--" + config.FlagAuthInfoName + "=" + testUser,
			"--" + config.FlagManifest + "=edge_cloud",
			"--" + config.FlagNamespace + "=kube-system",
		},
		expectedOutput: fmt.Sprintf("Context %q created.\n", tname),
		expectedConfig: expected,
	}
	test.run(t)
}

func TestSetCurrentContextNoOptions(t *testing.T) {
	tname := "def_target"
	given, cleanupGiven := config.InitConfig(t)
	defer cleanupGiven(t)

	expected, cleanupExpected := config.InitConfig(t)
	defer cleanupExpected(t)

	expected.CurrentContext = "def_target"

	test := setContextTest{
		givenConfig:    given,
		args:           []string{tname},
		expectedOutput: fmt.Sprintf("Context %q not modified. No new options provided.\n", tname),
		expectedConfig: expected,
	}
	test.run(t)
}

func TestModifyContext(t *testing.T) {
	tname := testCluster
	tctype := config.Ephemeral

	given, cleanupGiven := config.InitConfig(t)
	defer cleanupGiven(t)

	given.Contexts[testCluster] = config.NewContext()

	clusterName := config.NewClusterComplexName()
	clusterName.WithType(tname, tctype)
	given.Contexts[tname].NameInKubeconf = clusterName.Name()
	kContext := kubeconfig.NewContext()
	kContext.AuthInfo = testUser
	given.KubeConfig().Contexts[clusterName.Name()] = kContext
	given.Contexts[tname].SetKubeContext(kContext)

	expected, cleanupExpected := config.InitConfig(t)
	defer cleanupExpected(t)

	expected.Contexts[tname] = config.NewContext()
	expected.Contexts[tname].NameInKubeconf = clusterName.Name()
	expkContext := kubeconfig.NewContext()
	expkContext.AuthInfo = testUser
	expected.KubeConfig().Contexts[clusterName.Name()] = expkContext
	expected.Contexts[tname].SetKubeContext(expkContext)

	test := setContextTest{
		description: "Testing 'airshipctl config set-context' with an existing context",
		givenConfig: given,
		args:        []string{tname},
		flags: []string{
			"--" + config.FlagAuthInfoName + "=" + testUser,
		},
		expectedOutput: fmt.Sprintf("Context %q modified.\n", tname),
		expectedConfig: expected,
	}
	test.run(t)
}

func (test setContextTest) run(t *testing.T) {
	// Get the Environment
	settings := &environment.AirshipCTLSettings{}
	settings.SetConfig(test.givenConfig)

	buf := bytes.NewBuffer([]byte{})

	cmd := NewCmdConfigSetContext(settings)
	cmd.SetOut(buf)
	cmd.SetArgs(test.args)
	err := cmd.Flags().Parse(test.flags)
	require.NoErrorf(t, err, "unexpected error flags args to command: %v,  flags: %v", err, test.flags)

	// Execute the Command
	// Which should Persist the File
	err = cmd.Execute()
	require.NoErrorf(t, err, "unexpected error executing command: %v, args: %v, flags: %v", err, test.args, test.flags)

	afterRunConf := settings.Config()

	// Find the Context Created or Modified
	afterRunContext, err := afterRunConf.GetContext(test.args[0])
	require.NoError(t, err)
	require.NotNil(t, afterRunContext)

	afterKcontext := afterRunContext.KubeContext()
	require.NotNil(t, afterKcontext)

	testKcontext := test.expectedConfig.KubeConfig().Contexts[test.expectedConfig.Contexts[test.args[0]].NameInKubeconf]
	require.NotNil(t, testKcontext)

	assert.EqualValues(t, afterKcontext.AuthInfo, testKcontext.AuthInfo)

	// Test that the Return Message looks correct
	if len(test.expectedOutput) != 0 {
		assert.EqualValues(t, test.expectedOutput, buf.String())
	}
}
