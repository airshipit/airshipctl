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

import (
	"fmt"
)

// ErrKubeConfigPathEmpty returned when kubeconfig path is not specified
type ErrKubeConfigPathEmpty struct {
}

func (e *ErrKubeConfigPathEmpty) Error() string {
	return "kubeconfig path is not defined"
}

// ErrClusterNameEmpty returned when cluster name is not provided
type ErrClusterNameEmpty struct {
}

func (e ErrClusterNameEmpty) Error() string {
	return "cluster name is not defined"
}

// ErrMalformedSecret error returned if secret data value is lost or empty
type ErrMalformedSecret struct {
	ClusterName string
	Namespace   string
	SecretName  string
}

func (e ErrMalformedSecret) Error() string {
	return fmt.Sprintf(
		"can't retrieve data from secret %s in cluster %s(namespace: %s)",
		e.SecretName,
		e.ClusterName,
		e.Namespace,
	)
}
