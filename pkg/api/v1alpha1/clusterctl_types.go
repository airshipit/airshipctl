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

// Clusterctl provides information about clusterctl components
type Clusterctl struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Providers   []*Provider  `json:"providers,omitempty"`
	Action      ActionType   `json:"action,omitempty"`
	InitOptions *InitOptions `json:"init-options,omitempty"`
	MoveOptions *MoveOptions `json:"move-options,omitempty"`
	// AdditionalComponentVariables are variables that will be available to clusterctl
	// when reading provider components
	AdditionalComponentVariables map[string]string    `json:"additional-vars,omitempty"`
	ImageMetas                   map[string]ImageMeta `json:"images,omitempty"`
}

const (
	// CoreProviderType is a type reserved for Cluster API core repository.
	CoreProviderType = "CoreProvider"

	// BootstrapProviderType is the type associated with codebases that provide
	// bootstrapping capabilities.
	BootstrapProviderType = "BootstrapProvider"

	// InfrastructureProviderType is the type associated with codebases that provide
	// infrastructure capabilities.
	InfrastructureProviderType = "InfrastructureProvider"

	// ControlPlaneProviderType is the type associated with codebases that provide
	// control-plane capabilities.
	ControlPlaneProviderType = "ControlPlaneProvider"
)

// ImageMeta is part of clusterctl config
type ImageMeta struct {
	Repository string `json:"repository,omitempty"`
	Tag        string `json:"tag,omitempty"`
}

// Provider is part of clusterctl config
type Provider struct {
	Name string `json:"name,"`
	Type string `json:"type,"`
	// URL can contain remote URL of upstream Provider or relative to target path of the manifest
	URL string `json:"url,omitempty"`
}

// InitOptions container with exposed clusterctl InitOptions
type InitOptions struct {
	// CoreProvider version (e.g. cluster-api:v0.3.0) to add to the management cluster. If unspecified, the
	// cluster-api core provider's latest release is used.
	CoreProvider string `json:"core-provider,omitempty"`

	// BootstrapProviders and versions (comma separated, e.g. kubeadm:v0.3.0) to add to the management cluster.
	// If unspecified, the kubeadm bootstrap provider's latest release is used.
	BootstrapProviders string `json:"bootstrap-providers,omitempty"`

	// InfrastructureProviders and versions (comma separated, e.g. aws:v0.5.0,metal3:v0.4.0)
	// to add to the management cluster.
	InfrastructureProviders string `json:"infrastructure-providers,omitempty"`

	// ControlPlaneProviders and versions (comma separated, e.g. kubeadm:v0.3.0) to add to the management cluster.
	// If unspecified, the kubeadm control plane provider latest release is used.
	ControlPlaneProviders string `json:"control-plane-providers,omitempty"`
}

// ActionType for clusterctl
type ActionType string

// List of possible clusterctl actions
const (
	Init ActionType = "init"
	Move ActionType = "move"
)

// MoveOptions carries the options supported by move.
type MoveOptions struct {
	// Namespace where the objects describing the workload cluster exists. If unspecified, the current
	// namespace will be used.
	Namespace string `json:"namespace,omitempty"`
}

// DefaultClusterctl can be used to safely unmarshal Clusterctl object without nil pointers
func DefaultClusterctl() *Clusterctl {
	return &Clusterctl{
		InitOptions: &InitOptions{},
		MoveOptions: &MoveOptions{},
		Providers:   make([]*Provider, 0),
		ImageMetas:  make(map[string]ImageMeta),
	}
}

// ClusterctlOptions holds all necessary data to run clusterctl inside of KRM
type ClusterctlOptions struct {
	CmdOptions []string          `json:"cmd-options,omitempty"`
	Config     []byte            `json:"config,omitempty"`
	Components map[string][]byte `json:"components,omitempty"`
}

// GetKubeconfigOptions carries all the options to retrieve kubeconfig from parent cluster
type GetKubeconfigOptions struct {
	// Timeout is the maximum length of time to retrieve kubeconfig
	Timeout string
	// Namespace is the namespace in which secret is placed.
	ManagedClusterNamespace string
	// ManagedClusterName is the name of the managed cluster.
	ManagedClusterName string
}
