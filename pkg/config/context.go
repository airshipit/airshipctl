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

package config

import (
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
