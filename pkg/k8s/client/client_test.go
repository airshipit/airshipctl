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

package client_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"opendev.org/airship/airshipctl/pkg/k8s/client"
	"opendev.org/airship/airshipctl/testutil"
)

const (
	kubeconfigPath   = "testdata/kubeconfig.yaml"
	airshipConfigDir = "testdata"
)

func TestNewClient(t *testing.T) {
	conf, cleanup := testutil.InitConfig(t)
	defer cleanup(t)

	akp, err := filepath.Abs(kubeconfigPath)
	require.NoError(t, err)

	adir, err := filepath.Abs(airshipConfigDir)
	require.NoError(t, err)

	conf.SetLoadedConfigPath(adir)
	conf.SetKubeConfigPath(akp)

	client, err := client.NewClient(conf)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.NotNil(t, client.ClientSet())
	assert.NotNil(t, client.DynamicClient())
	assert.NotNil(t, client.ApiextensionsClientSet())
	assert.NotNil(t, client.Kubectl())
}
