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

// These types have been changed in kustomize 4.0.5.
// Copying the version from 3.9.2

// Gvk identifies a Kubernetes API type.
// https://github.com/kubernetes/community/blob/master/contributors/design-proposals/api-machinery/api-group.md
type Gvk struct {
	Group   string `json:"group,omitempty" yaml:"group,omitempty"`
	Version string `json:"version,omitempty" yaml:"version,omitempty"`
	Kind    string `json:"kind,omitempty" yaml:"kind,omitempty"`
}

// Target refers to a kubernetes object by Group, Version, Kind and Name
// gvk.Gvk contains Group, Version and Kind
// APIVersion is added to keep the backward compatibility of using ObjectReference
// for Var.ObjRef
type Target struct {
	APIVersion string `json:"apiVersion,omitempty" yaml:"apiVersion,omitempty"`
	Gvk        `json:",inline,omitempty" yaml:",inline,omitempty"`
	Name       string `json:"name" yaml:"name"`
	Namespace  string `json:"namespace,omitempty" yaml:"namespace,omitempty"`
}

// ResID is an identifier of a k8s resource object.
type ResID struct {
	// Gvk of the resource.
	Gvk `json:",inline,omitempty" yaml:",inline,omitempty"`

	// Name of the resource.
	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	// Namespace the resource belongs to, if it can belong to a namespace.
	Namespace string `json:"namespace,omitempty" yaml:"namespace,omitempty"`
}

// Selector specifies a set of resources.
// Any resource that matches intersection of all conditions
// is included in this set.
type Selector struct {
	// ResID refers to a GVKN/Ns of a resource.
	ResID `json:",inline,omitempty" yaml:",inline,omitempty"`

	// AnnotationSelector is a string that follows the label selection expression
	// https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#api
	// It matches with the resource annotations.
	AnnotationSelector string `json:"annotationSelector,omitempty" yaml:"annotationSelector,omitempty"`

	// LabelSelector is a string that follows the label selection expression
	// https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#api
	// It matches with the resource labels.
	LabelSelector string `json:"labelSelector,omitempty" yaml:"labelSelector,omitempty"`
}

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
	ObjRef   *Target `json:"objref,omitempty" yaml:"objref,omitempty"`
	FieldRef string  `json:"fieldref,omitempty" yaml:"fiedldref,omitempty"`
	Value    *string `json:"value,omitempty" yaml:"value,omitempty"`
}

// ReplTarget defines where a substitution is to.
type ReplTarget struct {
	ObjRef    *Selector `json:"objref,omitempty" yaml:"objref,omitempty"`
	FieldRefs []string  `json:"fieldrefs,omitempty" yaml:"fieldrefs,omitempty"`
}

// +kubebuilder:object:root=true

// ReplacementTransformer plugin configuration for airship document model
type ReplacementTransformer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Replacements list of source and target field to do a replacement
	Replacements []Replacement `json:"replacements,omitempty" yaml:"replacements,omitempty"`
}
