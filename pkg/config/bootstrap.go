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

import "sigs.k8s.io/yaml"

// Bootstrap holds configurations for bootstrap steps
type Bootstrap struct {
	// Configuration parameters for container
	Container *Container `json:"container,omitempty"`
	// Configuration parameters for ISO builder
	Builder *Builder `json:"builder,omitempty"`
	// Configuration parameters for ephemeral node remote management
	RemoteDirect *RemoteDirect `json:"remoteDirect,omitempty"`
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

// RemoteDirect configuration options
type RemoteDirect struct {
	// IsoURL specifies url to download ISO image for ephemeral node
	IsoURL string `json:"isoUrl,omitempty"`
}

// Bootstrap functions
func (b *Bootstrap) String() string {
	yamlData, err := yaml.Marshal(&b)
	if err != nil {
		return ""
	}
	return string(yamlData)
}

// String returns Container object in a serialized string format
func (c *Container) String() string {
	yamlData, err := yaml.Marshal(&c)
	if err != nil {
		return ""
	}
	return string(yamlData)
}

// String returns Builder object in a serialized string format
func (b *Builder) String() string {
	yamlData, err := yaml.Marshal(&b)
	if err != nil {
		return ""
	}
	return string(yamlData)
}
