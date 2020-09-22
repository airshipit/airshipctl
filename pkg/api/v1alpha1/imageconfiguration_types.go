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

// Container structure contains parameters related to Docker runtime, used for building image
type Container struct {
	// Container volume directory binding.
	Volume string `json:"volume,omitempty"`
	// ISO generator container image URL
	Image string `json:"image,omitempty"`
	// Container Runtime Interface driver
	ContainerRuntime string `json:"containerRuntime,omitempty"`
}

// Builder structure defines metadata files (including Cloud Init metadata) used for image
type Builder struct {
	// Cloud Init user-data file name placed to the container volume root
	UserDataFileName string `json:"userDataFileName,omitempty"`
	// Cloud Init network-config file name placed to the container volume root
	NetworkConfigFileName string `json:"networkConfigFileName,omitempty"`
	// File name for output metadata
	OutputMetadataFileName string `json:"outputMetadataFileName,omitempty"`
}

// +kubebuilder:object:root=true

// ImageConfiguration structure is inherited from apimachinery TypeMeta and ObjectMeta and is a top level
// configuration structure for building image
type ImageConfiguration struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Container *Container `json:"container,omitempty"`
	Builder   *Builder   `json:"builder,omitempty"`
}

// DefaultImageConfiguration can be used to safely unmarshal ImageConfiguration object without nil pointers
func DefaultImageConfiguration() *ImageConfiguration {
	return &ImageConfiguration{
		Container: &Container{},
		Builder:   &Builder{},
	}
}
