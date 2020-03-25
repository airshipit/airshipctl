package document

import (
	"strings"

	"sigs.k8s.io/kustomize/v3/pkg/gvk"
	"sigs.k8s.io/kustomize/v3/pkg/types"
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
	s.Gvk = gvk.Gvk{Group: group, Version: version, Kind: kind}
	return s
}

// ByKind select by Kind
func (s Selector) ByKind(kind string) Selector {
	s.Gvk = gvk.Gvk{Kind: kind}
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

// EphemeralCloudDataSelector returns selector to get BaremetalHost for ephemeral node
func NewEphemeralCloudDataSelector() Selector {
	return NewSelector().ByKind(SecretKind).ByLabel(EphemeralUserDataSelector)
}

// NewEphemeralBMHSelector returns selector to get BaremetalHost for ephemeral node
func NewEphemeralBMHSelector() Selector {
	return NewSelector().ByKind(BareMetalHostKind).ByLabel(EphemeralHostSelector)
}

// NewEphemeralNetworkDataSelector returns selector that can be used to get secret with
// network data bmhDoc argument is a document interface, that should hold fields
// spec.networkData.name and spec.networkData.namespace where to find the secret,
// if either of these fields are not defined in Document error will be returned
func NewEphemeralNetworkDataSelector(bmhDoc Document) (Selector, error) {
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
