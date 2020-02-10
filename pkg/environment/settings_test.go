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

package environment

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"opendev.org/airship/airshipctl/pkg/config"
)

func TestInitFlags(t *testing.T) {
	// Get the Environment
	settings := &AirshipCTLSettings{}
	testCmd := &cobra.Command{}
	settings.InitFlags(testCmd)
	assert.True(t, testCmd.HasPersistentFlags())
}

func TestInitConfig(t *testing.T) {
	t.Run("DefaultToHomeDirectory", func(subTest *testing.T) {
		// Set up a fake $HOME directory
		testDir := makeTestDir(t)
		defer deleteTestDir(t, testDir)
		defer setHome(testDir)()

		var testSettings AirshipCTLSettings
		expectedAirshipConfig := filepath.Join(testDir, config.AirshipConfigDir, config.AirshipConfig)
		expectedKubeConfig := filepath.Join(testDir, config.AirshipConfigDir, config.AirshipKubeConfig)

		testSettings.InitConfig()
		assert.Equal(t, expectedAirshipConfig, testSettings.AirshipConfigPath())
		assert.Equal(t, expectedKubeConfig, testSettings.KubeConfigPath())
	})

	t.Run("PreferEnvToDefault", func(subTest *testing.T) {
		// Set up a fake $HOME directory
		testDir := makeTestDir(t)
		defer deleteTestDir(t, testDir)
		defer setHome(testDir)()

		var testSettings AirshipCTLSettings
		expectedAirshipConfig := filepath.Join(testDir, "airshipEnv")
		expectedKubeConfig := filepath.Join(testDir, "kubeEnv")

		os.Setenv(config.AirshipConfigEnv, expectedAirshipConfig)
		os.Setenv(config.AirshipKubeConfigEnv, expectedKubeConfig)
		defer os.Unsetenv(config.AirshipConfigEnv)
		defer os.Unsetenv(config.AirshipKubeConfigEnv)

		testSettings.InitConfig()
		assert.Equal(t, expectedAirshipConfig, testSettings.AirshipConfigPath())
		assert.Equal(t, expectedKubeConfig, testSettings.KubeConfigPath())
	})

	t.Run("PreferCmdLineArgToDefault", func(subTest *testing.T) {
		// Set up a fake $HOME directory
		testDir := makeTestDir(t)
		defer deleteTestDir(t, testDir)
		defer setHome(testDir)()

		var testSettings AirshipCTLSettings
		expectedAirshipConfig := filepath.Join(testDir, "airshipCmdLine")
		expectedKubeConfig := filepath.Join(testDir, "kubeCmdLine")

		testSettings.SetAirshipConfigPath(expectedAirshipConfig)
		testSettings.SetKubeConfigPath(expectedKubeConfig)

		// InitConfig should not change any values
		testSettings.InitConfig()
		assert.Equal(t, expectedAirshipConfig, testSettings.AirshipConfigPath())
		assert.Equal(t, expectedKubeConfig, testSettings.KubeConfigPath())
	})

	t.Run("PreferCmdLineArgToEnv", func(subTest *testing.T) {
		// Set up a fake $HOME directory
		testDir := makeTestDir(t)
		defer deleteTestDir(t, testDir)
		defer setHome(testDir)()

		var testSettings AirshipCTLSettings
		expectedAirshipConfig := filepath.Join(testDir, "airshipCmdLine")
		expectedKubeConfig := filepath.Join(testDir, "kubeCmdLine")

		// set up "decoy" environment variables. These should be
		// ignored, since we're simulating passing in command line
		// arguments
		wrongAirshipConfig := filepath.Join(testDir, "wrongAirshipConfig")
		wrongKubeConfig := filepath.Join(testDir, "wrongKubeConfig")
		os.Setenv(config.AirshipConfigEnv, wrongAirshipConfig)
		os.Setenv(config.AirshipKubeConfigEnv, wrongKubeConfig)
		defer os.Unsetenv(config.AirshipConfigEnv)
		defer os.Unsetenv(config.AirshipKubeConfigEnv)

		testSettings.SetAirshipConfigPath(expectedAirshipConfig)
		testSettings.SetKubeConfigPath(expectedKubeConfig)

		testSettings.InitConfig()
		assert.Equal(t, expectedAirshipConfig, testSettings.AirshipConfigPath())
		assert.Equal(t, expectedKubeConfig, testSettings.KubeConfigPath())
	})
}

func makeTestDir(t *testing.T) string {
	testDir, err := ioutil.TempDir("", "airship-test")
	require.NoError(t, err)
	return testDir
}

func deleteTestDir(t *testing.T, path string) {
	err := os.Remove(path)
	require.NoError(t, err)
}

// setHome sets the HOME environment variable to `path`, and returns a function
// that can be used to reset HOME to its original value
func setHome(path string) (resetHome func()) {
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", path)
	return func() {
		os.Setenv("HOME", oldHome)
	}
}
