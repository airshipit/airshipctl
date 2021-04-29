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
	"sigs.k8s.io/kustomize/api/types"
)

// These types have been changed in kustomize 4.0.5.
// Copying the version from 3.9.2

// Replacement defines how to perform a substitution
// where it is from and where it is to.
type Replacement struct {
	Source  *ReplSource   `json:"source" yaml:"source"`
	Target  *ReplTarget   `json:"target" yaml:"target"`
	Targets []*ReplTarget `json:"targets" yaml:"targets"`
}

// ReplSource defines where a substitution is from
// It can from two different kinds of sources
//  - from a field of one resource
//  - from a string
type ReplSource struct {
	ObjRef   *types.Target `json:"objref,omitempty" yaml:"objref,omitempty"`
	FieldRef string        `json:"fieldref,omitempty" yaml:"fiedldref,omitempty"`
	Value    string        `json:"value,omitempty" yaml:"value,omitempty"`
}

// ReplTarget defines where a substitution is to.
type ReplTarget struct {
	ObjRef    *types.Selector `json:"objref,omitempty" yaml:"objref,omitempty"`
	FieldRefs []string        `json:"fieldrefs,omitempty" yaml:"fieldrefs,omitempty"`
}

// +kubebuilder:object:root=true

// ReplacementTransformer plugin configuration for airship document model
type ReplacementTransformer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Replacements list of source and target field to do a replacement
	Replacements []Replacement `json:"replacements,omitempty" yaml:"replacements,omitempty"`
}
