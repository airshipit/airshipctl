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

package applier_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/events"
	"opendev.org/airship/airshipctl/pkg/k8s/applier"
	"opendev.org/airship/airshipctl/pkg/k8s/kubeconfig"
	"opendev.org/airship/airshipctl/pkg/k8s/utils"
	"opendev.org/airship/airshipctl/pkg/phase"
	"opendev.org/airship/airshipctl/pkg/phase/ifc"
	"opendev.org/airship/airshipctl/testutil/fs"
)

const (
	ValidExecutorDoc = `apiVersion: airshipit.org/v1alpha1
kind: KubernetesApply
metadata:
  labels:
    airshipit.org/deploy-k8s: "false"
  name: kubernetes-apply
config:
  waitOptions:
    timeout: 600
  pruneOptions:
    prune: false
`
	ValidExecutorDocNamespaced = `apiVersion: airshipit.org/v1alpha1
kind: KubernetesApply
metadata:
  labels:
    airshipit.org/deploy-k8s: "false"
  name: kubernetes-apply-namespaced
  namespace: bundle
config:
  waitOptions:
    timeout: 600
  pruneOptions:
    prune: false
`
	testValidKubeconfig = `apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: ca-data
    server: https://10.0.1.7:6443
  name: kubernetes_target
contexts:
- context:
    cluster: kubernetes_target
    user: kubernetes-admin
  name: kubernetes-admin@kubernetes
current-context: ""
kind: Config
preferences: {}
users:
- name: kubernetes-admin
  user:
    client-certificate-data: cert-data
    client-key-data: client-keydata
`
)

func TestNewExecutor(t *testing.T) {
	tests := []struct {
		name        string
		cfgDoc      string
		expectedErr string
		helper      ifc.Helper
		kubeconf    kubeconfig.Interface
		bundleFunc  func(t *testing.T) document.Bundle
	}{
		{
			name:     "valid executor",
			cfgDoc:   ValidExecutorDoc,
			kubeconf: testKubeconfig(testValidKubeconfig),
			helper:   makeDefaultHelper(t),
			bundleFunc: func(t *testing.T) document.Bundle {
				return newBundle("testdata/source_bundle", t)
			},
		},
		{
			name: "wrong config document",
			cfgDoc: `apiVersion: v1
kind: ConfigMap
metadata:
  name: first-map
  namespace: default
  labels:
    cli-utils.sigs.k8s.io/inventory-id: "some id"`,
			expectedErr: "wrong config document",
			helper:      makeDefaultHelper(t),
			bundleFunc: func(t *testing.T) document.Bundle {
				return newBundle("testdata/source_bundle", t)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			doc, err := document.NewDocumentFromBytes([]byte(tt.cfgDoc))
			require.NoError(t, err)
			require.NotNil(t, doc)

			exec, err := applier.NewExecutor(
				applier.ExecutorOptions{
					ExecutorDocument: doc,
					ExecutorBundle:   tt.bundleFunc(t),
					Kubeconfig:       tt.kubeconf,
					Helper:           tt.helper,
				})
			if tt.expectedErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "")
				assert.Nil(t, exec)
			} else {
				require.NoError(t, err)
				require.NotNil(t, exec)
			}
		})
	}
}

// TODO We need valid test that checks that actual bundle has arrived to applier
// for that we need a way to inject fake applier, which is not doable with `black box` test currently
// since we tests are in different package from executor
func TestExecutorRun(t *testing.T) {
	tests := []struct {
		name        string
		containsErr string

		kubeconf   kubeconfig.Interface
		execDoc    document.Document
		bundleFunc func(t *testing.T) document.Bundle
		helper     ifc.Helper
	}{
		{
			name:        "cant read kubeconfig error",
			containsErr: "no such file or directory",
			helper:      makeDefaultHelper(t),
			bundleFunc: func(t *testing.T) document.Bundle {
				return newBundle("testdata/source_bundle", t)
			},
			kubeconf: testKubeconfig(`invalid kubeconfig`),
			execDoc:  toKubernetesApply(t, ValidExecutorDocNamespaced),
		},
		{
			name:        "Nil bundle provided",
			execDoc:     toKubernetesApply(t, ValidExecutorDoc),
			containsErr: "nil bundle provided",
			kubeconf:    testKubeconfig(testValidKubeconfig),
			helper:      makeDefaultHelper(t),
			bundleFunc: func(t *testing.T) document.Bundle {
				return nil
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			exec, err := applier.NewExecutor(
				applier.ExecutorOptions{
					ExecutorDocument: tt.execDoc,
					Helper:           tt.helper,
					ExecutorBundle:   tt.bundleFunc(t),
					Kubeconfig:       tt.kubeconf,
				})
			if tt.name == "Nil bundle provided" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.containsErr)
			} else {
				require.NoError(t, err)
				require.NotNil(t, exec)
				ch := make(chan events.Event)
				go exec.Run(ch, ifc.RunOptions{})
				processor := events.NewDefaultProcessor(utils.Streams())
				err = processor.Process(ch)
				if tt.containsErr != "" {
					require.Error(t, err)
					assert.Contains(t, err.Error(), tt.containsErr)
				} else {
					assert.NoError(t, err)
				}
			}
		})
	}
}

func TestRender(t *testing.T) {
	execDoc, err := document.NewDocumentFromBytes([]byte(ValidExecutorDoc))
	require.NoError(t, err)
	require.NotNil(t, execDoc)
	exec, err := applier.NewExecutor(applier.ExecutorOptions{
		ExecutorBundle:   newBundle("testdata/source_bundle", t),
		ExecutorDocument: execDoc,
	})
	require.NoError(t, err)
	require.NotNil(t, exec)

	writerReader := bytes.NewBuffer([]byte{})
	err = exec.Render(writerReader, ifc.RenderOptions{})
	require.NoError(t, err)

	result := writerReader.String()
	assert.Contains(t, result, "ReplicationController")
}

func makeDefaultHelper(t *testing.T) ifc.Helper {
	t.Helper()
	conf := &config.Config{
		CurrentContext: "default",
		Contexts: map[string]*config.Context{
			"default": {
				Manifest: "default-manifest",
			},
		},
		Manifests: map[string]*config.Manifest{
			"default-manifest": {
				MetadataPath: "metadata.yaml",
				TargetPath:   "testdata",
			},
		},
	}
	helper, err := phase.NewHelper(conf)
	require.NoError(t, err)
	require.NotNil(t, helper)
	return helper
}

// toKubernetesApply converts string to document object
func toKubernetesApply(t *testing.T, s string) document.Document {
	doc, err := document.NewDocumentFromBytes([]byte(s))
	require.NoError(t, err)
	require.NotNil(t, doc)
	return doc
}

func testKubeconfig(stringData string) kubeconfig.Interface {
	return kubeconfig.NewKubeConfig(
		kubeconfig.FromByte([]byte(stringData)),
		kubeconfig.InjectFileSystem(
			fs.MockFileSystem{
				MockTempFile: func(root, pattern string) (document.File, error) {
					return fs.TestFile{
						MockName:  func() string { return "kubeconfig-142398" },
						MockWrite: func() (int, error) { return 0, nil },
						MockClose: func() error { return nil },
					}, nil
				},
				MockRemoveAll: func() error { return nil },
			},
		))
}
