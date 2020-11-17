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

package kubectl_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"

	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/fs"
	"opendev.org/airship/airshipctl/pkg/k8s/kubectl"
	k8sutils "opendev.org/airship/airshipctl/pkg/k8s/utils"
	"opendev.org/airship/airshipctl/testutil"
	testfs "opendev.org/airship/airshipctl/testutil/fs"
	k8stest "opendev.org/airship/airshipctl/testutil/k8sutils"
)

var (
	kubeconfigPath = "testdata/kubeconfig.yaml"
	fixtureDir     = "testdata/"

	ErrWriteOutError = errors.New("ErrWriteOutError")
	ErrTempFileError = errors.New("ErrTempFileError")
)

func TestNewKubectlFromKubeConfigPath(t *testing.T) {
	f := k8sutils.FactoryFromKubeConfig(kubeconfigPath, "")
	kctl := kubectl.NewKubectl(f).WithBufferDir("/tmp/.airship")

	assert.NotNil(t, kctl.Factory)
	assert.NotNil(t, kctl.FileSystem)
	assert.NotNil(t, kctl.IOStreams)
}

func TestApply(t *testing.T) {
	b := testutil.NewTestBundle(t, fixtureDir)
	docs, err := b.GetByAnnotation("airshipit.org/initinfra")
	require.NoError(t, err, "failed to get documents from bundle")
	replicationController, err := b.SelectOne(document.NewSelector().ByKind("ReplicationController"))
	require.NoError(t, err)
	rcBytes, err := replicationController.AsYAML()
	require.NoError(t, err)
	f := k8stest.FakeFactory(t,
		[]k8stest.ClientHandler{
			&k8stest.GenericHandler{
				Obj:       &corev1.ReplicationController{},
				Bytes:     rcBytes,
				URLPath:   "/namespaces/%s/replicationcontrollers",
				Namespace: replicationController.GetNamespace(),
			},
		})
	defer f.Cleanup()
	kctl := kubectl.NewKubectl(f).WithBufferDir("/tmp/.airship")
	kctl.Factory = f
	ao, err := kctl.ApplyOptions()
	require.NoError(t, err, "failed to get documents from bundle")
	ao.SetDryRun(true)
	tests := []struct {
		name        string
		expectedErr error
		fs          fs.FileSystem
	}{
		{
			expectedErr: nil,
			fs: testfs.MockFileSystem{
				MockRemoveAll: func() error { return nil },
				MockTempFile: func(string, string) (fs.File, error) {
					return testfs.TestFile{
						MockName:  func() string { return filenameRC },
						MockWrite: func() (int, error) { return 0, nil },
						MockClose: func() error { return nil },
					}, nil
				},
			},
		},
		{
			expectedErr: ErrWriteOutError,
			fs: testfs.MockFileSystem{
				MockTempFile: func(string, string) (fs.File, error) { return nil, ErrWriteOutError }},
		},
		{
			expectedErr: ErrTempFileError,
			fs: testfs.MockFileSystem{
				MockRemoveAll: func() error { return nil },
				MockTempFile: func(string, string) (fs.File, error) {
					return testfs.TestFile{
						MockWrite: func() (int, error) { return 0, ErrTempFileError },
						MockName:  func() string { return filenameRC },
						MockClose: func() error { return nil },
					}, nil
				}},
		},
	}
	for _, test := range tests {
		kctl.FileSystem = test.fs
		assert.Equal(t, kctl.ApplyDocs(docs, ao), test.expectedErr)
	}
}
