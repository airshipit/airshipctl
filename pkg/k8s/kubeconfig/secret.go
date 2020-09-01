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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"opendev.org/airship/airshipctl/pkg/k8s/client"
	"opendev.org/airship/airshipctl/pkg/log"
)

// FromClusterOptions holds all configurable options for kubeconfig extraction
type FromClusterOptions struct {
	ClusterName string
	Namespace   string
	Client      client.Interface
}

// GetKubeconfigFromSecret extracts kubeconfig from secret data structure
func GetKubeconfigFromSecret(o *FromClusterOptions) ([]byte, error) {
	const defaultNamespace = "default"

	if o.ClusterName == "" {
		return nil, ErrClusterNameEmpty{}
	}
	if o.Namespace == "" {
		log.Printf("Namespace is not provided, using default one")
		o.Namespace = defaultNamespace
	}

	log.Debugf("Extracting kubeconfig from secret in cluster %s(namespace: %s)", o.ClusterName, o.Namespace)
	secretName := fmt.Sprintf("%s-kubeconfig", o.ClusterName)
	kubeCore := o.Client.ClientSet().CoreV1()

	secret, err := kubeCore.Secrets(o.Namespace).Get(secretName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	if secret.Data == nil {
		return nil, ErrMalformedSecret{
			ClusterName: o.ClusterName,
			Namespace:   o.Namespace,
			SecretName:  secretName,
		}
	}

	val, exist := secret.Data["value"]
	if !exist || len(val) == 0 {
		return nil, ErrMalformedSecret{
			ClusterName: o.ClusterName,
			Namespace:   o.Namespace,
			SecretName:  secretName,
		}
	}

	return val, nil
}
