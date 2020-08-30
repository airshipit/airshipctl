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

package poller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"opendev.org/airship/airshipctl/pkg/cluster"
	"opendev.org/airship/airshipctl/pkg/k8s/client/fake"
	"opendev.org/airship/airshipctl/pkg/k8s/poller"
	k8sutils "opendev.org/airship/airshipctl/pkg/k8s/utils"
)

func TestNewStatusPoller(t *testing.T) {
	airClient := fake.NewClient()

	f := k8sutils.FactoryFromKubeConfigPath("testdata/kubeconfig.yaml")
	restConfig, err := f.ToRESTConfig()
	require.NoError(t, err)
	restMapper, err := f.ToRESTMapper()
	require.NoError(t, err)
	restClient, err := client.New(restConfig, client.Options{Mapper: restMapper})
	require.NoError(t, err)
	statusmap, err := cluster.NewStatusMap(airClient)
	require.NoError(t, err)

	a := poller.NewStatusPoller(restClient, restMapper, statusmap)
	assert.NotNil(t, a)
}
