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

package metadata_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"opendev.org/airship/airshipctl/pkg/config"
	meta "opendev.org/airship/airshipctl/pkg/document/metadata"
	"opendev.org/airship/airshipctl/testutil"
)

func TestConfig(t *testing.T) {
	rs := testutil.DummyConfig()
	dummyManifest := rs.Manifests["dummy_manifest"]
	dummyManifest.TargetPath = "testdata"
	dummyManifest.PhaseRepositoryName = config.DefaultTestPhaseRepo
	dummyManifest.Repositories = map[string]*config.Repository{
		config.DefaultTestPhaseRepo: {
			URLString: "",
		},
	}
	tests := []struct {
		name         string
		metadataFile string
		expError     string
	}{
		{
			name:         "valid metadata file",
			metadataFile: "valid_metadata.yaml",
			expError:     "",
		},
		{
			name:         "non existent metadata file",
			metadataFile: "nonexistent_metadata.yaml",
			expError:     "no such file or directory",
		},
		{
			name:         "invalid metadata file",
			metadataFile: "invalid_metadata.yaml",
			expError:     "missing Resource metadata",
		},
		{
			name:         "incomplete metadata file",
			metadataFile: "incomplete_metadata.yaml",
			expError:     "no field named 'spec.phase.docEntryPointPrefix'",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			dummyManifest.MetadataPath = tt.metadataFile
			metadataPath := filepath.Join("testdata", tt.metadataFile)
			metadata, err := meta.Config(metadataPath)
			if tt.expError != "" {
				assert.Contains(t, err.Error(), tt.expError)
				assert.Equal(t, metadata, meta.Metadata{})
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, metadata)
			}
		})
	}
}
