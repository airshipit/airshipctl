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
	testThreshold = 5000

	expectedJSONOutput = `[
			{
				"name": "test-cluster-etcd",
				"namespace": "default",
				"certificate": {
					"ca.crt": "2030-08-31 10:12:49 +0000 UTC",
					"tls.crt": "2030-08-31 10:12:49 +0000 UTC"
				}
			}
		]`

	expectedYAMLOutput = `
---
- certificate:
    ca.crt: 2030-08-31 10:12:49 +0000 UTC
    tls.crt: 2030-08-31 10:12:49 +0000 UTC
  name: test-cluster-etcd
  namespace: default
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
			objects := []runtime.Object{getTLSSecret(t)}
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

func getTLSSecret(t *testing.T) *v1.Secret {
	t.Helper()
	object := readObjectFromFile(t, "testdata/tls-secret.yaml")
	secret, ok := object.(*v1.Secret)
	require.True(t, ok)
	return secret
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
