/*
Copyright 2014 The Kubernetes Authors.

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

package config

import (
	kubeconfig "k8s.io/client-go/tools/clientcmd/api"
)

// Where possible, json tags match the cli argument names.
// Top level config objects and all values required for proper functioning are not "omitempty".
// Any truly optional piece of config is allowed to be omitted.

// Config holds the information required by airshipctl commands
// It is somewhat a superset of what a kubeconfig looks like, we allow for this overlaps by providing
// a mechanism to consume or produce a kubeconfig into / from the airship config.
type Config struct {
	// +optional
	Kind string `json:"kind,omitempty"`

	// +optional
	APIVersion string `json:"apiVersion,omitempty"`

	// Clusters is a map of referenceable names to cluster configs
	Clusters map[string]*ClusterPurpose `json:"clusters"`

	// AuthInfos is a map of referenceable names to user configs
	AuthInfos map[string]*AuthInfo `json:"users"`

	// Contexts is a map of referenceable names to context configs
	Contexts map[string]*Context `json:"contexts"`

	// Manifests is a map of referenceable names to documents
	Manifests map[string]*Manifest `json:"manifests"`

	// CurrentContext is the name of the context that you would like to use by default
	CurrentContext string `json:"currentContext"`

	// Management configuration defines management information for all baremetal hosts in a cluster.
	ManagementConfiguration map[string]*ManagementConfiguration `json:"managementConfiguration"`

	// BootstrapInfo is the configuration for container runtime, ISO builder and remote management
	BootstrapInfo map[string]*Bootstrap `json:"bootstrapInfo"`

	// loadedConfigPath is the full path to the the location of the config
	// file from which this config was loaded
	// +not persisted in file
	loadedConfigPath string

	// kubeConfigPath is the full path to the the location of the
	// kubeconfig file associated with this airship config instance
	// +not persisted in file
	kubeConfigPath string

	// Private instance of Kube Config content as an object
	kubeConfig *kubeconfig.Config
}
