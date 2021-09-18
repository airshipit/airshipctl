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

package templater

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"sigs.k8s.io/kustomize/kyaml/kio"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
	"sigs.k8s.io/yaml"

	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
)

func TestTemplater(t *testing.T) {
	testCases := []struct {
		in          string
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
			expectedOut: `apiVersion: metal3.io/v1alpha1
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
    someKey:
      anotherKey: value
    of:
      - toYaml
template: |
  {{ toYaml . -}}
`,
			expectedOut: `test:
  of:
  - toYaml
  someKey:
    anotherKey: value
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
    someKey:
      anotherKey: value
    of:
      - toYaml
template: |
  {{- $_ := setItems getItems -}}
  {{ toYaml . -}}
`,
			expectedOut: `test:
  of:
  - toYaml
  someKey:
    anotherKey: value
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
			expectedErr: `template: notImportantHere:1: function "ignorethisbadinput" not defined`,
		},
		{
			cfg: `
apiVersion: airshipit.org/v1alpha1
kind: Templater
metadata:
  name: notImportantHere
template: |
  {{ end }`,
			expectedErr: "template: notImportantHere:1: unexpected \"}\" in end",
		},
		{
			cfg: `
apiVersion: airshipit.org/v1alpha1
kind: Templater
metadata:
  name: notImportantHere
values:
template: |
  touint32: {{ toUint32 10 -}}
`,
			expectedOut: `touint32: 10
`,
		},
		{
			cfg: `
apiVersion: airshipit.org/v1alpha1
kind: Templater
metadata:
  name: notImportantHere
values:
  regex: "^[a-z]{5,10}$"
  nomatchregex: "^[a-z]{0,4}$"
  limit: 10
template: |
  truepassword: {{ regexMatch .regex (regexGen .regex (.limit|int)) }}
  falsepassword: {{ regexMatch .nomatchregex (regexGen .regex (.limit|int)) }}
`,
			expectedOut: `truepassword: true
falsepassword: false
`,
		}, {
			cfg: `
apiVersion: airshipit.org/v1alpha1
kind: Templater
metadata:
  name: notImportantHere
values:
  name: test
  regex: "^[a-z]{5,10}$"
  limit: 0
template: |
  password: {{ regexGen .regex (.limit|int) }}
`,
			expectedErr: "template: notImportantHere:1:13: executing \"notImportantHere\" at " +
				"<regexGen .regex (.limit | int)>: error calling regexGen: " +
				"Limit cannot be less than or equal to 0",
		},
		{
			cfg: `
apiVersion: airshipit.org/v1alpha1
kind: Templater
metadata:
  name: notImportantHere
values:
  name: test
  regex: "^[a-z"
  limit: 10
template: |
  password: {{ regexGen .regex (.limit|int) }}
`,
			expectedErr: "template: notImportantHere:1:13: executing \"notImportantHere\" " +
				"at <regexGen .regex (.limit | int)>: error calling " +
				"regexGen: error parsing regexp: missing closing ]: `[a-z`",
		},
		// transformer tests
		{
			in: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: map1
`,
			cfg: `
apiVersion: airshipit.org/v1alpha1
kind: Templater
metadata:
  name: notImportantHere
values:
  annotationTransf: |
    kind: AnnotationSetter
    key: test-annotation
    value: %s
template: |
  {{- $_ := KPipe getItems (list (KYFilter (list (YFilter (printf .annotationTransf "testenvvalue"))))) -}}
`,
			expectedOut: `apiVersion: v1
kind: ConfigMap
metadata:
  name: map1
  annotations:
    test-annotation: 'testenvvalue'
`,
		},
		{
			in: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: map1
data:
  value: value1
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: map2
data:
  value: value2
`,
			cfg: `
apiVersion: airshipit.org/v1alpha1
kind: Templater
metadata:
  name: notImportantHere
values:
  map1grep: |
    kind: GrepFilter
    path:
    - metadata
    - name
    value: ^map1$
  pathGet1: |
    kind: PathGetter
    path:
    - data
    - value
  map2grep: |
    kind: GrepFilter
    path:
    - metadata
    - name
    value: ^map2$
  map2PathGet: |
    kind: PathGetter
    path:
    - data
  fieldSet: |
    kind: FieldSetter
    name: value
    stringValue: %s
template: |
  {{- $map1 := KPipe getItems (list (KFilter .map1grep)) -}}
  {{- $map1value := YValue (YPipe (index $map1 0) (list (YFilter .pathGet1))) -}}
  {{- $kyflt := KYFilter (list (YFilter .map2PathGet) (YFilter (printf .fieldSet $map1value))) -}}
  {{- $_ := KPipe getItems (list (KFilter .map2grep) $kyflt) -}}
`,
			expectedOut: `apiVersion: v1
kind: ConfigMap
metadata:
  name: map1
data:
  value: value1
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: map2
data:
  value: value1
`,
		},
		{
			in: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: map1
  annotations:
    test-annotation: x
data:
  value: value1
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: map2
data:
  value: value2
`,
			cfg: `
apiVersion: airshipit.org/v1alpha1
kind: Templater
metadata:
  name: notImportantHere
values:
  grep: |
    kind: GrepFilter
    path:
    - metadata
    - annotations
    - test-annotation
    value: ^x$
    invertMatch: true
template: |
  {{- $_ := setItems (KPipe getItems (list (KFilter .grep))) -}}
`,
			expectedOut: `apiVersion: v1
kind: ConfigMap
metadata:
  name: map2
data:
  value: value2
`,
		},
		{
			in: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: map1
  annotations:
    test-annotation: x
data:
  value: value1
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: map2
data:
  value: value2
`,
			cfg: `
apiVersion: airshipit.org/v1alpha1
kind: Templater
metadata:
  name: notImportantHere
values:
  grep: |
    kind: GrepFilter
    path:
    - metadata
    - annotations
    - test-annotation
    value: ^x$
    invertMatch: true
template: |
  {{- $_ := setItems (KOneFilter getItems .grep) -}}
`,
			expectedOut: `apiVersion: v1
kind: ConfigMap
metadata:
  name: map2
data:
  value: value2
`,
		},
		{
			in: ``,
			cfg: `
apiVersion: airshipit.org/v1alpha1
kind: Templater
metadata:
  name: notImportantHere
template: |
  {{ define "tmplx" }}
    {{- $name:= . -}}
  apiVersion: v1
  kind: ConfigMap
  metadata:
    name: {{ $name }}
  {{ end }}
  {{ include "tmplx" "cfg1" }}
  ---
  {{ include "tmplx" "cfg2" }}
`,
			expectedOut: `apiVersion: v1
kind: ConfigMap
metadata:
  name: cfg1
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: cfg2
`,
		},
		{
			in: `
apiVersion: airshipit.org/v1alpha1
kind: Templater
metadata:
  name: libModule
template: |
  {{/* grepTpl returns yaml that can be used to built KFilter that will
       filter with grep */}}
  {{- define "grepTpl" -}}
  kind: GrepFilter
  path: {{ index . 0 }}
  value: {{ index . 1 }}
    {{ if gt (len .) 2}}
  invertMatch: {{ index . 2 }}
    {{ end }}
  {{- end -}}
  {{/* test function */}}
  {{ define "fnFromModule" }}
  apiVersion: v1
  kind: ConfigMap
  metadata:
    name: {{ index . 0 }}
  {{ end }}
`,
			cfg: `
apiVersion: airshipit.org/v1alpha1
kind: Templater
metadata:
  name: notImportantHere
template: |
  {{/* remove all modules (they already imported) */}}
  {{- $_ := setItems (KOneFilter getItems (include "grepTpl" (list "[\"kind\"]" "^Templater$" "true"))) -}}
  {{/* call fn from imported module */}}
  {{ include "fnFromModule" (list "cfg1") }}
`,
			expectedOut: `apiVersion: v1
kind: ConfigMap
metadata:
  name: cfg1
`,
		},
	}

	for _, tc := range testCases {
		cfg := make(map[string]interface{})
		err := yaml.Unmarshal([]byte(tc.cfg), &cfg)
		require.NoError(t, err)
		plugin, err := New(cfg)
		require.NoError(t, err)

		nodesIn, err := (&kio.ByteReader{Reader: bytes.NewBufferString(tc.in)}).Read()
		require.NoError(t, err)

		buf := &bytes.Buffer{}
		nodes, err := plugin.Filter(nodesIn)
		if tc.expectedErr != "" {
			assert.EqualError(t, err, tc.expectedErr)
			continue
		}
		require.NoError(t, err)
		err = kio.ByteWriter{Writer: buf}.Write(nodes)
		require.NoError(t, err)
		assert.Equal(t, tc.expectedOut, buf.String())
	}
}

func TestGenSignedCertEx(t *testing.T) {
	testCases := []struct {
		cfg             string
		expectedSubject pkix.Name
	}{
		{
			cfg: `
apiVersion: airshipit.org/v1alpha1
kind: Templater
metadata:
  name: notImportantHere
values:
  name: test
  regex: "^[a-z]{5,10}$"
  limit: 0
template: |
  {{- $targetClusterCa:=genCAEx "Kubernetes API" 3650 }}
  {{- $targetKubeconfigCert:= genSignedCertEx "/CN=admin/O=system:masters" nil nil 365 $targetClusterCa }}
  cert: {{ $targetKubeconfigCert.Cert|b64enc|quote }}
`,
			expectedSubject: pkix.Name{
				CommonName: `admin`,
				Organization: []string{
					`system:masters`,
				},
			},
		},
	}

	for _, tc := range testCases {
		cfg := make(map[string]interface{})
		err := yaml.Unmarshal([]byte(tc.cfg), &cfg)
		require.NoError(t, err)
		plugin, err := New(cfg)
		require.NoError(t, err)
		buf := &bytes.Buffer{}
		nodes, err := plugin.Filter(nil)
		require.NoError(t, err)
		err = kio.ByteWriter{Writer: buf}.Write(nodes)
		require.NoError(t, err)

		res := make(map[string]string)
		err = yaml.Unmarshal(buf.Bytes(), &res)
		require.NoError(t, err)

		key, err := base64.StdEncoding.DecodeString(res["cert"])
		require.NoError(t, err)

		der, _ := pem.Decode(key)
		if der == nil {
			t.Errorf("failed to find PEM block")
			return
		}

		cert, err := x509.ParseCertificate(der.Bytes)
		if err != nil {
			t.Errorf("failed to parse: %s", err)
			return
		}
		cert.Subject.Names = nil
		assert.Equal(t, tc.expectedSubject, cert.Subject)
	}
}

func TestGetRNodes(t *testing.T) {
	//Prepare test data A, B, C,
	//var x []*yaml.RNode
	rnode1, err := kyaml.Parse(`x: y`)
	require.NoError(t, err)
	rnode2, err := kyaml.Parse(`z: "a"`)
	require.NoError(t, err)

	testA := []*kyaml.RNode{
		rnode1,
		rnode2,
	}

	testB := []interface{}{
		rnode1,
		rnode2,
	}

	testCases := []struct {
		rnodesarr   interface{}
		expectedOut string
		expectedErr bool
	}{
		{
			rnodesarr:   nil,
			expectedErr: true,
		},
		{
			rnodesarr: testA,
			expectedOut: `
x: y
---
z: "a"
`,
		},
		{
			rnodesarr: testB,
			expectedOut: `
x: y
---
z: "a"
`,
		},
	}

	for i, tc := range testCases {
		nodes, err := getRNodes(tc.rnodesarr)
		if tc.expectedErr && err != nil {
			continue
		}
		if tc.expectedErr {
			t.Errorf("expected error, but hasn't got it for the case %d", i)
			continue
		}
		if err != nil {
			t.Errorf("got unexpected error: %v", err)
			continue
		}

		// convert to string and compare with expected
		out := &bytes.Buffer{}
		err = kio.ByteWriter{Writer: out}.Write(nodes)
		require.NoError(t, err)
		assert.Equal(t, tc.expectedOut[1:], out.String())
	}
}

func TestDebug(t *testing.T) {
	i := 0

	os.Setenv("DEBUG_TEMPLATER", "false")
	debug(func() { i = 1 })
	assert.Equal(t, 0, i)

	os.Setenv("DEBUG_TEMPLATER", "true")
	debug(func() { i = 1 })
	assert.Equal(t, 1, i)
}
