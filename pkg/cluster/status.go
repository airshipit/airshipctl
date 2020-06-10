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

package cluster

import (
	"context"
	"encoding/json"
	"fmt"

	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/cli-utils/pkg/kstatus/polling/event"
	"sigs.k8s.io/cli-utils/pkg/kstatus/status"
	"sigs.k8s.io/cli-utils/pkg/object"

	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/k8s/client"
)

// StatusMap holds a mapping of schema.GroupVersionResource to various statuses
// a resource may be in, as well as the Expression used to check for that
// status.
type StatusMap struct {
	client     client.Interface
	GkMapping  []schema.GroupKind
	mapping    map[schema.GroupVersionResource]map[status.Status]Expression
	restMapper *meta.DefaultRESTMapper
}

// NewStatusMap creates a cluster-wide StatusMap. It iterates over all
// CustomResourceDefinitions in the cluster that are annotated with the
// airshipit.org/status-check annotation and creates a mapping from the
// GroupVersionResource to the various statuses and their associated
// expressions.
func NewStatusMap(client client.Interface) (*StatusMap, error) {
	statusMap := &StatusMap{
		client:     client,
		mapping:    make(map[schema.GroupVersionResource]map[status.Status]Expression),
		restMapper: meta.NewDefaultRESTMapper([]schema.GroupVersion{}),
	}
	client.ApiextensionsClientSet()
	crds, err := statusMap.client.ApiextensionsClientSet().
		ApiextensionsV1().
		CustomResourceDefinitions().
		List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, crd := range crds.Items {
		if err = statusMap.addCRD(crd); err != nil {
			return nil, err
		}
	}

	return statusMap, nil
}

// ReadStatus returns object status
func (sm *StatusMap) ReadStatus(ctx context.Context, resource object.ObjMetadata) *event.ResourceStatus {
	gk := resource.GroupKind
	gvr, err := sm.restMapper.RESTMapping(gk, "v1")
	if err != nil {
		return handleResourceStatusError(resource, err)
	}
	options := metav1.GetOptions{}
	object, err := sm.client.DynamicClient().Resource(gvr.Resource).
		Namespace(resource.Namespace).Get(resource.Name, options)
	if err != nil {
		return handleResourceStatusError(resource, err)
	}
	return sm.ReadStatusForObject(ctx, object)
}

// ReadStatusForObject returns resource status for object.
// Current status will be returned only if expression matched.
func (sm *StatusMap) ReadStatusForObject(
	ctx context.Context, resource *unstructured.Unstructured) *event.ResourceStatus {
	identifier := object.ObjMetadata{
		GroupKind: resource.GroupVersionKind().GroupKind(),
		Name:      resource.GetName(),
		Namespace: resource.GetNamespace(),
	}
	gvk := resource.GroupVersionKind()
	restMapping, err := sm.restMapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return &event.ResourceStatus{
			Identifier: identifier,
			Status:     status.UnknownStatus,
			Error:      err,
		}
	}

	gvr := restMapping.Resource

	obj, err := sm.client.DynamicClient().Resource(gvr).Namespace(resource.GetNamespace()).
		Get(resource.GetName(), metav1.GetOptions{})
	if err != nil {
		return &event.ResourceStatus{
			Identifier: identifier,
			Status:     status.UnknownStatus,
			Error:      err,
		}
	}

	// No need to check for existence - if there isn't a mapping for this
	// resource, the following for loop won't run anyway
	for currentstatus, expression := range sm.mapping[gvr] {
		var matched bool
		matched, err = expression.Match(obj)
		if err != nil {
			return &event.ResourceStatus{
				Identifier: identifier,
				Status:     status.UnknownStatus,
				Error:      err,
			}
		}
		if matched {
			return &event.ResourceStatus{
				Identifier: identifier,
				Status:     currentstatus,
				Resource:   resource,
				Message:    fmt.Sprintf("%s is %s", resource.GroupVersionKind().Kind, currentstatus.String()),
			}
		}
	}
	return &event.ResourceStatus{
		Identifier: identifier,
		Status:     status.UnknownStatus,
		Error:      nil,
	}
}

// GetStatusForResource iterates over all of the stored conditions for the
// resource and returns the first status whose conditions are met.
func (sm *StatusMap) GetStatusForResource(resource document.Document) (status.Status, error) {
	gvk := getGVK(resource)

	restMapping, err := sm.restMapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return "", ErrResourceNotFound{resource.GetName()}
	}

	gvr := restMapping.Resource
	obj, err := sm.client.DynamicClient().Resource(gvr).Namespace(resource.GetNamespace()).
		Get(resource.GetName(), metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	// No need to check for existence - if there isn't a mapping for this
	// resource, the following for loop won't run anyway
	expressionMap := sm.mapping[gvr]
	for status, expression := range expressionMap {
		matched, err := expression.Match(obj)
		if err != nil {
			return "", err
		}
		if matched {
			return status, nil
		}
	}

	return status.UnknownStatus, nil
}

// addCRD adds the mappings from the CRD to its associated statuses
func (sm *StatusMap) addCRD(crd apiextensions.CustomResourceDefinition) error {
	annotations := crd.GetAnnotations()
	rawStatusChecks, ok := annotations["airshipit.org/status-check"]
	if !ok {
		// This crd doesn't have a status-check
		// annotation, so we should skip it.
		return nil
	}
	statusChecks, err := parseStatusChecks(rawStatusChecks)
	if err != nil {
		return err
	}

	gvrs := getGVRs(crd)
	for _, gvr := range gvrs {
		sm.GkMapping = append(sm.GkMapping, crd.GroupVersionKind().GroupKind())
		gvk := gvr.GroupVersion().WithKind(crd.Spec.Names.Kind)
		gvrSingular := gvr.GroupVersion().WithResource(crd.Spec.Names.Singular)
		sm.mapping[gvr] = statusChecks
		sm.restMapper.AddSpecific(gvk, gvr, gvrSingular, meta.RESTScopeNamespace)
	}

	return nil
}

// getGVRs constructs a slice of schema.GroupVersionResource for
// CustomResources defined by the CustomResourceDefinition.
func getGVRs(crd apiextensions.CustomResourceDefinition) []schema.GroupVersionResource {
	gvrs := make([]schema.GroupVersionResource, 0, len(crd.Spec.Versions))
	for _, version := range crd.Spec.Versions {
		gvr := schema.GroupVersionResource{
			Group:    crd.Spec.Group,
			Version:  version.Name,
			Resource: crd.Spec.Names.Plural,
		}
		gvrs = append(gvrs, gvr)
	}
	return gvrs
}

// getGVK constructs a schema.GroupVersionKind for a document
func getGVK(doc document.Document) schema.GroupVersionKind {
	toSchemaGvk := schema.GroupVersionKind{
		Group:   doc.GetGroup(),
		Version: doc.GetVersion(),
		Kind:    doc.GetKind(),
	}
	return toSchemaGvk
}

// parseStatusChecks takes a string containing a map of status names (e.g.
// Healthy) to the JSONPath filters associated with the statuses, and returns
// the Go object equivalent.
func parseStatusChecks(raw string) (map[status.Status]Expression, error) {
	type statusCheckType struct {
		Status    string `json:"status"`
		Condition string `json:"condition"`
	}

	var mappings []statusCheckType
	if err := json.Unmarshal([]byte(raw), &mappings); err != nil {
		return nil, ErrInvalidStatusCheck{
			What: fmt.Sprintf("unable to parse jsonpath: %q: %v", raw, err.Error()),
		}
	}

	expressionMap := make(map[status.Status]Expression)
	for _, mapping := range mappings {
		if mapping.Status == "" {
			return nil, ErrInvalidStatusCheck{What: "missing status field"}
		}

		if mapping.Condition == "" {
			return nil, ErrInvalidStatusCheck{What: "missing condition field"}
		}

		expressionMap[status.Status(mapping.Status)] = Expression{Condition: mapping.Condition}
	}

	return expressionMap, nil
}

// handleResourceStatusError construct the appropriate ResourceStatus
// object based on the type of error.
func handleResourceStatusError(identifier object.ObjMetadata, err error) *event.ResourceStatus {
	if errors.IsNotFound(err) {
		return &event.ResourceStatus{
			Identifier: identifier,
			Status:     status.NotFoundStatus,
			Message:    "Resource not found",
		}
	}
	return &event.ResourceStatus{
		Identifier: identifier,
		Status:     status.UnknownStatus,
		Error:      err,
	}
}
