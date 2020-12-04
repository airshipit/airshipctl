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

package executors_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/kyaml/fn/runtime/runtimeutil"
	"sigs.k8s.io/kustomize/kyaml/runfn"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"

	"opendev.org/airship/airshipctl/pkg/api/v1alpha1"
	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/phase/executors"
	"opendev.org/airship/airshipctl/pkg/phase/ifc"
	yaml_util "opendev.org/airship/airshipctl/pkg/util/yaml"
)

const (
	containerExecutorDoc = `
apiVersion: airshipit.org/v1alpha1
kind: GenericContainer
metadata:
  name: generic-container
  labels:
    airshipit.org/deploy-k8s: "false"
spec:
  container:
      image: quay.io/test/image:v0.0.1
config: |
  apiVersion: airshipit.org/v1alpha1
  kind: GenericContainerValues
  object:
    executables:
    - name: test
      cmdline: /tmp/x/script.sh
      env:
      - name: var
        value: testval
      volumeMounts:
      - name: default
        mountPath: /tmp/x
    volumes:
    - name: default
      secret:
        name: test-script
        defaultMode: 0777`
	//nolint: lll
	transformedFunction = `apiVersion: airshipit.org/v1alpha1
kind: GenericContainerValues
object:
  executables:
  - name: test
    cmdline: /tmp/x/script.sh
    env:
    - name: var
      value: testval
    volumeMounts:
    - name: default
      mountPath: /tmp/x
  volumes:
  - name: default
    secret:
      name: test-script
      defaultMode: 0777
metadata:
  annotations:
    config.kubernetes.io/function: "container:\n  image: quay.io/test/image:v0.0.1\nexec:
      {}\nstarlark: {}\n"
`
	singleExecutorBundlePath = "../../container/testdata/single"
	firstDocInput            = `---
apiVersion: v1
kind: Secret
metadata:
  name: test-script
stringData:
  script.sh: |
    #!/bin/sh
    echo WORKS! $var >&2
type: Opaque`
	manyExecutorBundlePath = "../../container/testdata/many"
	secondDocInput         = `---
apiVersion: v1
kind: Secret
metadata:
  labels:
    airshipit.org/ephemeral-node: "true"
  name: master-0-bmc-secret
type: Opaque
`
)

func TestNewContainerExecutor(t *testing.T) {
	execDoc, err := document.NewDocumentFromBytes([]byte(containerExecutorDoc))
	require.NoError(t, err)
	_, err = executors.NewContainerExecutor(ifc.ExecutorConfig{
		ExecutorDocument: execDoc,
		BundleFactory:    testBundleFactory(singleExecutorBundlePath),
		Helper:           makeDefaultHelper(t, "../../container/testdata"),
	})
	require.NoError(t, err)
}

func TestSetInputSingleDocument(t *testing.T) {
	bundle, err := document.NewBundleByPath(singleExecutorBundlePath)
	require.NoError(t, err)
	execDoc, err := document.NewDocumentFromBytes([]byte(containerExecutorDoc))
	require.NoError(t, err)
	e := &executors.ContainerExecutor{
		ExecutorBundle:   bundle,
		ExecutorDocument: execDoc,

		ContConf: &v1alpha1.GenericContainer{
			Spec: runtimeutil.FunctionSpec{},
		},
		RunFns: runfn.RunFns{
			Functions: []*kyaml.RNode{},
		},
	}
	err = e.SetInput()
	require.NoError(t, err)

	// need to use kustomize here, because
	// it changes order of lines in document
	doc, err := document.NewDocumentFromBytes([]byte(firstDocInput))
	require.NoError(t, err)
	docBytes, err := doc.AsYAML()
	require.NoError(t, err)
	buf := &bytes.Buffer{}
	buf.Write([]byte(yaml_util.DashYamlSeparator))
	buf.Write(docBytes)
	buf.Write([]byte(yaml_util.DotYamlSeparator))

	assert.Equal(t, buf, e.RunFns.Input)
}

func TestSetInputManyDocuments(t *testing.T) {
	bundle, err := document.NewBundleByPath(manyExecutorBundlePath)
	require.NoError(t, err)
	execDoc, err := document.NewDocumentFromBytes([]byte(containerExecutorDoc))
	require.NoError(t, err)
	e := &executors.ContainerExecutor{
		ExecutorBundle:   bundle,
		ExecutorDocument: execDoc,

		ContConf: &v1alpha1.GenericContainer{
			Spec: runtimeutil.FunctionSpec{},
		},
		RunFns: runfn.RunFns{
			Functions: []*kyaml.RNode{},
		},
	}
	err = e.SetInput()
	require.NoError(t, err)

	// need to use kustomize here, because
	// it changes order of lines in document
	docSecond, err := document.NewDocumentFromBytes([]byte(secondDocInput))
	require.NoError(t, err)
	docSecondBytes, err := docSecond.AsYAML()
	require.NoError(t, err)

	buf := &bytes.Buffer{}
	buf.Write([]byte(yaml_util.DashYamlSeparator))
	buf.Write(docSecondBytes)
	buf.Write([]byte(yaml_util.DotYamlSeparator))

	docFirst, err := document.NewDocumentFromBytes([]byte(firstDocInput))
	require.NoError(t, err)
	docFirstBytes, err := docFirst.AsYAML()
	require.NoError(t, err)
	buf.Write([]byte(yaml_util.DashYamlSeparator))
	buf.Write(docFirstBytes)
	buf.Write([]byte(yaml_util.DotYamlSeparator))

	assert.Equal(t, buf, e.RunFns.Input)
}

func TestPrepareFunctions(t *testing.T) {
	bundle, err := document.NewBundleByPath(singleExecutorBundlePath)
	require.NoError(t, err)
	execDoc, err := document.NewDocumentFromBytes([]byte(containerExecutorDoc))
	require.NoError(t, err)
	contConf := &v1alpha1.GenericContainer{
		Spec: runtimeutil.FunctionSpec{},
	}
	err = execDoc.ToAPIObject(contConf, v1alpha1.Scheme)
	require.NoError(t, err)
	e := &executors.ContainerExecutor{
		ExecutorBundle:   bundle,
		ExecutorDocument: execDoc,

		ContConf: contConf,
		RunFns: runfn.RunFns{
			Functions: []*kyaml.RNode{},
		},
	}

	err = e.PrepareFunctions()
	require.NoError(t, err)
	strFuncs, err := e.RunFns.Functions[0].String()
	require.NoError(t, err)

	assert.Equal(t, transformedFunction, strFuncs)
}

func TestSetMounts(t *testing.T) {
	testCases := []struct {
		name        string
		targetPath  string
		in          []runtimeutil.StorageMount
		expectedOut []runtimeutil.StorageMount
	}{
		{
			name:        "Empty TargetPath and mounts",
			targetPath:  "",
			in:          nil,
			expectedOut: nil,
		},
		{
			name:       "Empty TargetPath with Src and DstPath",
			targetPath: "",
			in: []runtimeutil.StorageMount{
				{
					MountType: "bind",
					Src:       "src",
					DstPath:   "dst",
				},
			},
			expectedOut: []runtimeutil.StorageMount{
				{
					MountType: "bind",
					Src:       "src",
					DstPath:   "dst",
				},
			},
		},
		{
			name:       "Not empty TargetPath with Src and DstPath",
			targetPath: "target_path",
			in: []runtimeutil.StorageMount{
				{
					MountType: "bind",
					Src:       "src",
					DstPath:   "dst",
				},
			},
			expectedOut: []runtimeutil.StorageMount{
				{
					MountType: "bind",
					Src:       "target_path/src",
					DstPath:   "dst",
				},
			},
		},
	}

	for _, test := range testCases {
		tt := test
		t.Run(tt.name, func(t *testing.T) {
			c := executors.ContainerExecutor{
				ContConf: &v1alpha1.GenericContainer{
					Spec: runtimeutil.FunctionSpec{
						Container: runtimeutil.ContainerSpec{
							StorageMounts: tt.in,
						},
					},
				},
				TargetPath: tt.targetPath,
			}
			c.SetMounts()
			assert.Equal(t, c.RunFns.StorageMounts, tt.expectedOut)
		})
	}
}
