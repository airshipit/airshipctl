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

package inventory_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/inventory"
)

func TestBaremetalInventory(t *testing.T) {
	tests := []struct {
		name      string
		errString string

		factory config.Factory
	}{
		{
			name:      "error no metadata file",
			errString: "no such file or directory",

			factory: func() (*config.Config, error) {
				return config.NewConfig(), nil
			},
		},
		{
			name:      "error no management config",
			errString: "Management configuration",

			factory: func() (*config.Config, error) {
				cfg := config.NewConfig()
				cfg.ManagementConfiguration = nil
				return cfg, nil
			},
		},
		{
			name:      "error no manifest defined",
			errString: "missing configuration: manifest named",

			factory: func() (*config.Config, error) {
				cfg := config.NewConfig()
				// empty manifest map
				cfg.Manifests = make(map[string]*config.Manifest)
				return cfg, nil
			},
		},
		{
			name: "success config",

			factory: func() (*config.Config, error) {
				cfg := config.NewConfig()
				manifest, err := cfg.CurrentContextManifest()
				require.NoError(t, err)
				manifest.MetadataPath = "metadata.yaml"
				manifest.PhaseRepositoryName = "testdata"
				manifest.InventoryRepositoryName = "testdata"
				manifest.Repositories["testdata"] = &config.Repository{
					URLString: "/myrepo/testdata",
				}
				manifest.TargetPath = "."
				return cfg, nil
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			i := inventory.NewInventory(tt.factory)
			bmhInv, err := i.BaremetalInventory()
			if tt.errString != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errString)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, bmhInv)
			}
		})
	}
}
