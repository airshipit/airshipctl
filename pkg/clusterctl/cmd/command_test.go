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
	"opendev.org/airship/airshipctl/pkg/environment"
)

const (
	manifestName = "dummy_manifest"
	validContext = "dummy_cluster"
)

// TODO (kkalynovskyi) expand test cases
func TestNewCommand(t *testing.T) {
	rs := &environment.AirshipCTLSettings{
		AirshipConfigPath: "testdata/airshipconfig.yaml",
		KubeConfigPath:    "testdata/kubeconfig.yaml",
	}
	rs.InitConfig()
	tests := []struct {
		name          string
		expectErr     bool
		currentConext string
		manifests     map[string]*config.Manifest
	}{
		{
			name:          "default success",
			currentConext: validContext,
			manifests: map[string]*config.Manifest{
				manifestName: {
					TargetPath:            "testdata",
					SubPath:               "valid",
					PrimaryRepositoryName: "primary",
					Repositories: map[string]*config.Repository{
						"primary": {},
					},
				},
			},
		},
		{
			name:          "Bundle build failure",
			currentConext: validContext,
			expectErr:     true,
			manifests: map[string]*config.Manifest{
				manifestName: {
					TargetPath:            "testdata",
					SubPath:               "invalid-path",
					PrimaryRepositoryName: "primary",
					Repositories: map[string]*config.Repository{
						"primary": {},
					},
				},
			},
		},
		{
			name:          "invalid clusterctl kind",
			currentConext: validContext,
			expectErr:     true,
			manifests: map[string]*config.Manifest{
				manifestName: {
					TargetPath:            "testdata",
					SubPath:               "no-clusterctl",
					PrimaryRepositoryName: "primary",
					Repositories: map[string]*config.Repository{
						"primary": {},
					},
				},
			},
		},
		{
			name:      "no primary repo",
			expectErr: true,
			manifests: map[string]*config.Manifest{
				manifestName: {
					SubPath: "no-clusterctl",
				},
			},
		},
		{
			name:          "cant find context",
			currentConext: "invalid-context",
			expectErr:     true,
			manifests: map[string]*config.Manifest{
				manifestName: {
					TargetPath:            "testdata",
					SubPath:               "can't build bundle",
					PrimaryRepositoryName: "primary",
					Repositories: map[string]*config.Repository{
						"primary": {},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		rs.InitConfig()
		expectErr := tt.expectErr
		manifests := tt.manifests
		rs.Config.Manifests = manifests
		context := tt.currentConext
		t.Run(tt.name, func(t *testing.T) {
			rs.Config.Manifests = manifests
			rs.Config.CurrentContext = context
			command, err := NewCommand(rs)
			if expectErr {
				assert.Error(t, err)
				assert.Nil(t, command)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
