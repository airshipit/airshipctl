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

	"k8s.io/client-go/tools/clientcmd/api"
	"sigs.k8s.io/yaml"
)

// Context is a tuple of references to a cluster (how do I communicate with a kubernetes context),
// a user (how do I identify myself), and a namespace (what subset of resources do I want to work with)
type Context struct {
	// NameInKubeconf is the Context name in kubeconf
	NameInKubeconf string `json:"contextKubeconf"`

	// Manifest is the default manifest to be use with this context
	// +optional
	Manifest string `json:"manifest,omitempty"`

	// KubeConfig Context Object
	context *api.Context
}

func (c *Context) String() string {
	cyaml, err := yaml.Marshal(&c)
	if err != nil {
		return ""
	}
	kcluster := c.KubeContext()
	kyaml, err := yaml.Marshal(&kcluster)
	if err != nil {
		return string(cyaml)
	}
	return fmt.Sprintf("%s\n%s", string(cyaml), string(kyaml))
}

// PrettyString returns cluster name in a formatted string
func (c *Context) PrettyString() string {
	clusterName := NewClusterComplexNameFromKubeClusterName(c.NameInKubeconf)
	return fmt.Sprintf("Context: %s\n%s\n", clusterName.Name, c)
}

// KubeContext returns kube context object
func (c *Context) KubeContext() *api.Context {
	return c.context
}

// SetKubeContext updates kube contect with given context details
func (c *Context) SetKubeContext(kc *api.Context) {
	c.context = kc
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
