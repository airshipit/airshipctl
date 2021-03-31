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

package kubeconfig_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/client-go/tools/clientcmd/api/v1"
	kustfs "sigs.k8s.io/kustomize/api/filesys"

	"opendev.org/airship/airshipctl/pkg/api/v1alpha1"
	"opendev.org/airship/airshipctl/pkg/clusterctl/client"
	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/fs"
	"opendev.org/airship/airshipctl/pkg/k8s/kubeconfig"
	testfs "opendev.org/airship/airshipctl/testutil/fs"
)

const (
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

var (
	errTempFile            = fmt.Errorf("tempFile Error")
	errSourceFunc          = fmt.Errorf("source func error")
	errWriter              = fmt.Errorf("writer error")
	testValidKubeconfigAPI = &v1alpha1.KubeConfig{
		Config: v1.Config{
			CurrentContext: "test",
			Clusters: []v1.NamedCluster{
				{
					Name: "some-cluster",
					Cluster: v1.Cluster{
						CertificateAuthority: "ca",
						Server:               "https://10.0.1.7:6443",
					},
				},
			},
			APIVersion: "v1",
			Contexts: []v1.NamedContext{
				{
					Name: "test",
					Context: v1.Context{
						Cluster:  "some-cluster",
						AuthInfo: "some-user",
					},
				},
			},
			AuthInfos: []v1.NamedAuthInfo{
				{
					Name: "some-user",
					AuthInfo: v1.AuthInfo{
						ClientCertificate: "cert-data",
						ClientKey:         "client-key",
					},
				},
			},
		},
	}
)

func TestKubeconfigContent(t *testing.T) {
	expectedData := []byte(testValidKubeconfig)
	fSys := fs.NewDocumentFs()
	kubeconf := kubeconfig.NewKubeConfig(
		kubeconfig.FromByte(expectedData),
		kubeconfig.InjectFileSystem(fSys),
		kubeconfig.InjectTempRoot("."))
	path, clean, err := kubeconf.GetFile()
	require.NoError(t, err)
	defer clean()
	actualData, err := fSys.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, expectedData, actualData)
}

type MockClientInterface struct {
	MockGetKubeconfig func(options *client.GetKubeconfigOptions) (string, error)
	client.Interface
}

func (c MockClientInterface) GetKubeconfig(o *client.GetKubeconfigOptions) (string, error) {
	return c.MockGetKubeconfig(o)
}

func TestFromSecret(t *testing.T) {
	tests := []struct {
		name         string
		expectedData string
		err          error
	}{
		{
			name:         "valid kubeconfig",
			expectedData: testValidKubeconfig,
			err:          nil,
		},
		{
			name:         "failed to get kubeconfig",
			expectedData: "",
			err:          errors.New("error"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			cl := MockClientInterface{
				MockGetKubeconfig: func(_ *client.GetKubeconfigOptions) (string, error) { return tt.expectedData, tt.err },
			}
			kubeconf, err := kubeconfig.FromSecret(cl, nil)()
			if tt.err != nil {
				require.Error(t, err)
				assert.Nil(t, kubeconf)
			} else {
				require.NoError(t, err)
				assert.Equal(t, []byte(tt.expectedData), kubeconf)
			}
		})
	}
}

func TestFromBundle(t *testing.T) {
	tests := []struct {
		name             string
		rootPath         string
		expectedContains string
		shouldFail       bool
	}{
		{
			name:             "valid kubeconfig",
			rootPath:         "testdata",
			shouldFail:       false,
			expectedContains: "parent_parent_context",
		},
		{
			name:       "wrong path",
			rootPath:   "wrong/path",
			shouldFail: true,
		},
		{
			name:       "kubeconfig not found",
			rootPath:   "testdata_fail",
			shouldFail: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			bundle, err := document.NewBundleByPath(tt.rootPath)
			if tt.shouldFail {
				require.Error(t, err)
				return
			}
			kubeconf, err := kubeconfig.FromBundle(bundle)()
			require.NoError(t, err)
			assert.Contains(t, string(kubeconf), tt.expectedContains)
		})
	}
}

func TestNewKubeConfig(t *testing.T) {
	tests := []struct {
		shouldPanic           bool
		name                  string
		expectedPathContains  string
		expectedErrorContains string
		src                   kubeconfig.KubeSourceFunc
		options               []kubeconfig.Option
	}{
		{
			name: "write to temp file",
			src:  kubeconfig.FromByte([]byte(testValidKubeconfig)),
			options: []kubeconfig.Option{
				kubeconfig.InjectFileSystem(
					testfs.MockFileSystem{
						MockTempFile: func(root, pattern string) (fs.File, error) {
							return testfs.TestFile{
								MockName:  func() string { return "kubeconfig-142398" },
								MockWrite: func() (int, error) { return 0, nil },
								MockClose: func() error { return nil },
							}, nil
						},
						MockRemoveAll: func() error { return nil },
					},
				),
			},
			expectedPathContains: "kubeconfig-142398",
		},
		{
			name:                 "cleanup with dump root",
			expectedPathContains: "kubeconfig-142398",
			src:                  kubeconfig.FromByte([]byte(testValidKubeconfig)),
			options: []kubeconfig.Option{
				kubeconfig.InjectTempRoot("/my-unique-root"),
				kubeconfig.InjectFileSystem(
					testfs.MockFileSystem{
						MockTempFile: func(root, _ string) (fs.File, error) {
							// check if root path is passed to the TempFile interface
							if root != "/my-unique-root" {
								return nil, errTempFile
							}
							return testfs.TestFile{
								MockName:  func() string { return "kubeconfig-142398" },
								MockWrite: func() (int, error) { return 0, nil },
								MockClose: func() error { return nil },
							}, nil
						},
						MockRemoveAll: func() error { return nil },
					},
				),
			},
		},
		{
			name: "from file, and fs option",
			src:  kubeconfig.FromFile("/my/kubeconfig", fsWithFile(t, "/my/kubeconfig")),
			options: []kubeconfig.Option{
				kubeconfig.InjectFilePath("/my/kubeconfig", fsWithFile(t, "/my/kubeconfig")),
			},
			expectedPathContains: "/my/kubeconfig",
		},
		{
			name:                 "write to real fs",
			src:                  kubeconfig.FromAPIalphaV1(testValidKubeconfigAPI),
			expectedPathContains: "kubeconfig-",
		},
		{
			name:                 "from file, use SourceFile",
			src:                  kubeconfig.FromFile("/my/kubeconfig", fsWithFile(t, "/my/kubeconfig")),
			expectedPathContains: "kubeconfig-",
		},
		{
			name:                  "temp file error",
			src:                   kubeconfig.FromAPIalphaV1(testValidKubeconfigAPI),
			expectedErrorContains: errTempFile.Error(),
			options: []kubeconfig.Option{
				kubeconfig.InjectFileSystem(
					testfs.MockFileSystem{
						MockTempFile: func(string, string) (fs.File, error) {
							return nil, errTempFile
						},
						MockRemoveAll: func() error { return nil },
					},
				),
			},
		},
		{
			name:                  "source func error",
			src:                   func() ([]byte, error) { return nil, errSourceFunc },
			expectedPathContains:  "kubeconfig-",
			expectedErrorContains: errSourceFunc.Error(),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			kubeconf := kubeconfig.NewKubeConfig(tt.src, tt.options...)
			path, clean, err := kubeconf.GetFile()
			if tt.expectedErrorContains != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrorContains)
			} else {
				require.NoError(t, err)
				actualPath := path
				assert.Contains(t, actualPath, tt.expectedPathContains)
				clean()
			}
		})
	}
}

func TestKubeConfigWrite(t *testing.T) {
	tests := []struct {
		name                  string
		expectedContent       string
		expectedErrorContains string

		readWrite io.ReadWriter
		options   []kubeconfig.Option
		src       kubeconfig.KubeSourceFunc
	}{
		{
			name:            "Basic write",
			src:             kubeconfig.FromByte([]byte(testValidKubeconfig)),
			expectedContent: testValidKubeconfig,
			readWrite:       bytes.NewBuffer([]byte{}),
		},
		{
			name:                  "Source error",
			src:                   func() ([]byte, error) { return nil, errSourceFunc },
			expectedErrorContains: errSourceFunc.Error(),
		},
		{
			name:                  "Writer error",
			src:                   kubeconfig.FromByte([]byte(testValidKubeconfig)),
			expectedErrorContains: errWriter.Error(),
			readWrite:             fakeReaderWriter{writeErr: errWriter},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			kubeconf := kubeconfig.NewKubeConfig(tt.src, tt.options...)
			err := kubeconf.Write(tt.readWrite)
			if tt.expectedErrorContains != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrorContains)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedContent, read(t, tt.readWrite))
			}
		})
	}
}

func TestKubeConfigWriteFile(t *testing.T) {
	tests := []struct {
		name                  string
		expectedContent       string
		path                  string
		expectedErrorContains string

		fs  fs.FileSystem
		src kubeconfig.KubeSourceFunc
	}{
		{
			name:            "Basic write file",
			src:             kubeconfig.FromByte([]byte(testValidKubeconfig)),
			expectedContent: testValidKubeconfig,
			fs:              fsWithFile(t, "/test-path"),
			path:            "/test-path",
		},
		{
			name:                  "Source error",
			src:                   func() ([]byte, error) { return nil, errSourceFunc },
			expectedErrorContains: errSourceFunc.Error(),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			kubeconf := kubeconfig.NewKubeConfig(tt.src, kubeconfig.InjectFileSystem(tt.fs))
			err := kubeconf.WriteFile(tt.path)
			if tt.expectedErrorContains != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrorContains)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedContent, readFile(t, tt.path, tt.fs))
			}
		})
	}
}

func readFile(t *testing.T, path string, fSys fs.FileSystem) string {
	b, err := fSys.ReadFile(path)
	require.NoError(t, err)
	return string(b)
}

func read(t *testing.T, r io.Reader) string {
	b, err := ioutil.ReadAll(r)
	require.NoError(t, err)
	return string(b)
}

func fsWithFile(t *testing.T, path string) fs.FileSystem {
	fSys := testfs.MockFileSystem{
		FileSystem: kustfs.MakeFsInMemory(),
		MockRemoveAll: func() error {
			return nil
		},
	}
	err := fSys.WriteFile(path, []byte(testValidKubeconfig))
	require.NoError(t, err)
	return fSys
}

type fakeReaderWriter struct {
	readErr  error
	writeErr error
}

var _ io.Reader = fakeReaderWriter{}
var _ io.Writer = fakeReaderWriter{}

func (f fakeReaderWriter) Read(p []byte) (n int, err error) {
	return 0, f.readErr
}

func (f fakeReaderWriter) Write(p []byte) (n int, err error) {
	return 0, f.writeErr
}
