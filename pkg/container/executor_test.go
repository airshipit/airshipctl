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

package container_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/kustomize/kyaml/fn/runtime/runtimeutil"
	"sigs.k8s.io/kustomize/kyaml/runfn"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"

	"opendev.org/airship/airshipctl/pkg/api/v1alpha1"
	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/container"
	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/events"
	"opendev.org/airship/airshipctl/pkg/phase"
	"opendev.org/airship/airshipctl/pkg/phase/ifc"
)

const (
	executorDoc = `
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
    config.kubernetes.io/function: "container:\n  image: quay.io/test/image:v0.0.1\n
      \ network: {}\nexec: {}\nstarlark: {}\n"
`
	singleExecutorBundlePath = "testdata/single"
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
	manyExecutorBundlePath = "testdata/many"
	secondDocInput         = `---
apiVersion: v1
kind: Secret
metadata:
  labels:
    airshipit.org/ephemeral-node: "true"
  name: master-0-bmc-secret
type: Opaque
`
	yamlSeparator = "---\n"
)

func TestRegisterExecutor(t *testing.T) {
	registry := make(map[schema.GroupVersionKind]ifc.ExecutorFactory)
	expectedGVK := schema.GroupVersionKind{
		Group:   "airshipit.org",
		Version: "v1alpha1",
		Kind:    "GenericContainer",
	}
	err := container.RegisterExecutor(registry)
	require.NoError(t, err)

	_, found := registry[expectedGVK]
	assert.True(t, found)
}

func TestNewExecutor(t *testing.T) {
	execDoc, err := document.NewDocumentFromBytes([]byte(executorDoc))
	require.NoError(t, err)
	_, err = container.NewExecutor(ifc.ExecutorConfig{
		ExecutorDocument: execDoc,
		BundleFactory:    testBundleFactory(singleExecutorBundlePath),
		Helper:           makeDefaultHelper(t),
	})
	require.NoError(t, err)
}

func TestSetInputSingleDocument(t *testing.T) {
	bundle, err := document.NewBundleByPath(singleExecutorBundlePath)
	require.NoError(t, err)
	execDoc, err := document.NewDocumentFromBytes([]byte(executorDoc))
	require.NoError(t, err)
	e := &container.Executor{
		ExecutorBundle:   bundle,
		ExecutorDocument: execDoc,

		ContConf: &v1alpha1.GenericContainer{
			Spec: runtimeutil.FunctionSpec{},
		},
		RunFns: runfn.RunFns{
			Functions: []*kyaml.RNode{},
		},
	}
	ch := make(chan events.Event)
	e.SetInput(ch)
	assert.Empty(t, ch)

	// need to use kustomize here, because
	// it changes order of lines in document
	doc, err := document.NewDocumentFromBytes([]byte(firstDocInput))
	require.NoError(t, err)
	docBytes, err := doc.AsYAML()
	require.NoError(t, err)
	docBytes = append([]byte(yamlSeparator), docBytes...)

	assert.Equal(t, bytes.NewReader(docBytes), e.RunFns.Input)
}

func TestSetInputManyDocuments(t *testing.T) {
	bundle, err := document.NewBundleByPath(manyExecutorBundlePath)
	require.NoError(t, err)
	execDoc, err := document.NewDocumentFromBytes([]byte(executorDoc))
	require.NoError(t, err)
	e := &container.Executor{
		ExecutorBundle:   bundle,
		ExecutorDocument: execDoc,

		ContConf: &v1alpha1.GenericContainer{
			Spec: runtimeutil.FunctionSpec{},
		},
		RunFns: runfn.RunFns{
			Functions: []*kyaml.RNode{},
		},
	}
	ch := make(chan events.Event)
	e.SetInput(ch)
	assert.Empty(t, ch)

	// need to use kustomize here, because
	// it changes order of lines in document
	docSecond, err := document.NewDocumentFromBytes([]byte(secondDocInput))
	require.NoError(t, err)
	docSecondBytes, err := docSecond.AsYAML()
	require.NoError(t, err)
	docBytes := append([]byte(yamlSeparator), docSecondBytes...)

	docFirst, err := document.NewDocumentFromBytes([]byte(firstDocInput))
	require.NoError(t, err)
	docFirstBytes, err := docFirst.AsYAML()
	require.NoError(t, err)
	docBytes = append(docBytes, []byte(yamlSeparator)...)
	docBytes = append(docBytes, docFirstBytes...)

	assert.Equal(t, bytes.NewReader(docBytes), e.RunFns.Input)
}

func TestPrepareFunctions(t *testing.T) {
	bundle, err := document.NewBundleByPath(singleExecutorBundlePath)
	require.NoError(t, err)
	execDoc, err := document.NewDocumentFromBytes([]byte(executorDoc))
	require.NoError(t, err)
	contConf := &v1alpha1.GenericContainer{
		Spec: runtimeutil.FunctionSpec{},
	}
	err = execDoc.ToAPIObject(contConf, v1alpha1.Scheme)
	require.NoError(t, err)
	e := &container.Executor{
		ExecutorBundle:   bundle,
		ExecutorDocument: execDoc,

		ContConf: contConf,
		RunFns: runfn.RunFns{
			Functions: []*kyaml.RNode{},
		},
	}

	ch := make(chan events.Event)
	e.PrepareFunctions(ch)
	assert.Empty(t, ch)
	strFuncs, err := e.RunFns.Functions[0].String()
	require.NoError(t, err)

	assert.Equal(t, transformedFunction, strFuncs)
}

func testBundleFactory(path string) document.BundleFactoryFunc {
	return func() (document.Bundle, error) {
		return document.NewBundleByPath(path)
	}
}

func makeDefaultHelper(t *testing.T) ifc.Helper {
	t.Helper()
	cfg := config.NewConfig()
	cfg.Manifests[config.AirshipDefaultManifest].TargetPath = "./testdata"
	cfg.Manifests[config.AirshipDefaultManifest].MetadataPath = "metadata.yaml"
	cfg.Manifests[config.AirshipDefaultManifest].Repositories[config.DefaultTestPhaseRepo].URLString = ""
	cfg.SetLoadedConfigPath(".")
	helper, err := phase.NewHelper(cfg)
	require.NoError(t, err)
	require.NotNil(t, helper)
	return helper
}
