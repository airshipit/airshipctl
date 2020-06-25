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
	clusterctlv1 "sigs.k8s.io/cluster-api/cmd/clusterctl/api/v1alpha3"
)

// +kubebuilder:object:root=true

// Clusterctl provides information about clusterctl components
type Clusterctl struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Providers   []*Provider  `json:"providers,omitempty"`
	InitOptions *InitOptions `json:"init-options,omitempty"`
	MoveOptions *MoveOptions `json:"move-options,omitempty"`
}

// Provider is part of clusterctl config
type Provider struct {
	Name string `json:"name,"`
	Type string `json:"type,"`
	URL  string `json:"url,omitempty"`

	// IsClusterctlRepository if set to true, clusterctl provider's repository implementation will be used
	// if omitted or set to false, airshipctl repository implementation will be used.
	IsClusterctlRepository bool `json:"clusterctl-repository,omitempty"`

	// Map of versions where each key is a version and value is path relative to target path of the manifest
	// ignored if IsClusterctlRepository is set to true
	Versions map[string]string `json:"versions,omitempty"`
}

// InitOptions container with exposed clusterctl InitOptions
type InitOptions struct {
	// CoreProvider version (e.g. cluster-api:v0.3.0) to add to the management cluster. If unspecified, the
	// cluster-api core provider's latest release is used.
	CoreProvider string `json:"core-provider,omitempty"`

	// BootstrapProviders and versions (e.g. kubeadm:v0.3.0) to add to the management cluster.
	// If unspecified, the kubeadm bootstrap provider's latest release is used.
	BootstrapProviders []string `json:"bootstrap-providers,omitempty"`

	// InfrastructureProviders and versions (e.g. aws:v0.5.0) to add to the management cluster.
	InfrastructureProviders []string `json:"infrastructure-providers,omitempty"`

	// ControlPlaneProviders and versions (e.g. kubeadm:v0.3.0) to add to the management cluster.
	// If unspecified, the kubeadm control plane provider latest release is used.
	ControlPlaneProviders []string `json:"control-plane-providers,omitempty"`
}

// Provider returns provider filtering by name and type
func (c *Clusterctl) Provider(name string, providerType clusterctlv1.ProviderType) *Provider {
	t := string(providerType)
	for _, prov := range c.Providers {
		if prov.Name == name && prov.Type == t {
			return prov
		}
	}
	return nil
}

// MoveOptions carries the options supported by move.
type MoveOptions struct {
	// The namespace where the workload cluster is hosted. If unspecified, the target context's namespace is used.
	Namespace string `json:"namespace,omitempty"`
}
