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

// BaremetalManager allows execution of control operations against baremetal hosts
type BaremetalManager struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec BaremetalManagerSpec `json:"spec"`
}

// BaremetalManagerSpec holds configuration for baremetal manager
type BaremetalManagerSpec struct {
	Operation        BaremetalOperation        `json:"operation"`
	HostSelector     BaremetalHostSelector     `json:"hostSelector"`
	OperationOptions BaremetalOperationOptions `json:"operationOptions"`
	// Timeout in seconds
	Timeout int `json:"timeout"`
}

// BaremetalOperationOptions hold operation options
type BaremetalOperationOptions struct {
	RemoteDirect RemoteDirectOptions `json:"remoteDirect"`
}

// RemoteDirectOptions holds configuration for remote direct operation
type RemoteDirectOptions struct {
	ISOURL string `json:"isoURL"`
}

// BaremetalHostSelector allows to select a host by label selector, by name and namespace
type BaremetalHostSelector struct {
	LabelSelector string `json:"labelSelector"`
	Name          string `json:"name"`
	Namespace     string `json:"namespace"`
}

// BaremetalOperation defines an operation to be performed against baremetal host
type BaremetalOperation string

const (
	// BaremetalOperationReboot reboot
	BaremetalOperationReboot BaremetalOperation = "reboot"
	// BaremetalOperationPowerOff power off
	BaremetalOperationPowerOff BaremetalOperation = "power-off"
	// BaremetalOperationPowerOn power on
	BaremetalOperationPowerOn BaremetalOperation = "power-on"
	// BaremetalOperationRemoteDirect boot iso with given url
	BaremetalOperationRemoteDirect BaremetalOperation = "remote-direct"
	// BaremetalOperationEjectVirtualMedia eject virtual media
	BaremetalOperationEjectVirtualMedia BaremetalOperation = "eject-virtual-media"
)

// DefaultBaremetalManager returns BaremetalManager executor document with default values
func DefaultBaremetalManager() *BaremetalManager {
	return &BaremetalManager{Spec: BaremetalManagerSpec{Timeout: 300}}
}
