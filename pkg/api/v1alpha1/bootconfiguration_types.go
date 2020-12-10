/*
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true

// BootConfiguration structure is inherited from apimachinery TypeMeta and ObjectMeta and is a top level
// configuration structure for the bootstrap container
type BootConfiguration struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	BootstrapContainer BootstrapContainer `json:"bootstrapContainer"`
	EphemeralCluster   EphemeralCluster   `json:"ephemeralCluster"`
}

// EphemeralCluster structure contains the data for the ephemeral cluster
type EphemeralCluster struct {
	BootstrapCommand string `json:"bootstrapCommand,omitempty"`
	ConfigFilename   string `json:"configFilename,omitempty"`
}

// BootstrapContainer structure contains the data for the bootstrap container
type BootstrapContainer struct {
	ContainerRuntime string `json:"containerRuntime,omitempty"`
	Image            string `json:"image,omitempty"`
	Volume           string `json:"volume,omitempty"`
	Kubeconfig       string `json:"saveKubeconfigFileName,omitempty"`
}

// DefaultBootConfiguration can be used to safely unmarshal BootConfiguration object without nil pointers
func DefaultBootConfiguration() *BootConfiguration {
	return &BootConfiguration{}
}
