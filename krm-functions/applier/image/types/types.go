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

package types

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ApplyConfig provides instructions on how to apply resources to kubernetes cluster
type ApplyConfig struct {
	WaitOptions     ApplyWaitOptions  `json:"waitOptions,omitempty"`
	PruneOptions    ApplyPruneOptions `json:"pruneOptions,omitempty"`
	Kubeconfig      string            `json:"kubeconfig,omitempty"`
	Context         string            `json:"context,omitempty"`
	DryRun          bool              `json:"druRun,omitempty"`
	Debug           bool              `json:"debug,omitempty"`
	PhaseName       string            `json:"phaseName,omitempty"`
	InventoryPolicy string            `json:"inventoryPolicy,omitempty"`
}

// ApplyWaitOptions provides instructions how to wait for kubernetes resources
type ApplyWaitOptions struct {
	// Timeout in seconds
	Timeout int `json:"timeout,omitempty"`
	// PollInterval in seconds
	PollInterval int         `json:"pollInterval,omitempty"`
	Conditions   []Condition `json:"conditions,omitempty"`
}

// ApplyPruneOptions provides instructions how to prune for kubernetes resources
type ApplyPruneOptions struct {
	Prune bool `json:"prune,omitempty"`
}

// Condition is a jsonpath for particular TypeMeta which indicates what state to wait
type Condition struct {
	metav1.TypeMeta `json:",inline"`
	JSONPath        string `json:"jsonPath,omitempty"`
	// Value is desired state to wait for, if no value specified - just existence of provided jsonPath will be checked
	Value string `json:"value,omitempty"`
}
