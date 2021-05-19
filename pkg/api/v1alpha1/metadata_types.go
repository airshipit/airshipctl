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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true

// ManifestMetadata defines site specific metadata like inventory and phase path
type ManifestMetadata struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Phase             PhaseSpec     `json:"phase,omitempty"`
	Inventory         InventorySpec `json:"inventory,omitempty"`
}

// PhaseSpec represents configuration for a particular phase. It contains a reference to
// the site specific manifest path and doument entry prefix
type PhaseSpec struct {
	Path                     string `json:"path"`
	DocumentEntryPointPrefix string `json:"documentEntryPointPrefix"`
}

// InventorySpec contains the path to the host inventory
type InventorySpec struct {
	Path string `json:"path"`
}

// DefaultManifestMetadata can be used to safely unmarshal phase object without nil pointers
func DefaultManifestMetadata() *ManifestMetadata {
	return &ManifestMetadata{
		Phase:     PhaseSpec{},
		Inventory: InventorySpec{},
	}
}
