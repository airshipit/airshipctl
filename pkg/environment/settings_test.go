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

func TestSpecifyAirConfigFromEnv(t *testing.T) {
	fakeConfig := "FakeConfigPath"
	err := os.Setenv(config.AirshipConfigEnv, fakeConfig)
	require.NoError(t, err)

	settings := &AirshipCTLSettings{}
	settings.InitConfig()

	assert.EqualValues(t, fakeConfig, settings.AirshipConfigPath())
}

func TestGetSetPaths(t *testing.T) {
	settings := &AirshipCTLSettings{}
	settings.InitConfig()
	airConfigFile := filepath.Join(config.AirshipConfigDir, config.AirshipConfig)
	kConfigFile := filepath.Join(config.AirshipConfigDir, config.AirshipKubeConfig)
	settings.SetAirshipConfigPath(airConfigFile)
	assert.EqualValues(t, airConfigFile, settings.AirshipConfigPath())

	settings.SetKubeConfigPath(kConfigFile)
	assert.EqualValues(t, kConfigFile, settings.KubeConfigPath())
}
