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
	"encoding/json"
	"fmt"

	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/k8s/client"
)

// A Status represents a kubernetes resource's state.
type Status string

// These represent the default statuses
const (
	UnknownStatus = Status("Unknown")
)

// StatusMap holds a mapping of schema.GroupVersionResource to various statuses
// a resource may be in, as well as the Expression used to check for that
// status.
type StatusMap struct {
	client     client.Interface
	mapping    map[schema.GroupVersionResource]map[Status]Expression
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
		mapping:    make(map[schema.GroupVersionResource]map[Status]Expression),
		restMapper: meta.NewDefaultRESTMapper([]schema.GroupVersion{}),
	}

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

// GetStatusForResource iterates over all of the stored conditions for the
// resource and returns the first status whose conditions are met.
func (sm *StatusMap) GetStatusForResource(resource document.Document) (Status, error) {
	gvk, err := getGVK(resource)
	if err != nil {
		return "", err
	}

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

	return UnknownStatus, nil
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
//
// TODO(howell): This should probably be a member method of the
// document.Document interface.
func getGVK(doc document.Document) (schema.GroupVersionKind, error) {
	apiVersion, err := doc.GetString("apiVersion")
	if err != nil {
		return schema.GroupVersionKind{}, err
	}
	gv, err := schema.ParseGroupVersion(apiVersion)
	if err != nil {
		return schema.GroupVersionKind{}, err
	}

	return gv.WithKind(doc.GetKind()), nil
}

// parseStatusChecks takes a string containing a map of status names (e.g.
// Healthy) to the JSONPath filters associated with the statuses, and returns
// the Go object equivalent.
func parseStatusChecks(raw string) (map[Status]Expression, error) {
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

	expressionMap := make(map[Status]Expression)
	for _, mapping := range mappings {
		if mapping.Status == "" {
			return nil, ErrInvalidStatusCheck{What: "missing status field"}
		}

		if mapping.Condition == "" {
			return nil, ErrInvalidStatusCheck{What: "missing condition field"}
		}

		expressionMap[Status(mapping.Status)] = Expression{Condition: mapping.Condition}
	}

	return expressionMap, nil
}
