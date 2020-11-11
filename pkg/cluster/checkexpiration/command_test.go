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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"

	"opendev.org/airship/airshipctl/pkg/cluster/checkexpiration"
	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/k8s/client"
	"opendev.org/airship/airshipctl/pkg/k8s/client/fake"
	"opendev.org/airship/airshipctl/testutil"
)

const (
	testNotImplementedErr = "not implemented: check certificate expiration logic"
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
				Threshold:  5000,
				FormatType: "json",
				Kubeconfig: "",
			},
			testErr: testNotImplementedErr,
		},
		{
			testCaseName: "valid-input-format-yaml",
			cfgFactory: func() (*config.Config, error) {
				cfg, _ := testutil.InitConfig(t)
				return cfg, nil
			},
			checkFlags: checkexpiration.CheckFlags{
				Threshold:  5000,
				FormatType: "yaml",
			},
			testErr: testNotImplementedErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.testCaseName, func(t *testing.T) {
			var objects []runtime.Object
			// TODO (guhan) append a dummy object for testing
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
			}
			// TODO (guhan) add else part to check the actual vs expected o/p
		})
	}
}
