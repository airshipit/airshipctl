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
	"fmt"
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

const (
	testMimeType = ".yaml"
	testDataDir  = "../../pkg/config/testdata"
)

func TestGetCluster(t *testing.T) {
	tname := "def"
	tctype := config.Ephemeral

	conf := config.InitConfig(t)

	// Retrieve one of the test
	theClusterIWant, err := conf.GetCluster(tname, tctype)
	require.NoError(t, err)
	require.NotNil(t, theClusterIWant)

	err = conf.Purge()
	require.NoErrorf(t, err, "unexpected error , unable to Purge before persisting the expected configuration: %v", err)
	err = conf.PersistConfig()
	require.NoErrorf(t, err, "unexpected error , unable to Persist the expected configuration: %v", err)

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

	expected := ""
	clusters, err := conf.GetClusters()
	require.NoError(t, err)
	for _, cluster := range clusters {
		expected += fmt.Sprintf("%s", cluster.PrettyString())
	}

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
	require.NoErrorf(t, err, "unexpected error flags args to command: %v,  flags: %v", err, test.flags)

	err = cmd.Execute()
	assert.NoErrorf(t, err, "unexpected error executing command: %v", err)
	if len(test.expected) != 0 {
		assert.EqualValues(t, test.expected, buf.String())
	}
}
