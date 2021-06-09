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

package checkexpiration_test

import (
	"bytes"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/kubectl/pkg/scheme"

	"opendev.org/airship/airshipctl/pkg/cluster/checkexpiration"
	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/k8s/client"
	"opendev.org/airship/airshipctl/pkg/k8s/client/fake"
	"opendev.org/airship/airshipctl/testutil"
)

const (
	testThreshold = 7200

	nodeFile      = "testdata/node.yaml"
	kubeconfFile  = "testdata/kubeconfig.yaml"
	tlsSecretFile = "testdata/tls-secret.yaml" //nolint:gosec

	expectedJSONOutput = ` {
		"tlsSecrets": [
			{
				"name": "test-cluster-etcd",
				"namespace": "target-infra",
				"certificate": {
					"ca.crt": "2030-08-31 10:12:49 +0000 UTC",
					"tls.crt": "2030-08-31 10:12:49 +0000 UTC"
				}
			}
		],
		"kubeconfs": [
			{
				"secretName": "test-cluster-kubeconfig",
				"secretNamespace": "target-infra",
				"cluster": [
					{
						"name": "workload-cluster",
						"certificateName": "CertificateAuthorityData",
						"expirationDate": "2030-08-31 10:12:48 +0000 UTC"
					}
				],
				"user": [
					{
						"name": "workload-cluster-admin",
						"certificateName": "ClientCertificateData",
						"expirationDate": "2021-09-02 10:12:50 +0000 UTC"
					}
				]
			}
		],
		"nodeCerts": [
					{
						"name": "test-node",
						"certificate": {
							"admin.conf": "2021-08-06 12:36:00 +0000 UTC",
							"apiserver": "2021-08-06 12:36:00 +0000 UTC",
							"apiserver-etcd-client": "2021-08-06 12:36:00 +0000 UTC",
							"apiserver-kubelet-client": "2021-08-06 12:36:00 +0000 UTC",
							"ca": "2021-08-04 12:36:00 +0000 UTC",
							"controller-manager.conf": "2021-08-06 12:36:00 +0000 UTC",
							"etcd-ca": "2021-08-04 12:36:00 +0000 UTC",
							"etcd-healthcheck-client": "2021-08-06 12:36:00 +0000 UTC",
							"etcd-peer": "2021-08-06 12:36:00 +0000 UTC",
							"etcd-server": "2021-08-06 12:36:00 +0000 UTC",
							"front-proxy-ca": "2021-08-04 12:36:00 +0000 UTC",
							"front-proxy-client": "2021-08-06 12:36:00 +0000 UTC",
							"scheduler.conf": "2021-08-06 12:36:00 +0000 UTC"
					}
				}
			]
	}`

	expectedYAMLOutput = `
---
kubeconfs:
- cluster:
  - certificateName: CertificateAuthorityData
    expirationDate: 2030-08-31 10:12:48 +0000 UTC
    name: workload-cluster
  secretName: test-cluster-kubeconfig
  secretNamespace: target-infra
  user:
  - certificateName: ClientCertificateData
    expirationDate: 2021-09-02 10:12:50 +0000 UTC
    name: workload-cluster-admin
tlsSecrets:
- certificate:
    ca.crt: 2030-08-31 10:12:49 +0000 UTC
    tls.crt: 2030-08-31 10:12:49 +0000 UTC
  name: test-cluster-etcd
  namespace: target-infra
nodeCerts:
- name: test-node
  certificate:
    admin.conf: 2021-08-06 12:36:00 +0000 UTC
    apiserver: 2021-08-06 12:36:00 +0000 UTC
    apiserver-etcd-client: 2021-08-06 12:36:00 +0000 UTC
    apiserver-kubelet-client: 2021-08-06 12:36:00 +0000 UTC
    ca: 2021-08-04 12:36:00 +0000 UTC
    controller-manager.conf: 2021-08-06 12:36:00 +0000 UTC
    etcd-ca: 2021-08-04 12:36:00 +0000 UTC
    etcd-healthcheck-client: 2021-08-06 12:36:00 +0000 UTC
    etcd-peer: 2021-08-06 12:36:00 +0000 UTC
    etcd-server: 2021-08-06 12:36:00 +0000 UTC
    front-proxy-ca: 2021-08-04 12:36:00 +0000 UTC
    front-proxy-client: 2021-08-06 12:36:00 +0000 UTC
    scheduler.conf: 2021-08-06 12:36:00 +0000 UTC
...
`
)

func TestRunE(t *testing.T) {
	tests := []struct {
		testCaseName   string
		testErr        string
		checkFlags     checkexpiration.CheckFlags
		cfgFactory     config.Factory
		expectedOutput string
	}{
		{
			testCaseName: "invalid-input-format",
			cfgFactory: func() (*config.Config, error) {
				return nil, nil
			},
			checkFlags: checkexpiration.CheckFlags{
				Threshold:  0,
				FormatType: "test-yaml",
			},
			testErr: checkexpiration.ErrInvalidFormat{RequestedFormat: "test-yaml"}.Error(),
		},
		{
			testCaseName: "valid-input-format-json",
			cfgFactory: func() (*config.Config, error) {
				cfg, _ := testutil.InitConfig(t)
				return cfg, nil
			},
			checkFlags: checkexpiration.CheckFlags{
				Threshold:  testThreshold,
				FormatType: "json",
				Kubeconfig: "",
			},
			testErr:        "",
			expectedOutput: expectedJSONOutput,
		},
		{
			testCaseName: "valid-input-format-yaml",
			cfgFactory: func() (*config.Config, error) {
				cfg, _ := testutil.InitConfig(t)
				return cfg, nil
			},
			checkFlags: checkexpiration.CheckFlags{
				Threshold:  testThreshold,
				FormatType: "yaml",
			},
			testErr:        "",
			expectedOutput: expectedYAMLOutput,
		},
	}

	for _, tt := range tests {
		t.Run(tt.testCaseName, func(t *testing.T) {
			objects := []runtime.Object{
				getSecretObject(t, tlsSecretFile),
				getSecretObject(t, kubeconfFile),
				getNodeObject(t, nodeFile, "2021"),
			}
			ra := fake.WithTypedObjects(objects...)

			command := checkexpiration.CheckCommand{
				Options:    tt.checkFlags,
				CfgFactory: tt.cfgFactory,
				ClientFactory: func(_ string, _ string) (client.Interface, error) {
					return fake.NewClient(ra), nil
				},
			}

			var buffer bytes.Buffer
			err := command.RunE(&buffer)

			if tt.testErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.testErr)
			} else {
				require.NoError(t, err)
				t.Log(buffer.String())
				switch tt.checkFlags.FormatType {
				case "json":
					assert.JSONEq(t, tt.expectedOutput, buffer.String())
				case "yaml":
					assert.YAMLEq(t, tt.expectedOutput, buffer.String())
				}
			}
		})
	}
}

func getSecretObject(t *testing.T, fileName string) *v1.Secret {
	t.Helper()

	object := readObjectFromFile(t, fileName)
	secret, ok := object.(*v1.Secret)
	require.True(t, ok)

	return secret
}

func getNodeObject(t *testing.T, fileName string, expirationYear string) *v1.Node {
	t.Helper()

	object := readObjectFromFile(t, fileName)
	node, ok := object.(*v1.Node)
	require.True(t, ok)

	node.Annotations["cert-expiration"] = strings.ReplaceAll(node.Annotations["cert-expiration"],
		"{{year}}", expirationYear)

	return node
}

func readObjectFromFile(t *testing.T, fileName string) runtime.Object {
	t.Helper()

	contents, err := ioutil.ReadFile(fileName)
	require.NoError(t, err)

	jsonContents, err := yaml.ToJSON(contents)
	require.NoError(t, err)

	object, err := runtime.Decode(scheme.Codecs.UniversalDeserializer(), jsonContents)
	require.NoError(t, err)

	return object
}
