/*
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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"opendev.org/airship/airshipctl/pkg/config"
	redfishdell "opendev.org/airship/airshipctl/pkg/remote/redfish/vendors/dell"
)

func TestNewManagementConfiguration(t *testing.T) {
	cfg := config.NewManagementConfiguration()
	assert.Equal(t, cfg.Type, config.AirshipDefaultManagementType)
}

func TestSetType(t *testing.T) {
	cfg := config.NewManagementConfiguration()

	err := cfg.SetType(redfishdell.ClientType)
	require.NoError(t, err)

	assert.Equal(t, cfg.Type, redfishdell.ClientType)
}

func TestSetTypeInvalid(t *testing.T) {
	cfg := config.NewManagementConfiguration()

	err := cfg.SetType("invalid")
	require.Error(t, err)

	assert.Equal(t, cfg.Type, config.AirshipDefaultManagementType)
}
func TestValidateDefault(t *testing.T) {
	cfg := config.NewManagementConfiguration()

	err := cfg.Validate()
	assert.NoError(t, err)
}

func TestValidateRedfishDell(t *testing.T) {
	cfg := config.NewManagementConfiguration()
	cfg.Type = redfishdell.ClientType

	err := cfg.Validate()
	assert.NoError(t, err)
}

func TestValidateInvalidManagementType(t *testing.T) {
	cfg := config.NewManagementConfiguration()
	cfg.Type = "invalid"

	err := cfg.Validate()
	assert.Error(t, err)
}
