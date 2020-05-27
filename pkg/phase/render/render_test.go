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

package render_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"opendev.org/airship/airshipctl/pkg/environment"
	"opendev.org/airship/airshipctl/pkg/phase/render"
	"opendev.org/airship/airshipctl/testutil"
)

func TestRender(t *testing.T) {
	rs := &environment.AirshipCTLSettings{Config: testutil.DummyConfig()}
	fixturePath := "testdata/phase"
	tests := []struct {
		name       string
		settings   *render.Settings
		expResFile string
		expErr     error
	}{
		{
			name:       "No Filters",
			settings:   &render.Settings{AirshipCTLSettings: rs},
			expResFile: "noFilter.yaml",
			expErr:     nil,
		},
		{
			name: "All Filters",
			settings: &render.Settings{
				AirshipCTLSettings: rs,
				Label:              "airshipit.org/deploy-k8s=false",
				Annotation:         "airshipit.org/clustertype=ephemeral",
				APIVersion:         "metal3.io/v1alpha1",
				Kind:               "BareMetalHost",
			},
			expResFile: "allFilters.yaml",
			expErr:     nil,
		},
		{
			name: "Multiple Labels",
			settings: &render.Settings{
				AirshipCTLSettings: rs,
				Label:              "airshipit.org/deploy-k8s=false, airshipit.org/ephemeral-node=true",
			},
			expResFile: "multiLabels.yaml",
			expErr:     nil,
		},
		{
			name: "Malformed Label",
			settings: &render.Settings{
				AirshipCTLSettings: rs,
				Label:              "app=(",
			},
			expResFile: "",
			expErr:     fmt.Errorf("unable to parse requirement: found '(', expected: identifier"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			var expectedOut []byte
			var err error
			if tt.expResFile != "" {
				expectedOut, err = ioutil.ReadFile(path.Join("testdata", "expected", tt.expResFile))
				require.NoError(t, err)
			}
			out := &bytes.Buffer{}
			err = tt.settings.Render(fixturePath, out)
			assert.Equal(t, tt.expErr, err)
			assert.Equal(t, expectedOut, out.Bytes())
		})
	}
}
