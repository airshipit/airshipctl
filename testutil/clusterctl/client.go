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

package clusterctl

import (
	"fmt"

	"github.com/stretchr/testify/mock"

	"opendev.org/airship/airshipctl/pkg/clusterctl/client"
)

var _ client.Interface = &MockInterface{}

// MockInterface provides mock interface for clusterctl
type MockInterface struct {
	mock.Mock
}

// Init to be implemented
func (m *MockInterface) Init(kubeconfigPath, kubeconfigContext string) error {
	return nil
}

// Move to be implemented
func (m *MockInterface) Move(fkp, fkc, tkp, tkc, namespace string) error {
	return nil
}

// Render to be implemented
func (m *MockInterface) Render(client.RenderOptions) ([]byte, error) {
	return nil, nil
}

// GetKubeconfig allows to control exepected input to the function and check expected output
// example usage:
// c := &clusterctl.MockInterface{
// 	Mock: mock.Mock{},
// }
// c.On("GetKubeconfig").Once().Return(&client.GetKubeconfigOptions{
// 	ParentKubeconfigPath:    filepath.Join("testdata", kubeconfig.KubeconfigPrefix),
// 	ParentKubeconfigContext: "dummy_cluster",
// 	ManagedClusterNamespace: clustermap.DefaultClusterAPIObjNamespace,
// 	ManagedClusterName:      childCluster,
// }, "kubeconfig data", nil)
// first argument in return function is what you expect as input
// second argument is resulting expected string
// third is resulting error
func (m *MockInterface) GetKubeconfig(options *client.GetKubeconfigOptions) (string, error) {
	args := m.Called(options)
	expectedResult, ok := args.Get(0).(string)
	if !ok {
		return "", fmt.Errorf("wrong input")
	}
	return expectedResult, args.Error(1)
}
