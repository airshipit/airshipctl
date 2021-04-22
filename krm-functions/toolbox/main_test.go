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

package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

const (
	dir          = "image"
	targetFile   = "my-script.sh"
	dataKey      = "script"
	wrongDataKey = "foobar"
	bundlePath   = "bundle.yaml"
	script       = `#!/bin/bash
echo -n 'stderr' 1>&2
echo -n 'stdout'`
	wrongScript = `#!/usr/bin/p
print("Hello world!")`
	inputString = `kind: testkind
metadata:
  name: test-name
  namespace: test-namespace
`
)

func TestCmdRun(t *testing.T) {
	tests := []struct {
		name        string
		workdir     string
		errContains string
		configMap   *v1.ConfigMap
	}{
		{
			name:    "Successful run",
			workdir: dir,
			configMap: &v1.ConfigMap{
				Data: map[string]string{
					dataKey: script,
				},
			},
		},
		{
			name:        "Wrong key in ConfigMap",
			workdir:     dir,
			errContains: "ConfigMap '/' doesnt' have specified script key 'script'",
			configMap: &v1.ConfigMap{
				Data: map[string]string{
					wrongDataKey: "",
				},
			},
		},
		{
			name:        "WorkDir that doesnt' exist",
			workdir:     "foobar",
			errContains: "open foobar/my-script.sh: no such file or directory",
			configMap: &v1.ConfigMap{
				Data: map[string]string{
					dataKey: script,
				},
			},
		},
		{
			name:        "Wrong interpreter",
			workdir:     dir,
			errContains: "fork/exec image/my-script.sh: no such file or directory",
			configMap: &v1.ConfigMap{
				Data: map[string]string{
					dataKey: wrongScript,
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			input, err := yaml.Parse(inputString)
			require.NoError(t, err)

			stderr := bytes.NewBuffer([]byte{})
			stdout := bytes.NewBuffer([]byte{})

			cmd := &ScriptRunner{
				ScriptFile:         targetFile,
				WorkDir:            tt.workdir,
				DataKey:            dataKey,
				ErrStream:          stderr,
				OutStream:          stdout,
				ResourceList:       &framework.ResourceList{Items: []*yaml.RNode{input}},
				ConfigMap:          tt.configMap,
				RenderedBundleFile: bundlePath,
			}
			err = cmd.Run()
			defer func() {
				require.NoError(t, cmd.Cleanup())
			}()

			if tt.errContains != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, "stderr", stderr.String())
				assert.Equal(t, "stdout", stdout.String())
				bundleFullPath := filepath.Join(tt.workdir, bundlePath)
				assert.FileExists(t, bundleFullPath)
				result, err := ioutil.ReadFile(filepath.Join(tt.workdir, bundlePath))
				require.NoError(t, err)
				assert.Contains(t, string(result), "testkind")
				assert.Contains(t, string(result), "test-name")
				assert.Contains(t, string(result), "test-namespace")
			}
		})
	}
}

func TestCmdRunCleanup(t *testing.T) {
	cMap := &v1.ConfigMap{
		Data: map[string]string{
			dataKey: script,
		},
	}

	input, err := yaml.Parse(inputString)
	require.NoError(t, err)

	stderr := bytes.NewBuffer([]byte{})
	stdout := bytes.NewBuffer([]byte{})

	cmd := &ScriptRunner{
		ScriptFile:         targetFile,
		WorkDir:            dir,
		DataKey:            dataKey,
		ErrStream:          stderr,
		OutStream:          stdout,
		ResourceList:       &framework.ResourceList{Items: []*yaml.RNode{input}},
		ConfigMap:          cMap,
		RenderedBundleFile: bundlePath,
	}

	require.NoError(t, cmd.Cleanup())
	err = cmd.Run()
	defer func() {
		require.NoError(t, cmd.Cleanup())
	}()
	assert.NoError(t, err)
}

func TestFilterBundle(t *testing.T) {
	rawDocs := `---
apiVersion: test/v1
kind: Pod
metadata:
  name: t1
---
apiVersion: test/v1
kind: Secret
metadata:
  name: t2
---
apiVersion: test/v1beta1
kind: Deployment
metadata:
  name: t3
`
	docs, err := (&kio.ByteReader{Reader: bytes.NewBufferString(rawDocs)}).Read()
	require.NoError(t, err)

	testCases := []struct {
		name          string
		inputDocs     []*yaml.RNode
		groupFilter   string
		versionFilter string
		kindFilter    string
		errContains   string
		expectedDocs  string
	}{
		{
			name:          "Correct documents",
			inputDocs:     docs,
			groupFilter:   "test",
			versionFilter: "v1",
			kindFilter:    "Pod",
			errContains:   "",
			expectedDocs: `apiVersion: test/v1
kind: Pod
metadata:
  name: t1
`,
		},
		{
			name:          "Empty kind",
			inputDocs:     docs,
			groupFilter:   "test",
			versionFilter: "v1beta1",
			kindFilter:    "",
			errContains:   "",
			expectedDocs: `apiVersion: test/v1beta1
kind: Deployment
metadata:
  name: t3
`,
		},
		{
			name:          "Empty all filters",
			inputDocs:     docs,
			groupFilter:   "",
			versionFilter: "",
			kindFilter:    "",
			errContains:   "",
			expectedDocs: `apiVersion: test/v1
kind: Pod
metadata:
  name: t1
---
apiVersion: test/v1
kind: Secret
metadata:
  name: t2
---
apiVersion: test/v1beta1
kind: Deployment
metadata:
  name: t3
`,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			cMap := &v1.ConfigMap{
				Data: map[string]string{
					dataKey: script,
				},
			}

			input, err := yaml.Parse(inputString)
			require.NoError(t, err)

			stderr := bytes.NewBuffer([]byte{})
			stdout := bytes.NewBuffer([]byte{})

			os.Setenv(ResourceGroupFilter, tc.groupFilter)
			defer os.Unsetenv(ResourceGroupFilter)
			os.Setenv(ResourceVersionFilter, tc.versionFilter)
			defer os.Unsetenv(ResourceVersionFilter)
			os.Setenv(ResourceKindFilter, tc.kindFilter)
			defer os.Unsetenv(ResourceKindFilter)

			cmd := &ScriptRunner{
				ScriptFile:         targetFile,
				WorkDir:            dir,
				DataKey:            dataKey,
				ErrStream:          stderr,
				OutStream:          stdout,
				ResourceList:       &framework.ResourceList{Items: []*yaml.RNode{input}},
				ConfigMap:          cMap,
				RenderedBundleFile: bundlePath,
			}

			filteredDocs, err := cmd.FilterBundle(tc.inputDocs)
			if tc.errContains != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errContains)
			} else {
				require.NoError(t, err)

				buf := &bytes.Buffer{}
				err = kio.ByteWriter{Writer: buf}.Write(filteredDocs)
				assert.Equal(t, tc.expectedDocs, buf.String())
				require.NoError(t, err)
			}
		})
	}
}
