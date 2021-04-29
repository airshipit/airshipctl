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

package kyamlutils_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"

	"opendev.org/airship/airshipctl/pkg/document/plugin/kyamlutils"
)

func documents(t *testing.T) []*yaml.RNode {
	docs := `---
apiVersion: v1
kind: Pod
metadata:
  labels:
    app: pod1
  name: p1
  namespace: capi
---
apiVersion: v1
kind: Pod
metadata:
  name: p2
  namespace: capi
---
apiVersion: v1beta1
kind: Deployment
metadata:
  name: p1
`
	rns, err := (&kio.ByteReader{Reader: bytes.NewBufferString(docs)}).Read()
	require.NoError(t, err)
	return rns
}

func TestFilter(t *testing.T) {
	docs := documents(t)
	testCases := []struct {
		name         string
		selector     kyamlutils.DocumentSelector
		expectedErr  error
		expectedDocs string
	}{
		{
			name: "Get by GVK + name + namespace",
			selector: kyamlutils.DocumentSelector{}.
				ByGVK("", "v1", "Pod").
				ByName("p1").
				ByNamespace("capi"),
			expectedDocs: `apiVersion: v1
kind: Pod
metadata:
  labels:
    app: pod1
  name: p1
  namespace: capi`,
		},
		{
			name:     "No filters",
			selector: kyamlutils.DocumentSelector{},
			expectedDocs: `apiVersion: v1
kind: Pod
metadata:
  labels:
    app: pod1
  name: p1
  namespace: capi
---
apiVersion: v1
kind: Pod
metadata:
  name: p2
  namespace: capi
---
apiVersion: v1beta1
kind: Deployment
metadata:
  name: p1`,
		},
		{
			name:     "Get by apiVersion",
			selector: kyamlutils.DocumentSelector{}.ByAPIVersion("v1beta1"),
			expectedDocs: `apiVersion: v1beta1
kind: Deployment
metadata:
  name: p1`,
		},
		{
			name:     "Get by empty name",
			selector: kyamlutils.DocumentSelector{}.ByAPIVersion("v1beta1").ByName(""),
			expectedDocs: `apiVersion: v1beta1
kind: Deployment
metadata:
  name: p1`,
		},
		{
			name:     "Get by version only",
			selector: kyamlutils.DocumentSelector{}.ByGVK("", "v1", ""),
			expectedDocs: `apiVersion: v1
kind: Pod
metadata:
  labels:
    app: pod1
  name: p1
  namespace: capi
---
apiVersion: v1
kind: Pod
metadata:
  name: p2
  namespace: capi`,
		},
		{
			name:     "Get by kind only",
			selector: kyamlutils.DocumentSelector{}.ByGVK("", "", "Deployment"),
			expectedDocs: `apiVersion: v1beta1
kind: Deployment
metadata:
  name: p1`,
		},
		{
			name:     "Get by empty namespace",
			selector: kyamlutils.DocumentSelector{}.ByGVK("", "v1beta1", "Deployment").ByNamespace(""),
			expectedDocs: `apiVersion: v1beta1
kind: Deployment
metadata:
  name: p1`,
		},
		{
			name:     "Get by label exact match",
			selector: kyamlutils.DocumentSelector{}.ByLabel("app=pod1"),
			expectedDocs: `apiVersion: v1
kind: Pod
metadata:
  labels:
    app: pod1
  name: p1
  namespace: capi`,
		},
		{
			name:     "Get by empty label",
			selector: kyamlutils.DocumentSelector{}.ByLabel(""),
			expectedDocs: `apiVersion: v1
kind: Pod
metadata:
  labels:
    app: pod1
  name: p1
  namespace: capi
---
apiVersion: v1
kind: Pod
metadata:
  name: p2
  namespace: capi
---
apiVersion: v1beta1
kind: Deployment
metadata:
  name: p1`,
		},
		{
			name:     "Get by label not equal",
			selector: kyamlutils.DocumentSelector{}.ByLabel("app!=pod1"),
			expectedDocs: `apiVersion: v1
kind: Pod
metadata:
  name: p2
  namespace: capi
---
apiVersion: v1beta1
kind: Deployment
metadata:
  name: p1`,
		},
		{
			name:     "Get by label inclusion",
			selector: kyamlutils.DocumentSelector{}.ByLabel("app in (pod1)"),
			expectedDocs: `apiVersion: v1
kind: Pod
metadata:
  labels:
    app: pod1
  name: p1
  namespace: capi`,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			filteredDocs, err := tc.selector.Filter(docs)
			assert.Equal(t, tc.expectedErr, err)

			buf := &bytes.Buffer{}
			err = kio.ByteWriter{Writer: buf}.Write(filteredDocs)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedDocs, strings.TrimSuffix(buf.String(), "\n"))
		})
	}
}
