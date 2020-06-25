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

package document

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"

	"sigs.k8s.io/kustomize/api/resid"
	"sigs.k8s.io/kustomize/api/types"
)

// Selector provides abstraction layer in front of kustomize selector
type Selector struct {
	types.Selector
}

// NewSelector returns instance of Selector container
func NewSelector() Selector {
	return Selector{}
}

// Following set of functions allows to build selector object
// by name, gvk, label selector and annotation selector

// ByName select by name
func (s Selector) ByName(name string) Selector {
	s.Name = name
	return s
}

// ByNamespace select by namepace
func (s Selector) ByNamespace(namespace string) Selector {
	s.Namespace = namespace
	return s
}

// ByGvk select by gvk
func (s Selector) ByGvk(group, version, kind string) Selector {
	s.Gvk = resid.Gvk{Group: group, Version: version, Kind: kind}
	return s
}

// ByKind select by Kind
func (s Selector) ByKind(kind string) Selector {
	s.Gvk = resid.Gvk{Kind: kind}
	return s
}

// ByLabel select by label selector
func (s Selector) ByLabel(labelSelector string) Selector {
	if s.LabelSelector != "" {
		s.LabelSelector = strings.Join([]string{s.LabelSelector, labelSelector}, ",")
	} else {
		s.LabelSelector = labelSelector
	}
	return s
}

// ByAnnotation select by annotation selector.
func (s Selector) ByAnnotation(annotationSelector string) Selector {
	if s.AnnotationSelector != "" {
		s.AnnotationSelector = strings.Join([]string{s.AnnotationSelector, annotationSelector}, ",")
	} else {
		s.AnnotationSelector = annotationSelector
	}
	return s
}

// ByObject select by runtime object defined in API schema
func (s Selector) ByObject(obj runtime.Object, scheme *runtime.Scheme) (Selector, error) {
	gvks, _, err := scheme.ObjectKinds(obj)
	if err != nil {
		return Selector{}, err
	}

	if len(gvks) != 1 {
		return Selector{}, ErrRuntimeObjectKind{Obj: obj}
	}
	result := NewSelector().ByGvk(gvks[0].Group, gvks[0].Version, gvks[0].Kind)

	accessor, err := meta.Accessor(obj)
	if err != nil {
		return Selector{}, err
	}
	if name := accessor.GetName(); name != "" {
		result = result.ByName(name)
	}
	return result, nil
}

// String is a convenience function which dumps all relevant information about a Selector in the following format:
// [Key1=Value1, Key2=Value2, ...]
func (s Selector) String() string {
	var components []string
	if s.Group != "" {
		components = append(components, fmt.Sprintf("Group=%q", s.Group))
	}
	if s.Version != "" {
		components = append(components, fmt.Sprintf("Version=%q", s.Version))
	}
	if s.Kind != "" {
		components = append(components, fmt.Sprintf("Kind=%q", s.Kind))
	}
	if s.Namespace != "" {
		components = append(components, fmt.Sprintf("Namespace=%q", s.Namespace))
	}
	if s.Name != "" {
		components = append(components, fmt.Sprintf("Name=%q", s.Name))
	}
	if s.AnnotationSelector != "" {
		components = append(components, fmt.Sprintf("Annotations=%q", s.AnnotationSelector))
	}
	if s.LabelSelector != "" {
		components = append(components, fmt.Sprintf("Labels=%q", s.LabelSelector))
	}

	if len(components) == 0 {
		return "No selection conditions specified"
	}

	return fmt.Sprintf("[%s]", strings.Join(components, ", "))
}

// NewEphemeralCloudDataSelector returns selector to get BaremetalHost for ephemeral node
func NewEphemeralCloudDataSelector() Selector {
	return NewSelector().ByKind(SecretKind).ByLabel(EphemeralUserDataSelector)
}

// NewEphemeralBMHSelector returns selector to get BaremetalHost for ephemeral node
func NewEphemeralBMHSelector() Selector {
	return NewSelector().ByKind(BareMetalHostKind).ByLabel(EphemeralHostSelector)
}

// NewBMCCredentialsSelector returns selector to get BaremetalHost BMC credentials
func NewBMCCredentialsSelector(name string) Selector {
	return NewSelector().ByKind(SecretKind).ByName(name)
}

// NewNetworkDataSelector returns selector that can be used to get secret with
// network data bmhDoc argument is a document interface, that should hold fields
// spec.networkData.name and spec.networkData.namespace where to find the secret,
// if either of these fields are not defined in Document error will be returned
func NewNetworkDataSelector(bmhDoc Document) (Selector, error) {
	selector := NewSelector()
	// extract the network data document pointer from the bmh document
	netConfDocName, err := bmhDoc.GetString("spec.networkData.name")
	if err != nil {
		return selector, err
	}
	netConfDocNamespace, err := bmhDoc.GetString("spec.networkData.namespace")
	if err != nil {
		return selector, err
	}

	// try and find these documents in our bundle
	selector = selector.
		ByKind(SecretKind).
		ByNamespace(netConfDocNamespace).
		ByName(netConfDocName)

	return selector, nil
}

// NewDeployToK8sSelector returns a selector to get documents that are to be deployed
// to kubernetes cluster.
func NewDeployToK8sSelector() Selector {
	return NewSelector().ByLabel(DeployToK8sSelector)
}

// NewClusterctlMetadataSelector returns selector to get clusterctl metadata documents
func NewClusterctlMetadataSelector() Selector {
	return NewSelector().ByGvk(ClusterctlMetadataGroup,
		ClusterctlMetadataVersion,
		ClusterctlMetadataKind)
}
