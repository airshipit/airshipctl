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
	"fmt"
	"strings"

	"sigs.k8s.io/yaml"
)

// Context is a tuple of references to a cluster (how do I communicate with a kubernetes context),
// a user (how do I identify myself), and a namespace (what subset of resources do I want to work with)
type Context struct {
	// NameInKubeconf is the Context name in kubeconf
	NameInKubeconf string `json:"contextKubeconf"`

	// Manifest is the default manifest to be used with this context
	// +optional
	Manifest string `json:"manifest,omitempty"`

	// EncryptionConfig is the default encryption config to be used with this context
	// +optional
	EncryptionConfig string `json:"encryptionConfig,omitempty"`

	// Management configuration which will be used for all hosts in the cluster
	ManagementConfiguration string `json:"managementConfiguration"`
}

func (c *Context) String() string {
	cyaml, err := yaml.Marshal(&c)
	if err != nil {
		return ""
	}
	if err != nil {
		return string(cyaml)
	}
	return string(cyaml)
}

// PrettyString returns cluster name in a formatted string
func (c *Context) PrettyString() string {
	clusterName := NewClusterComplexNameFromKubeClusterName(c.NameInKubeconf)
	return fmt.Sprintf("Context: %s\n%s\n", clusterName.Name, c)
}

// ClusterType returns cluster type by extracting the type portion from
// the complex cluster name
func (c *Context) ClusterType() string {
	return NewClusterComplexNameFromKubeClusterName(c.NameInKubeconf).Type
}

// ClusterName returns cluster name by extracting the name portion from
// the complex cluster name
func (c *Context) ClusterName() string {
	return NewClusterComplexNameFromKubeClusterName(c.NameInKubeconf).Name
}

// ClusterComplexName holds the complex cluster name information
// Encapsulates the different operations around using it.
type ClusterComplexName struct {
	Name string
	Type string
}

// NewClusterComplexName returns a ClusterComplexName with the given name and type.
func NewClusterComplexName(clusterName, clusterType string) ClusterComplexName {
	return ClusterComplexName{
		Name: clusterName,
		Type: clusterType,
	}
}

// NewClusterComplexNameFromKubeClusterName takes the name of a cluster in a
// format which might be found in a kubeconfig file. This may be a simple
// string (e.g. myCluster), or it may be prepended with the type of the cluster
// (e.g. myCluster_target)
//
// If a valid cluster type was appended, the returned ClusterComplexName will
// have that type. If no cluster type is provided, the
// AirshipDefaultClusterType will be used.
func NewClusterComplexNameFromKubeClusterName(kubeClusterName string) ClusterComplexName {
	parts := strings.Split(kubeClusterName, AirshipClusterNameSeparator)

	if len(parts) == 1 {
		return NewClusterComplexName(kubeClusterName, AirshipDefaultClusterType)
	}

	// kubeClusterName matches the format myCluster_something.
	// Let's check if "something" is a clusterType.
	potentialType := parts[len(parts)-1]
	for _, ct := range AllClusterTypes {
		if potentialType == ct {
			// Rejoin the parts in the case of "my_cluster_etc_etc_<clusterType>"
			name := strings.Join(parts[:len(parts)-1], AirshipClusterNameSeparator)
			return NewClusterComplexName(name, potentialType)
		}
	}

	// "something" is not a valid clusterType, so just use the default
	return NewClusterComplexName(kubeClusterName, AirshipDefaultClusterType)
}
