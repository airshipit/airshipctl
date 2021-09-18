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

	"bytes"
	"strings"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"

	kfilters "sigs.k8s.io/kustomize/kyaml/kio/filters"
)

func TestKFilter(t *testing.T) {
	testCases := []struct {
		in          string
		expectedOut kio.Filter
	}{
		{
			in: `
kind: GrepFilter
path:
- metadata
- annotations
- test-annotation
value: ^x$
invertMatch: true
`,
			expectedOut: kfilters.GrepFilter{
				Path: []string{
					"metadata",
					"annotations",
					"test-annotation",
				},
				Value:       "^x$",
				InvertMatch: true,
			},
		},
		{
			in: `
kind: NonExistentFilter
path:
- metadata
`,
			expectedOut: nil,
		},
		{
			in: `
kind: GrepFilter
path: "incorrectdata"
`,
			expectedOut: nil,
		},
		{
			in: `
kind: Modifier
pipeline: "incorrectdata"
`,
			expectedOut: nil,
		},
	}

	for _, tc := range testCases {
		r := kFilter(tc.in)

		// GrepFilter is a special case
		grepFilter, ok := r.(kfilters.GrepFilter)
		if ok {
			require.NotNil(t, grepFilter.Compare)
			grepFilter.Compare = nil
			r = grepFilter
		}

		assert.Equal(t, tc.expectedOut, r)
	}
}

func TestKPipe(t *testing.T) {
	testCases := []struct {
		in          string
		filters     string
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
			filters: `
kind: GrepFilter
path:
- metadata
- name
value: cf1
---
kind: GrepFilter
path:
- kind
value: ConfigMap
`,
			expectedOut: `
apiVersion: v1
kind: ConfigMap
metadata:
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
			filters: `
kind: InvalidFilter
`,
			expectedOut: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: cf1
`,
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

		// get list of filters
		kfilters := []interface{}{}
		for _, flt := range strings.Split(tc.filters, "\n---\n") {
			kfilters = append(kfilters, kFilter(flt))
		}

		nodes := kPipe(b.Nodes, kfilters)

		// convert to string and compare with expected
		out := &bytes.Buffer{}
		err = kio.ByteWriter{Writer: out}.Write(nodes)
		require.NoError(t, err)
		assert.Equal(t, tc.expectedOut[1:], out.String())
	}
}

func TestYFilter(t *testing.T) {
	testCases := []struct {
		in          string
		expectedOut yaml.Filter
	}{
		{
			in: `
kind: PathGetter
path: ["data", "fld1"]
`,
			expectedOut: &yaml.PathGetter{
				Kind: "PathGetter",
				Path: []string{
					"data",
					"fld1",
				},
			},
		},
		{in: `
kind: PathGetter
path: "data"
`,
			expectedOut: nil,
		},
		{in: `
kind: nonExistingFilter
path: "data"
`,
			expectedOut: nil,
		},
	}

	for _, tc := range testCases {
		out := yFilter(tc.in)
		assert.Equal(t, tc.expectedOut, out)
	}
}

func TestYPipe(t *testing.T) {
	testCases := []struct {
		in          string
		filters     string
		expectedIn  string
		expectedOut string
	}{
		{
			in: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: cf1
`,
			filters: `
kind: PathGetter
path: ["metadata"]
---
kind: FieldSetter
name: "name"
stringValue: "cf2"
`,
			expectedIn: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: cf2
`,
			expectedOut: `
cf2
`,
		},
		{
			in: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: cf1
`,
			filters: `
kind: InvalidPathGetter
path: ["metadata"]
`,
		},
		{
			in: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: cf1
`,
			filters: `
kind: PathGetter
path: ["xmetadata"]
---
kind: FieldSetter
name: "namex"
stringValue: "cf2"
`,
			expectedIn: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: cf1
`,
		},
	}
	for _, tc := range testCases {
		inRNode, err := yaml.Parse(tc.in)
		require.NoError(t, err)

		// get list of filters
		yfilters := []interface{}{}
		for _, flt := range strings.Split(tc.filters, "\n---\n") {
			yfilters = append(yfilters, yFilter(flt))
		}

		outRNode := yPipe(inRNode, yfilters)

		if tc.expectedOut != "" {
			require.NotNil(t, outRNode)
			out, err := outRNode.String()
			require.NoError(t, err)
			assert.Equal(t, tc.expectedOut[1:], out)
		}

		if tc.expectedIn != "" {
			in, err := inRNode.String()
			require.NoError(t, err)
			assert.Equal(t, tc.expectedIn[1:], in)
		}
	}
}

func TestYValue(t *testing.T) {
	testCases := []struct {
		in          string
		expectedOut interface{}
	}{
		{
			in: `
x
`,
			expectedOut: "x",
		},
		{
			in: `
kind: x
value: b
list:
- a
- b
`,
			expectedOut: map[string]interface{}{
				"kind": "x",
				"list": []interface{}{
					"a",
					"b",
				},
				"value": "b",
			},
		},
	}

	for _, tc := range testCases {
		inRNode, err := yaml.Parse(tc.in)
		require.NoError(t, err)

		out := yValue(inRNode)
		assert.Equal(t, tc.expectedOut, out)
	}
}

func TestKYFilter(t *testing.T) {
	testCases := []struct {
		in          string
		filters     string
		expectedOut string
	}{
		{
			in: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: cf1
  labels: {}
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: cf2
  labels: {}
---
apiVersion: v1
kind: Deployment
metadata:
  name: cf1
  labels: {}
`,
			filters: `
kind: PathGetter
path: ["metadata", "labels"]
---
kind: FieldSetter
name: "newlabel"
stringValue: "newvalue"
`,
			expectedOut: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: cf1
  labels: {newlabel: newvalue}
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: cf2
  labels: {newlabel: newvalue}
---
apiVersion: v1
kind: Deployment
metadata:
  name: cf1
  labels: {newlabel: newvalue}
`,
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

		// get list of filters
		yfilters := []interface{}{}
		for _, flt := range strings.Split(tc.filters, "\n---\n") {
			yfilters = append(yfilters, yFilter(flt))
		}

		kfilters := []interface{}{newKYFilter(yfilters)}
		nodes := kPipe(b.Nodes, kfilters)

		// convert to string and compare with expected
		out := &bytes.Buffer{}
		err = kio.ByteWriter{Writer: out}.Write(nodes)
		require.NoError(t, err)
		assert.Equal(t, tc.expectedOut[1:], out.String())
	}
}
