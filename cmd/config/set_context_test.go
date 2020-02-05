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
	config         *config.Config
	args           []string
	flags          []string
	expected       string
	expectedConfig *config.Config
}

func TestConfigSetContext(t *testing.T) {
	cmdTests := []*testutil.CmdTest{
		{
			Name:    "config-cmd-set-context-with-help",
			CmdLine: "--help",
			Cmd:     NewCmdConfigSetContext(nil),
		},
	}

	for _, tt := range cmdTests {
		testutil.RunTest(t, tt)
	}
}

func TestSetContext(t *testing.T) {
	conf := config.InitConfig(t)

	tname := "dummycontext"
	tctype := config.Ephemeral

	expconf := config.InitConfig(t)
	expconf.Contexts[tname] = config.NewContext()
	clusterName := config.NewClusterComplexName()
	clusterName.WithType(tname, tctype)
	expconf.Contexts[tname].NameInKubeconf = clusterName.Name()
	expconf.Contexts[tname].Manifest = "edge_cloud"

	expkContext := kubeconfig.NewContext()
	expkContext.AuthInfo = testUser
	expkContext.Namespace = "kube-system"
	expconf.KubeConfig().Contexts[expconf.Contexts[tname].NameInKubeconf] = expkContext

	test := setContextTest{
		description: "Testing 'airshipctl config set-context' with a new context",
		config:      conf,
		args:        []string{tname},
		flags: []string{
			"--" + config.FlagClusterType + "=" + config.Target,
			"--" + config.FlagAuthInfoName + "=" + testUser,
			"--" + config.FlagManifest + "=edge_cloud",
			"--" + config.FlagNamespace + "=kube-system",
		},
		expected:       `Context "` + tname + `" created.` + "\n",
		expectedConfig: expconf,
	}
	test.run(t)
}

func TestSetCurrentContext(t *testing.T) {
	tname := "def_target"
	conf := config.InitConfig(t)

	expconf := config.InitConfig(t)
	expconf.CurrentContext = "def_target"

	test := setContextTest{
		description: "Testing 'airshipctl config set-context' with a new current context",
		config:      conf,
		args:        []string{tname},
		flags: []string{
			"--" + config.FlagCurrentContext + "=true",
		},
		expected:       `Context "` + tname + `" modified.` + "\n",
		expectedConfig: expconf,
	}
	test.run(t)
}
func TestModifyContext(t *testing.T) {
	tname := testCluster
	tctype := config.Ephemeral

	conf := config.InitConfig(t)
	conf.Contexts[tname] = config.NewContext()

	clusterName := config.NewClusterComplexName()
	clusterName.WithType(tname, tctype)
	conf.Contexts[tname].NameInKubeconf = clusterName.Name()
	kContext := kubeconfig.NewContext()
	kContext.AuthInfo = testUser
	conf.KubeConfig().Contexts[clusterName.Name()] = kContext
	conf.Contexts[tname].SetKubeContext(kContext)

	expconf := config.InitConfig(t)
	expconf.Contexts[tname] = config.NewContext()
	expconf.Contexts[tname].NameInKubeconf = clusterName.Name()
	expkContext := kubeconfig.NewContext()
	expkContext.AuthInfo = testUser
	expconf.KubeConfig().Contexts[clusterName.Name()] = expkContext
	expconf.Contexts[tname].SetKubeContext(expkContext)

	test := setContextTest{
		description: "Testing 'airshipctl config set-context' with an existing context",
		config:      conf,
		args:        []string{tname},
		flags: []string{
			"--" + config.FlagAuthInfoName + "=" + testUser,
		},
		expected:       `Context "` + tname + `" modified.` + "\n",
		expectedConfig: expconf,
	}
	test.run(t)
}

func (test setContextTest) run(t *testing.T) {
	// Get the Environment
	settings := &environment.AirshipCTLSettings{}
	settings.SetConfig(test.config)

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
	if len(test.expected) != 0 {
		assert.EqualValuesf(t, buf.String(), test.expected, "expected %v, but got %v", test.expected, buf.String())
	}
}
