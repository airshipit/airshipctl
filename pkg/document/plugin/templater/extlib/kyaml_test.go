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

package extlib

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"bytes"

	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestKOneFilter(t *testing.T) {
	testCases := []struct {
		in          string
		filter      string
		expectedOut string
	}{
		{
			in: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: cf1
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: cf2
---
apiVersion: v1
kind: Deployment
metadata:
  name: cf1
`,
			filter: `
kind: GrepFilter
path:
- metadata
- name
value: cf2
`,
			expectedOut: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: cf2
`,
		},
		{
			in: `
somedata: a
`,
			filter: `
kind: invalidFilter
`,
			expectedOut: "",
		},
	}

	for _, tc := range testCases {
		// convert in to []*yaml.RNode
		b := kio.PackageBuffer{}
		p := kio.Pipeline{
			Inputs:  []kio.Reader{&kio.ByteReader{Reader: bytes.NewBufferString(tc.in)}},
			Outputs: []kio.Writer{&b},
		}
		err := p.Execute()
		require.NoError(t, err)

		nodes := kOneFilter(b.Nodes, tc.filter)
		if tc.expectedOut == "" && nodes == nil {
			continue
		}
		// convert to string and compare with expected
		out := &bytes.Buffer{}
		err = kio.ByteWriter{Writer: out}.Write(nodes)
		require.NoError(t, err)
		assert.Equal(t, tc.expectedOut[1:], out.String())
	}
}

func TestYOneFilter(t *testing.T) {
	testCases := []struct {
		in          string
		filter      string
		expectedOut string
	}{
		{
			in: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: cf1
`,
			filter: `
kind: PathGetter
path: ["metadata"]
`,
			expectedOut: `
name: cf1
`,
		},
		{
			in: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: cf1
`,
			filter: `
kind: InvalidFilter
`,
			expectedOut: "",
		},
	}

	for _, tc := range testCases {
		inRNode, err := yaml.Parse(tc.in)
		require.NoError(t, err)

		outRNode := yOneFilter(inRNode, tc.filter)

		if tc.expectedOut != "" {
			require.NotNil(t, outRNode)
			out, err := outRNode.String()
			require.NoError(t, err)
			assert.Equal(t, tc.expectedOut[1:], out)
		}
	}
}

func TestMerge(t *testing.T) {
	y1 := strToY(`
kind: x1
value1: y`)
	y2 := strToY(`
kind: x2
value2: z`)
	ym := yMerge(y1, y2)
	res, err := ym.String()
	require.NoError(t, err)
	assert.Equal(t, `
kind: x1
value2: z
value1: y
`[1:],
		res)
}

func TestListAppend(t *testing.T) {
	y1 := strToY(`values:
- name: x
- name: z
`)
	list, err := y1.Pipe(yaml.PathGetter{Path: []string{"values"}})
	require.NoError(t, err)
	y2 := strToY(`
name: y`)
	yListAppend(list, y2)
	res, err := y1.String()
	require.NoError(t, err)
	assert.Equal(t, `values:
- name: x
- name: z
- name: y
`,
		res)
}
