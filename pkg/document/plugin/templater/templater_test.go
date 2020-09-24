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

package templater_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/yaml"

	"opendev.org/airship/airshipctl/pkg/document/plugin/templater"
)

func TestTemplater(t *testing.T) {
	testCases := []struct {
		cfg         string
		expectedOut string
		expectedErr string
	}{
		{
			cfg: `
apiVersion: airshipit.org/v1alpha1
kind: Templater
metadata:
  name: notImportantHere
values:
  hosts:
    - macAddress: 00:aa:bb:cc:dd
      name: node-1
    - macAddress: 00:aa:bb:cc:ee
      name: node-2
template: |
  {{ range .hosts -}}
  ---
  apiVersion: metal3.io/v1alpha1
  kind: BareMetalHost
  metadata:
    name: {{ .name }}
  spec:
    bootMACAddress: {{ .macAddress }}
  {{ end -}}`,
			expectedOut: `---
apiVersion: metal3.io/v1alpha1
kind: BareMetalHost
metadata:
  name: node-1
spec:
  bootMACAddress: 00:aa:bb:cc:dd
---
apiVersion: metal3.io/v1alpha1
kind: BareMetalHost
metadata:
  name: node-2
spec:
  bootMACAddress: 00:aa:bb:cc:ee
`,
		},
		{
			cfg: `
apiVersion: airshipit.org/v1alpha1
kind: Templater
metadata:
  name: notImportantHere
values:
  test:
    of:
      - toYaml
template: |
  {{ toYaml . -}}
`,
			expectedOut: `test:
  of:
  - toYaml
`,
		},
		{
			cfg: `
apiVersion: airshipit.org/v1alpha1
kind: Templater
metadata:
  name: notImportantHere
values:
  test:
    of:
      - badToYamlInput
template: |
  {{ toYaml ignorethisbadinput -}}
`,
			expectedOut: ``,
		},
		{
			cfg: `
apiVersion: airshipit.org/v1alpha1
kind: Templater
metadata:
  name: notImportantHere
template: |
  {{ end }`,
			expectedErr: "template: tmpl:1: unexpected \"}\" in end",
		},
	}

	for _, tc := range testCases {
		cfg := make(map[string]interface{})
		err := yaml.Unmarshal([]byte(tc.cfg), &cfg)
		require.NoError(t, err)
		plugin, err := templater.New(cfg)
		require.NoError(t, err)
		buf := &bytes.Buffer{}
		err = plugin.Run(nil, buf)
		if tc.expectedErr != "" {
			assert.EqualError(t, err, tc.expectedErr)
		}
		assert.Equal(t, tc.expectedOut, buf.String())
	}
}
