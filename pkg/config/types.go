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
	"net/url"

	"k8s.io/client-go/tools/clientcmd"
	kubeconfig "k8s.io/client-go/tools/clientcmd/api"
)

// Where possible, json tags match the cli argument names.
// Top level config objects and all values required for proper functioning are not "omitempty".
// Any truly optional piece of config is allowed to be omitted.

// Config holds the information required by airshipct commands
// It is somewhat a superset of what akubeconfig looks like, we allow for this overlaps by providing
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
	CurrentContext string `json:"current-context"`

	// Modules Section
	// Will store configuration required by the different airshipctl modules
	// Such as Bootstrap, Workflows, Document, etc
	ModulesConfig *Modules `json:"modules-config"`

	// Private LoadedConfigPath is the full path to the the location of the config file
	// from which these config was loaded
	// +not persisted in file
	loadedConfigPath string

	// Private loadedPathOptions is the full path to the the location of the kubeconfig file
	// associated with this airship config instance
	// +not persisted in file
	loadedPathOptions *clientcmd.PathOptions

	// Private instance of Kube Config content as an object
	kubeConfig *kubeconfig.Config
}

// Encapsultaes the Cluster Type as an enumeration
type ClusterPurpose struct {
	// Cluster map of referenceable names to cluster configs
	ClusterTypes map[string]*Cluster `json:"cluster-type"`
}

// Cluster contains information about how to communicate with a kubernetes cluster
type Cluster struct {
	// Complex cluster name defined by the using <cluster name>_<clustertype)
	NameInKubeconf string `json:"cluster-kubeconf"`

	// Kubeconfig Cluster Object
	kCluster *kubeconfig.Cluster

	// Boostrap configuration this clusters ephemral hosts will rely on
	Bootstrap string `json:"bootstrap-info"`
}

// Modules, generic configuration for modules
// Configuration that the Bootstrap Module would need
// Configuration that the Document Module would need
// Configuration that the Workflows Module would need
type Modules struct {
	BootstrapInfo map[string]*Bootstrap `json:"bootstrapInfo"`
}

// Context is a tuple of references to a cluster (how do I communicate with a kubernetes context),
// a user (how do I identify myself), and a namespace (what subset of resources do I want to work with)
type Context struct {
	// Context name in kubeconf
	NameInKubeconf string `json:"context-kubeconf"`

	// Manifest is the default manifest to be use with this context
	// +optional
	Manifest string `json:"manifest,omitempty"`

	// Kubeconfig Context Object
	kContext *kubeconfig.Context
}

type AuthInfo struct {
	// Empty in purpose
	// Will implement Interface to Set/Get fields from kubeconfig as needed
}

// Manifests is a tuple of references to a Manifest (how do Identify, collect ,
// find the yaml manifests that airship uses to perform its operations)
type Manifest struct {
	// Repositories is the map of repository adddressable by a name
	Repositories map[string]*Repository `json:"repositories"`

	// Local Targer path for working or home dirctory for all Manifest Cloned/Returned/Generated
	TargetPath string `json:"target-path"`
}

// Repository is a tuple that holds the information for the remote sources of manifest yaml documents.
// Information such as location, authentication info,
// as well as details of what to get such as branch, tag, commit it, etc.
type Repository struct {
	// URL for Repositor,
	Url *url.URL `json:"url"`

	// Username is the username for authentication to the repository .
	// +optional
	Username string `json:"username,omitempty"`

	// Clone To Name  Should always be relative to the setting of Manifest TargetPath.
	// Defines where ths repo will be cloned to locally.
	TargetPath string `json:"target-path"`
}

// Holds the complex cluster name information
// Encapsulates the different operations around using it.
type ClusterComplexName struct {
	clusterName string
	clusterType string
}

// Bootstrap holds configurations for bootstrap steps
type Bootstrap struct {
	// Configuration parameters for container
	Container *Container `json:"container,omitempty"`
	// Configuration parameters for ISO builder
	Builder *Builder `json:"builder,omitempty"`
}

// Container parameters
type Container struct {
	// Container volume directory binding.
	Volume string `json:"volume,omitempty"`
	// ISO generator container image URL
	Image string `json:"image,omitempty"`
	// Container Runtime Interface driver
	ContainerRuntime string `json:"containerRuntime,omitempty"`
}

// Builder parameters
type Builder struct {
	// Cloud Init user-data file name placed to the container volume root
	UserDataFileName string `json:"userDataFileName,omitempty"`
	// Cloud Init network-config file name placed to the container volume root
	NetworkConfigFileName string `json:"networkConfigFileName,omitempty"`
	// File name for output metadata
	OutputMetadataFileName string `json:"outputMetadataFileName,omitempty"`
}
