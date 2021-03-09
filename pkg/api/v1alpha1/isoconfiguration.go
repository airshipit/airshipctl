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

	"opendev.org/airship/airshipctl/pkg/document"
)

// IsoContainer structure contains parameters related to Docker runtime, used for building image
type IsoContainer struct {
	// Container volume directory binding.
	Volume string `json:"volume,omitempty"`
	// ISO generator container image URL
	Image string `json:"image,omitempty"`
	// Container Runtime Interface driver
	ContainerRuntime string `json:"containerRuntime,omitempty"`
}

// Isogen structure defines document selection criteria for cloud-init metadata
type Isogen struct {
	// Cloud Init user data will be retrieved from the doc matching this object
	UserDataSelector document.Selector `json:"userDataSelector,omitempty"`
	// Cloud init user data will be retrieved from this document key
	UserDataKey string `json:"userDataKey,omitempty"`
	// Cloud Init network config will be retrieved from the doc matching this object
	NetworkConfigSelector document.Selector `json:"networkConfigSelector,omitempty"`
	// Cloud init network config will be retrieved from this document key
	NetworkConfigKey string `json:"networkConfigKey,omitempty"`
	// File name to use for the output image that will be written to the container volume root
	OutputFileName string `json:"outputFileName,omitempty"`
}

// +kubebuilder:object:root=true

// IsoConfiguration structure is inherited from apimachinery TypeMeta and ObjectMeta and is a top level
// configuration structure for building image
type IsoConfiguration struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	IsoContainer *IsoContainer `json:"container,omitempty"`
	Isogen       *Isogen       `json:"builder,omitempty"`
}

// DefaultIsoConfiguration can be used to safely unmarshal IsoConfiguration object without nil pointers
func DefaultIsoConfiguration() *IsoConfiguration {
	return &IsoConfiguration{
		IsoContainer: &IsoContainer{},
		Isogen:       &Isogen{},
	}
}
