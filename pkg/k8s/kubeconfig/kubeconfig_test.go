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
	"fmt"
	"io"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/client-go/tools/clientcmd/api/v1"
	kustfs "sigs.k8s.io/kustomize/api/filesys"

	"opendev.org/airship/airshipctl/pkg/api/v1alpha1"
	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/k8s/kubeconfig"
	"opendev.org/airship/airshipctl/testutil/fs"
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
	//nolint: lll
	testFullValidKubeconfig = `apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUN5RENDQWJDZ0F3SUJBZ0lCQURBTkJna3Foa2lHOXcwQkFRc0ZBREFWTVJNd0VRWURWUVFERXdwcmRXSmwKY201bGRHVnpNQjRYRFRFNU1USXlOakE0TWpneU5Gb1hEVEk1TVRJeU16QTRNamd5TkZvd0ZURVRNQkVHQTFVRQpBeE1LYTNWaVpYSnVaWFJsY3pDQ0FTSXdEUVlKS29aSWh2Y05BUUVCQlFBRGdnRVBBRENDQVFvQ2dnRUJBTTFSClM0d3lnajNpU0JBZjlCR0JUS1p5VTFwYmdDaGQ2WTdJektaZWRoakM2K3k1ZEJpWm81ZUx6Z2tEc2gzOC9YQ1MKenFPS2V5cE5RcDN5QVlLdmJKSHg3ODZxSFZZNjg1ZDVYVDNaOHNyVVRzVDR5WmNzZHAzV3lHdDM0eXYzNi9BSQoxK1NlUFErdU5JemN6bzNEdWhXR0ZoQjk3VjZwRitFUTBlVWN5bk05c2hkL3AwWVFzWDR1ZlhxaENENVpzZnZUCnBka3UvTWkyWnVGUldUUUtNeGpqczV3Z2RBWnBsNnN0L2ZkbmZwd1Q5cC9WTjRuaXJnMEsxOURTSFFJTHVrU2MKb013bXNBeDJrZmxITWhPazg5S3FpMEloL2cyczRFYTRvWURZemt0Y2JRZ24wd0lqZ2dmdnVzM3pRbEczN2lwYQo4cVRzS2VmVGdkUjhnZkJDNUZNQ0F3RUFBYU1qTUNFd0RnWURWUjBQQVFIL0JBUURBZ0trTUE4R0ExVWRFd0VCCi93UUZNQU1CQWY4d0RRWUpLb1pJaHZjTkFRRUxCUUFEZ2dFQkFJek9BL00xWmRGUElzd2VoWjFuemJ0VFNURG4KRHMyVnhSV0VnclFFYzNSYmV3a1NkbTlBS3MwVGR0ZHdEbnBEL2tRYkNyS2xEeFF3RWg3NFZNSFZYYkFadDdsVwpCSm90T21xdXgxYThKYklDRTljR0FHRzFvS0g5R29jWERZY0JzOTA3ckxIdStpVzFnL0xVdG5hN1dSampqZnBLCnFGelFmOGdJUHZIM09BZ3B1RVVncUx5QU8ya0VnelZwTjZwQVJxSnZVRks2TUQ0YzFmMnlxWGxwNXhrN2dFSnIKUzQ4WmF6d0RmWUVmV3Jrdld1YWdvZ1M2SktvbjVEZ0Z1ZHhINXM2Snl6R3lPVnZ0eG1TY2FvOHNxaCs3UXkybgoyLzFVcU5ZK0hlN0x4d04rYkhwYkIxNUtIMTU5ZHNuS3BRbjRORG1jSTZrVnJ3MDVJMUg5ZGRBbGF0bz0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=
    server: https://10.23.25.101:6443
  name: dummycluster_ephemeral
contexts:
- context:
    cluster: dummycluster_ephemeral
    user: kubernetes-admin
  name: dummy_cluster
current-context: dummy_cluster
kind: Config
preferences: {}
users:
- name: kubernetes-admin
  user:
    client-certificate-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUQwRENDQXJnQ0ZFdFBveEZYSjVrVFNWTXQ0OVlqcHBQL3hCYnlNQTBHQ1NxR1NJYjNEUUVCQ3dVQU1CVXgKRXpBUkJnTlZCQU1UQ210MVltVnlibVYwWlhNd0hoY05NakF3TVRJME1Ua3hOVEV3V2hjTk1qa3hNakF5TVRreApOVEV3V2pBME1Sa3dGd1lEVlFRRERCQnJkV0psY201bGRHVnpMV0ZrYldsdU1SY3dGUVlEVlFRS0RBNXplWE4wClpXMDZiV0Z6ZEdWeWN6Q0NBaUl3RFFZSktvWklodmNOQVFFQkJRQURnZ0lQQURDQ0Fnb0NnZ0lCQU1iaFhUUmsKVjZiZXdsUjBhZlpBdTBGYWVsOXRtRThaSFEvaGtaSHhuTjc2bDZUUFltcGJvaDRvRjNGMFFqbzROS1o5NVRuWgo0OWNoV240eFJiZVlPU25EcDBpV0Qzd0pXUlZ5aVFvVUFyYTlNcHVPNkVFU1FpbFVGNXNxc0VXUVdVMjBETStBCkdxK1k0Z2c3eDJ1Q0hTdk1GUmkrNEw5RWlXR2xnRDIvb1hXUm5NWEswNExQajZPb3Vkb2Zid2RmT3J6dTBPVkUKUzR0eGtuS1BCY1BUU3YxMWVaWVhja0JEVjNPbExENEZ3dTB3NTcwcnczNzAraEpYdlZxd3Zjb2RjZjZEL1BXWQowamlnd2ppeUJuZ2dXYW04UVFjd1Nud3o0d05sV3hKOVMyWUJFb1ptdWxVUlFaWVk5ZXRBcEpBdFMzTjlUNlQ2ClovSlJRdEdhZDJmTldTYkxEck5qdU1OTGhBYWRMQnhJUHpBNXZWWk5aalJkdEMwU25pMlFUMTVpSFp4d1RxcjQKakRQQ0pYRXU3KytxcWpQVldUaUZLK3JqcVNhS1pqVWZVaUpHQkJWcm5RZkJENHNtRnNkTjB5cm9tYTZOYzRMNQpKS21RV1NHdmd1aG0zbW5sYjFRaVRZanVyZFJQRFNmdmwrQ0NHbnA1QkkvZ1pwMkF1SHMvNUpKVTJlc1ZvL0xsCkVPdHdSOXdXd3dXcTAvZjhXS3R4bVRrMTUyOUp2dFBGQXQweW1CVjhQbHZlYnVwYmJqeW5pL2xWbTJOYmV6dWUKeCtlMEpNbGtWWnFmYkRSS243SjZZSnJHWW1CUFV0QldoSVkzb1pJVTFEUXI4SUlIbkdmYlZoWlR5ME1IMkFCQQp1dlVQcUtSVk80UGkxRTF4OEE2eWVPeVRDcnB4L0pBazVyR2RBZ01CQUFFd0RRWUpLb1pJaHZjTkFRRUxCUUFECmdnRUJBSWNFM1BxZHZDTVBIMnJzMXJESk9ESHY3QWk4S01PVXZPRi90RjlqR2EvSFBJbkh3RlVFNEltbldQeDYKVUdBMlE1bjFsRDFGQlU0T0M4eElZc3VvS1VQVHk1T0t6SVNMNEZnL0lEcG54STlrTXlmNStMR043aG8rblJmawpCZkpJblVYb0tERW1neHZzSWFGd1h6bGtSTDJzL1lKYUZRRzE1Uis1YzFyckJmd2dJOFA5Tkd6aEM1cXhnSmovCm04K3hPMGhXUmJIYklrQ21NekRib2pCSWhaL00rb3VYR1doei9TakpodXhZTVBnek5MZkFGcy9PMTVaSjd3YXcKZ3ZoSGc3L2E5UzRvUCtEYytPa3VrMkV1MUZjL0E5WHpWMzc5aWhNWW5ub3RQMldWeFZ3b0ZZQUg0NUdQcDZsUApCQmwyNnkxc2JMbjl6aGZYUUJIMVpFN0EwZVE9Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K
    client-key-data: LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFcFFJQkFBS0NBUUVBdXpHSll2UFo2RG9pNDIxREs4V0phWkNuTkFkMnF6NXAvMDQyb0Z6UVBickFnekUyCnFZVmt6T0w4eEFWZVI3U041d1dvVFdFeUY4RVY3cnIvNCtIaEhHaXE1UG1xdUlGeXp6bjYvSVpjOGpVOXhFZnoKdmlrY2lyTGZVNHZSWEpRd1Z3Z0FTTmwyQVdBSWgyZERkVHJqQkNmaUtXTUh5ajBSYkhhbHNCek9wZ1QvSFR2MwpHUXpuVVF6Rkt2MmRqNVYxTmtTL0RIanlSUkorRUw2UUJZbUdzZWc1UDROYkM5ZWJ1aXBtTVRBcS9KdW1Pb29kCitGakwyblpxTDZmSTZmQnRGNU9HbHBDQjFZSjhmekN0R1FRWjdIVEliZGIydHA0M0ZWT2h5UWJWY0hxVEEwNFAKSjE1KzBXQXltVUpVejhYQTU0NHIvYnY3NEpjSlVSRmhhclppUXdJREFRQUJBb0lCQVFDU0pycjlaeVpiQ2dqegpSL3VKMFZEWCt2aVF4c01BTUZyUjJsOE1GV3NBeHk1SFA4Vk4xYmc5djN0YUVGYnI1U3hsa3lVMFJRNjNQU25DCm1uM3ZqZ3dVQWlScllnTEl5MGk0UXF5VFBOU1V4cnpTNHRxTFBjM3EvSDBnM2FrNGZ2cSsrS0JBUUlqQnloamUKbnVFc1JpMjRzT3NESlM2UDE5NGlzUC9yNEpIM1M5bFZGbkVuOGxUR2c0M1kvMFZoMXl0cnkvdDljWjR5ZUNpNwpjMHFEaTZZcXJZaFZhSW9RRW1VQjdsbHRFZkZzb3l4VDR6RTE5U3pVbkRoMmxjYTF1TzhqcmI4d2xHTzBoQ2JyClB1R1l2WFFQa3Q0VlNmalhvdGJ3d2lBNFRCVERCRzU1bHp6MmNKeS9zSS8zSHlYbEMxcTdXUmRuQVhhZ1F0VzkKOE9DZGRkb0JBb0dCQU5NcUNtSW94REtyckhZZFRxT1M1ZFN4cVMxL0NUN3ZYZ0pScXBqd2Y4WHA2WHo0KzIvTAozVXFaVDBEL3dGTkZkc1Z4eFYxMnNYMUdwMHFWZVlKRld5OVlCaHVSWGpTZ0ZEWldSY1Z1Y01sNVpPTmJsbmZGCjVKQ0xnNXFMZ1g5VTNSRnJrR3A0R241UDQxamg4TnhKVlhzZG5xWE9xNTFUK1RRT1UzdkpGQjc1QW9HQkFPTHcKalp1cnZtVkZyTHdaVGgvRDNpWll5SVV0ZUljZ2NKLzlzbTh6L0pPRmRIbFd4dGRHUFVzYVd1MnBTNEhvckFtbgpqTm4vSTluUXd3enZ3MWUzVVFPbUhMRjVBczk4VU5hbk5TQ0xNMW1yaXZHRXJ1VHFnTDM1bU41eFZPdTUxQU5JCm4yNkFtODBJT2JDeEtLa0R0ZXJSaFhHd3g5c1pONVJCbG9VRThZNGJBb0dBQ3ZsdVhMZWRxcng5VkE0bDNoNXUKVDJXRVUxYjgxZ1orcmtRc1I1S0lNWEw4cllBTElUNUpHKzFuendyN3BkaEFXZmFWdVV2SDRhamdYT0h6MUs5aQpFODNSVTNGMG9ldUg0V01PY1RwU0prWm0xZUlXcWRiaEVCb1FGdUlWTXRib1BsV0d4ZUhFRHJoOEtreGp4aThSCmdEcUQyajRwY1IzQ0g5QjJ5a0lqQjVFQ2dZRUExc0xXLys2enE1c1lNSm14K1JXZThhTXJmL3pjQnVTSU1LQWgKY0dNK0wwMG9RSHdDaUU4TVNqcVN1ajV3R214YUFuanhMb3ZwSFlRV1VmUEVaUW95UE1YQ2VhRVBLOU4xbk8xMwp0V2lHRytIZkIxaU5PazFCc0lhNFNDbndOM1FRVTFzeXBaeEgxT3hueS9LYmkvYmEvWEZ5VzNqMGFUK2YvVWxrCmJGV1ZVdWtDZ1lFQTBaMmRTTFlmTjV5eFNtYk5xMWVqZXdWd1BjRzQxR2hQclNUZEJxdHFac1doWGE3aDdLTWEKeHdvamh5SXpnTXNyK2tXODdlajhDQ2h0d21sQ1p5QU92QmdOZytncnJ1cEZLM3FOSkpKeU9YREdHckdpbzZmTQp5aXB3Q2tZVGVxRThpZ1J6UkI5QkdFUGY4eVpjMUtwdmZhUDVhM0lRZmxiV0czbGpUemNNZVZjPQotLS0tLUVORCBSU0EgUFJJVkFURSBLRVktLS0tLQo=
`
)

var (
	errTempFile            = fmt.Errorf("TempFile Error")
	errSourceFunc          = fmt.Errorf("Source func error")
	errWriter              = fmt.Errorf("Writer error")
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
	fs := document.NewDocumentFs()
	kubeconf := kubeconfig.NewKubeConfig(
		kubeconfig.FromByte(expectedData),
		kubeconfig.InjectFileSystem(fs),
		kubeconfig.InjectTempRoot("."))
	path, clean, err := kubeconf.GetFile()
	require.NoError(t, err)
	defer clean()
	actualData, err := fs.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, expectedData, actualData)
}

func TestFromBundle(t *testing.T) {
	tests := []struct {
		name         string
		rootPath     string
		shouldFail   bool
		expectedData []byte
	}{
		{
			name:         "valid kubeconfig",
			rootPath:     "testdata",
			shouldFail:   false,
			expectedData: []byte(testFullValidKubeconfig),
		},
		{
			name:         "wrong path",
			rootPath:     "wrong/path",
			shouldFail:   true,
			expectedData: nil,
		},
		{
			name:         "kubeconfig not found",
			rootPath:     "testdata_fail",
			shouldFail:   true,
			expectedData: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			kubeconf, err := kubeconfig.FromBundle(tt.rootPath)()
			if tt.shouldFail {
				require.Error(t, err)
				assert.Nil(t, kubeconf)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedData, kubeconf)
			}
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
					fs.MockFileSystem{
						MockTempFile: func(root, _ string) (document.File, error) {
							// check if root path is passed to the TempFile interface
							if root != "/my-unique-root" {
								return nil, errTempFile
							}
							return fs.TestFile{
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
					fs.MockFileSystem{
						MockTempFile: func(string, string) (document.File, error) {
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

		fs  document.FileSystem
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

func readFile(t *testing.T, path string, fs document.FileSystem) string {
	b, err := fs.ReadFile(path)
	require.NoError(t, err)
	return string(b)
}

func read(t *testing.T, r io.Reader) string {
	b, err := ioutil.ReadAll(r)
	require.NoError(t, err)
	return string(b)
}

func fsWithFile(t *testing.T, path string) document.FileSystem {
	fSys := fs.MockFileSystem{
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
