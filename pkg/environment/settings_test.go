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

package environment_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"

	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/environment"
	"opendev.org/airship/airshipctl/testutil"
)

func TestInitFlags(t *testing.T) {
	// Get the Environment
	settings := &environment.AirshipCTLSettings{}
	testCmd := &cobra.Command{}
	settings.InitFlags(testCmd)
	assert.True(t, testCmd.HasPersistentFlags())
}

func TestInitConfig(t *testing.T) {
	t.Run("DefaultToHomeDirectory", func(subTest *testing.T) {
		// Set up a fake $HOME directory
		testDir, cleanup := testutil.TempDir(t, "test-home")
		defer cleanup(t)
		defer setHome(testDir)()

		var testSettings environment.AirshipCTLSettings
		expectedAirshipConfig := filepath.Join(testDir, config.AirshipConfigDir, config.AirshipConfig)
		expectedKubeConfig := filepath.Join(testDir, config.AirshipConfigDir, config.AirshipKubeConfig)
		expectedPluginPath := filepath.Join(testDir, config.AirshipConfigDir, config.AirshipPluginPath)

		testSettings.InitAirshipConfigPath()
		testSettings.InitKubeConfigPath()
		environment.InitPluginPath()
		assert.Equal(t, expectedAirshipConfig, testSettings.AirshipConfigPath)
		assert.Equal(t, expectedKubeConfig, testSettings.KubeConfigPath)
		assert.Equal(t, expectedPluginPath, environment.PluginPath())
	})

	t.Run("PreferEnvToDefault", func(subTest *testing.T) {
		// Set up a fake $HOME directory
		testDir, cleanup := testutil.TempDir(t, "test-home")
		defer cleanup(t)
		defer setHome(testDir)()

		var testSettings environment.AirshipCTLSettings
		expectedAirshipConfig := filepath.Join(testDir, "airshipEnv")
		expectedKubeConfig := filepath.Join(testDir, "kubeEnv")
		expectedPluginPath := filepath.Join(testDir, "pluginPath")

		os.Setenv(config.AirshipConfigEnv, expectedAirshipConfig)
		os.Setenv(config.AirshipKubeConfigEnv, expectedKubeConfig)
		os.Setenv(config.AirshipPluginPathEnv, expectedPluginPath)
		defer os.Unsetenv(config.AirshipConfigEnv)
		defer os.Unsetenv(config.AirshipKubeConfigEnv)
		defer os.Unsetenv(config.AirshipPluginPathEnv)

		testSettings.InitAirshipConfigPath()
		testSettings.InitKubeConfigPath()
		environment.InitPluginPath()
		assert.Equal(t, expectedAirshipConfig, testSettings.AirshipConfigPath)
		assert.Equal(t, expectedKubeConfig, testSettings.KubeConfigPath)
		assert.Equal(t, expectedPluginPath, environment.PluginPath())
	})

	t.Run("PreferCmdLineArgToDefault", func(subTest *testing.T) {
		// Set up a fake $HOME directory
		testDir, cleanup := testutil.TempDir(t, "test-home")
		defer cleanup(t)
		defer setHome(testDir)()

		expectedAirshipConfig := filepath.Join(testDir, "airshipCmdLine")
		expectedKubeConfig := filepath.Join(testDir, "kubeCmdLine")

		testSettings := environment.AirshipCTLSettings{
			AirshipConfigPath: expectedAirshipConfig,
			KubeConfigPath:    expectedKubeConfig,
		}

		testSettings.InitAirshipConfigPath()
		testSettings.InitKubeConfigPath()
		assert.Equal(t, expectedAirshipConfig, testSettings.AirshipConfigPath)
		assert.Equal(t, expectedKubeConfig, testSettings.KubeConfigPath)
	})

	t.Run("PreferCmdLineArgToEnv", func(subTest *testing.T) {
		// Set up a fake $HOME directory
		testDir, cleanup := testutil.TempDir(t, "test-home")
		defer cleanup(t)
		defer setHome(testDir)()

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

		testSettings := environment.AirshipCTLSettings{
			AirshipConfigPath: expectedAirshipConfig,
			KubeConfigPath:    expectedKubeConfig,
		}

		testSettings.InitAirshipConfigPath()
		testSettings.InitKubeConfigPath()
		assert.Equal(t, expectedAirshipConfig, testSettings.AirshipConfigPath)
		assert.Equal(t, expectedKubeConfig, testSettings.KubeConfigPath)
	})
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
