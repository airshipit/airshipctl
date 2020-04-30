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

package apply_test

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"opendev.org/airship/airshipctl/pkg/environment"
	"opendev.org/airship/airshipctl/pkg/k8s/client"
	"opendev.org/airship/airshipctl/pkg/k8s/client/fake"
	"opendev.org/airship/airshipctl/pkg/k8s/kubectl"
	"opendev.org/airship/airshipctl/pkg/phase/apply"
	"opendev.org/airship/airshipctl/testutil/k8sutils"
)

const (
	kubeconfigPath    = "testdata/kubeconfig.yaml"
	filenameRC        = "testdata/primary/site/test-site/ephemeral/initinfra/replicationcontroller.yaml"
	airshipConfigFile = "testdata/config.yaml"
)

var (
	DynamicClientError = errors.New("DynamicClientError")
)

func TestDeploy(t *testing.T) {
	rs := makeNewFakeRootSettings(t, kubeconfigPath, airshipConfigFile)
	tf := k8sutils.NewFakeFactoryForRC(t, filenameRC)
	defer tf.Cleanup()

	ao := apply.NewOptions(rs)
	ao.PhaseName = "initinfra"
	ao.DryRun = true

	kctl := kubectl.NewKubectl(tf)

	tests := []struct {
		theApplyOptions *apply.Options
		client          client.Interface
		prune           bool
		expectedError   error
	}{
		{

			client: fake.NewClient(fake.WithKubectl(
				kubectl.NewKubectl(k8sutils.
					NewMockKubectlFactory().
					WithDynamicClientByError(nil, DynamicClientError)))),
			expectedError: DynamicClientError,
		},
		{
			expectedError: nil,
			prune:         false,
			client:        fake.NewClient(fake.WithKubectl(kctl)),
		},
		{
			expectedError: nil,
			prune:         true,
			client:        fake.NewClient(fake.WithKubectl(kctl)),
		},
	}

	for _, test := range tests {
		ao.Prune = test.prune
		ao.Client = test.client
		actualErr := ao.Run()
		assert.Equal(t, test.expectedError, actualErr)
	}
}

// makeNewFakeRootSettings takes kubeconfig path and directory path to fixture dir as argument.
func makeNewFakeRootSettings(t *testing.T, kp string, dir string) *environment.AirshipCTLSettings {
	t.Helper()

	akp, err := filepath.Abs(kp)
	require.NoError(t, err)

	adir, err := filepath.Abs(dir)
	require.NoError(t, err)

	settings := &environment.AirshipCTLSettings{
		AirshipConfigPath: adir,
		KubeConfigPath:    akp,
	}

	settings.InitConfig()
	return settings
}
