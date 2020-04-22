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

	"k8s.io/client-go/tools/clientcmd/api"
	"sigs.k8s.io/yaml"
)

// Cluster contains information about how to communicate with a kubernetes cluster
type Cluster struct {
	// Complex cluster name defined by the using <cluster name>_<cluster type>)
	NameInKubeconf string `json:"clusterKubeconf"`

	// KubeConfig Cluster Object
	cluster *api.Cluster

	// Management configuration which will be used for all hosts in the cluster
	ManagementConfiguration string `json:"managementConfiguration"`

	// Bootstrap configuration this clusters ephemeral hosts will rely on
	Bootstrap string `json:"bootstrapInfo"`
}

// ClusterPurpose encapsulates the Cluster Type as an enumeration
type ClusterPurpose struct {
	// Cluster map of referenceable names to cluster configs
	ClusterTypes map[string]*Cluster `json:"clusterType"`
}

// ClusterComplexName holds the complex cluster name information
// Encapsulates the different operations around using it.
type ClusterComplexName struct {
	Name string
	Type string
}

func (c *Cluster) String() string {
	cyaml, err := yaml.Marshal(&c)
	if err != nil {
		return ""
	}
	kcluster := c.KubeCluster()
	kyaml, err := yaml.Marshal(&kcluster)
	if err != nil {
		return string(cyaml)
	}

	return fmt.Sprintf("%s\n%s", string(cyaml), string(kyaml))
}

func (c *Cluster) PrettyString() string {
	clusterName := NewClusterComplexNameFromKubeClusterName(c.NameInKubeconf)
	return fmt.Sprintf("Cluster: %s\n%s:\n%s", clusterName.Name, clusterName.Type, c)
}

func (c *Cluster) KubeCluster() *api.Cluster {
	return c.cluster
}
func (c *Cluster) SetKubeCluster(kc *api.Cluster) {
	c.cluster = kc
}

func (c *ClusterComplexName) String() string {
	return strings.Join([]string{c.Name, c.Type}, AirshipClusterNameSeparator)
}

func ValidClusterType(clusterType string) error {
	for _, validType := range AllClusterTypes {
		if clusterType == validType {
			return nil
		}
	}
	return fmt.Errorf("cluster type must be one of %v", AllClusterTypes)
}

// NewClusterPurpose is a convenience function that returns a new ClusterPurpose
func NewClusterPurpose() *ClusterPurpose {
	return &ClusterPurpose{
		ClusterTypes: make(map[string]*Cluster),
	}
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
