/*
 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     https://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"opendev.org/airship/airshipctl/pkg/config"
)

const (
	manifestName = "dummy_manifest"
	validContext = "dummy_cluster"
)

// TODO (kkalynovskyi) expand test cases
func TestNewCommand(t *testing.T) {
	airshipConfigPath := "testdata/airshipconfig.yaml"
	cfg, err := config.CreateFactory(&airshipConfigPath)()
	require.NoError(t, err)

	tests := []struct {
		name           string
		expectErr      bool
		currentContext string
		manifests      map[string]*config.Manifest
	}{
		{
			name:           "default success",
			currentContext: validContext,
			manifests: map[string]*config.Manifest{
				manifestName: {
					TargetPath:          "testdata",
					SubPath:             "valid",
					PhaseRepositoryName: "primary",
					Repositories: map[string]*config.Repository{
						"primary": {},
					},
				},
			},
		},
		{
			name:           "Bundle build failure",
			currentContext: validContext,
			expectErr:      true,
			manifests: map[string]*config.Manifest{
				manifestName: {
					TargetPath:          "testdata",
					SubPath:             "invalid-path",
					PhaseRepositoryName: "primary",
					Repositories: map[string]*config.Repository{
						"primary": {},
					},
				},
			},
		},
		{
			name:           "invalid clusterctl kind",
			currentContext: validContext,
			expectErr:      true,
			manifests: map[string]*config.Manifest{
				manifestName: {
					TargetPath:          "testdata",
					SubPath:             "no-clusterctl",
					PhaseRepositoryName: "primary",
					Repositories: map[string]*config.Repository{
						"primary": {},
					},
				},
			},
		},
		{
			name:      "no phase repo",
			expectErr: true,
			manifests: map[string]*config.Manifest{
				manifestName: {
					SubPath: "no-clusterctl",
				},
			},
		},
		{
			name:           "cant find context",
			currentContext: "invalid-context",
			expectErr:      true,
			manifests: map[string]*config.Manifest{
				manifestName: {
					TargetPath:          "testdata",
					SubPath:             "can't build bundle",
					PhaseRepositoryName: "primary",
					Repositories: map[string]*config.Repository{
						"primary": {},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		expectErr := tt.expectErr
		manifests := tt.manifests
		cfg.Manifests = manifests
		context := tt.currentContext
		t.Run(tt.name, func(t *testing.T) {
			cfg.Manifests = manifests
			cfg.CurrentContext = context
			command, err := NewCommand(func() (*config.Config, error) {
				return cfg, nil
			}, "")
			if expectErr {
				assert.Error(t, err)
				assert.Nil(t, command)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
