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

	"opendev.org/airship/airshipctl/pkg/config"
)

func TestContextOptionsValidate(t *testing.T) {
	tests := []struct {
		name        string
		testOptions config.ContextOptions
		expectError bool
	}{
		{
			name: "MissingName",
			testOptions: config.ContextOptions{
				Name: "",
			},
			expectError: true,
		},
		{
			name: "SettingCurrentContext",
			testOptions: config.ContextOptions{
				Name:           "testContext",
				CurrentContext: true,
			},
			expectError: false,
		},
		{
			name: "NoClusterType",
			testOptions: config.ContextOptions{
				Name: "testContext",
			},
			expectError: false,
		},
		{
			name: "ValidClusterType",
			testOptions: config.ContextOptions{
				Name:        "testContext",
				ClusterType: "target",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(subTest *testing.T) {
			err := tt.testOptions.Validate()
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
