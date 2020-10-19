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

package resetsatoken_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"opendev.org/airship/airshipctl/pkg/cluster/resetsatoken"
	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/k8s/client"
	"opendev.org/airship/airshipctl/pkg/k8s/client/fake"
)

func TestRunE(t *testing.T) {
	airshipConfigPath := "testdata/airshipconfig.yaml"
	kubeConfigPath := "testdata/kubeconfig.yaml"

	tests := []struct {
		testCaseName string
		testErr      string
		resetFlags   resetsatoken.ResetFlags
		cfgFactory   config.Factory
	}{
		{
			testCaseName: "invalid config factory",
			cfgFactory: func() (*config.Config, error) {
				return nil, fmt.Errorf("test config error")
			},
			resetFlags: resetsatoken.ResetFlags{},
			testErr:    "test config error",
		},
		{
			testCaseName: "valid config factory",
			cfgFactory:   config.CreateFactory(&airshipConfigPath),
			resetFlags: resetsatoken.ResetFlags{
				SecretName: "test-secret",
				Namespace:  "test-namespace",
			},
			testErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.testCaseName, func(t *testing.T) {
			command := resetsatoken.ResetCommand{
				Options:    tt.resetFlags,
				CfgFactory: tt.cfgFactory,
			}
			err := command.RunE()
			if tt.testErr != "" {
				assert.Contains(t, err.Error(), tt.testErr)
			} else {
				fakeConfig, err := command.CfgFactory()
				assert.NoError(t, err)

				factory := client.DefaultClient
				_, err = factory(fakeConfig.LoadedConfigPath(), kubeConfigPath)
				assert.NoError(t, err)

				fakeClient := fake.NewClient()
				assert.NotEmpty(t, fakeClient)

				clientset := fakeClient.ClientSet()
				fakeManager, err := resetsatoken.NewTokenManager(clientset)
				assert.NoError(t, err)

				err = fakeManager.RotateToken(command.Options.Namespace, command.Options.SecretName)
				assert.Error(t, err)
			}
		})
	}
}
