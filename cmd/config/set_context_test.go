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

package config_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	cmd "opendev.org/airship/airshipctl/cmd/config"
	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/testutil"
)

const (
	testEncryptionConfig = "test_encryption_config"
	defaultManifest      = "edge_cloud"
	testManifest         = "test_manifest"
)

type setContextTest struct {
	givenConfig *config.Config
	cmdTest     *testutil.CmdTest
	contextName string
	manifest    string
}

func TestConfigSetContext(t *testing.T) {
	cmdTests := []*testutil.CmdTest{
		{
			Name:    "config-cmd-set-context-with-help",
			CmdLine: "--help",
			Cmd:     cmd.NewSetContextCommand(nil),
		},
		{
			Name:    "config-cmd-set-context-no-flags",
			CmdLine: "context",
			Cmd:     cmd.NewSetContextCommand(nil),
		},
		{
			Name:    "config-cmd-set-context-too-many-args",
			CmdLine: "arg1 arg2",
			Cmd:     cmd.NewSetContextCommand(nil),
			Error:   fmt.Errorf("accepts at most %d arg(s), received %d", 1, 2),
		},
	}

	for _, tt := range cmdTests {
		testutil.RunTest(t, tt)
	}
}

func TestSetContext(t *testing.T) {
	given, cleanupGiven := testutil.InitConfig(t)
	defer cleanupGiven(t)

	tests := []struct {
		testName         string
		contextName      string
		flags            []string
		givenConfig      *config.Config
		manifest         string
		encryptionConfig string
	}{
		{
			testName:    "set-context",
			contextName: "dummycontext",
			flags: []string{
				"--cluster-type=target",
				"--manifest=" + defaultManifest,
				"--encryption-config=" + testEncryptionConfig,
			},
			givenConfig:      given,
			manifest:         defaultManifest,
			encryptionConfig: testEncryptionConfig,
		},
		{
			testName:    "set-current-context",
			contextName: "def_target",
			flags:       []string{},
			givenConfig: given,
		},
		{
			testName:    "modify-context",
			contextName: "def_target",
			flags: []string{
				"--manifest=" + testManifest,
			},
			givenConfig: given,
			manifest:    testManifest,
		},
		{
			testName:    "modify-context",
			contextName: "def_target",
			flags: []string{
				"--encryption-config=" + testEncryptionConfig,
			},
			givenConfig:      given,
			encryptionConfig: testEncryptionConfig,
		},
	}

	for _, tt := range tests {
		tt := tt
		cmd := &testutil.CmdTest{
			Name:    tt.testName,
			CmdLine: fmt.Sprintf("%s %s", tt.contextName, strings.Join(tt.flags, " ")),
		}
		test := setContextTest{
			givenConfig: tt.givenConfig,
			cmdTest:     cmd,
			contextName: tt.contextName,
			manifest:    tt.manifest,
		}
		test.run(t)
	}
}

func (test setContextTest) run(t *testing.T) {
	// Get the Environment
	settings := func() (*config.Config, error) {
		return test.givenConfig, nil
	}

	test.cmdTest.Cmd = cmd.NewSetContextCommand(settings)
	testutil.RunTest(t, test.cmdTest)

	afterRunConf := test.givenConfig

	// Find the Context Created or Modified
	afterRunContext, err := afterRunConf.GetContext(test.contextName)
	require.NoError(t, err)
	require.NotNil(t, afterRunContext)

	if test.manifest != "" {
		assert.EqualValues(t, afterRunContext.Manifest, test.manifest)
	}
}
