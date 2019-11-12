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

package config

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/environment"
)

type getClusterTest struct {
	config   *config.Config
	args     []string
	flags    []string
	expected string
}

func TestGetCluster(t *testing.T) {
	tname := "def"
	tctype := config.Ephemeral

	conf := config.InitConfig(t)

	// Retrieve one of the test
	theClusterIWant, err := conf.GetCluster(tname, tctype)
	require.NoError(t, err)

	err = conf.Purge()
	require.NoError(t, err, "Unable to Purge before persisting the expected configuration")
	err = conf.PersistConfig()
	require.NoError(t, err, "Unable to Persist the expected configuration")

	test := getClusterTest{
		config: conf,
		args:   []string{tname},
		flags: []string{
			"--" + config.FlagClusterType + "=" + config.Ephemeral,
		},
		expected: theClusterIWant.PrettyString(),
	}

	test.run(t)
}

func TestGetAllClusters(t *testing.T) {
	conf := config.InitConfig(t)

	testDir := filepath.Dir(conf.LoadedConfigPath())
	kubeconfigPath := filepath.Join(testDir, "kubeconfig")

	expected := `Cluster: def
ephemeral:
bootstrap-info: ""
cluster-kubeconf: def_ephemeral

LocationOfOrigin: ` + kubeconfigPath + `
insecure-skip-tls-verify: true
server: http://5.6.7.8

Cluster: def
target:
bootstrap-info: ""
cluster-kubeconf: def_target

LocationOfOrigin: ` + kubeconfigPath + `
insecure-skip-tls-verify: true
server: http://1.2.3.4

Cluster: onlyinkubeconf
target:
bootstrap-info: ""
cluster-kubeconf: onlyinkubeconf_target

LocationOfOrigin: ` + kubeconfigPath + `
insecure-skip-tls-verify: true
server: http://9.10.11.12

Cluster: wrongonlyinkubeconf
target:
bootstrap-info: ""
cluster-kubeconf: wrongonlyinkubeconf_target

LocationOfOrigin: ` + kubeconfigPath + `
certificate-authority: cert_file
server: ""

`

	test := getClusterTest{
		config:   conf,
		args:     []string{},
		flags:    []string{},
		expected: expected,
	}

	test.run(t)
}

func (test getClusterTest) run(t *testing.T) {
	// Get the Environment
	settings := &environment.AirshipCTLSettings{}
	settings.SetConfig(test.config)
	buf := bytes.NewBuffer([]byte{})
	cmd := NewCmdConfigGetCluster(settings)
	cmd.SetOutput(buf)
	cmd.SetArgs(test.args)
	err := cmd.Flags().Parse(test.flags)
	require.NoErrorf(t, err, "unexpected error flags args to command: %v, flags: %v", err, test.flags)

	err = cmd.Execute()
	require.NoError(t, err)
	if len(test.expected) != 0 {
		assert.EqualValues(t, test.expected, buf.String())
	}
}
