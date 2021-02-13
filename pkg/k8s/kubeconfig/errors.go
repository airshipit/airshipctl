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

package kubeconfig

import "fmt"

// ErrKubeConfigPathEmpty returned when kubeconfig path is not specified
type ErrKubeConfigPathEmpty struct {
}

func (e *ErrKubeConfigPathEmpty) Error() string {
	return "kubeconfig path is not defined"
}

// ErrClusterctlKubeconfigWrongContextsCount is returned when clusterctl client returns
// multiple or no contexts in kubeconfig for the child cluster
type ErrClusterctlKubeconfigWrongContextsCount struct {
	ContextCount int
}

func (e *ErrClusterctlKubeconfigWrongContextsCount) Error() string {
	return fmt.Sprintf("clusterctl client returned '%d' contexts in kubeconfig "+
		"context count must exactly one", e.ContextCount)
}
