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

package poller

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	apitypes "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/cli-utils/pkg/kstatus/polling/engine"
	"sigs.k8s.io/cli-utils/pkg/kstatus/polling/event"
	"sigs.k8s.io/cli-utils/pkg/kstatus/status"
	"sigs.k8s.io/cli-utils/pkg/object"

	"opendev.org/airship/airshipctl/krm-functions/applier/image/types"
)

// CustomResourceReader is a wrapper for clu-utils genericClusterReader struct
type CustomResourceReader struct {
	// Reader is an implementation of the ClusterReader interface. It provides a
	// way for the StatusReader to fetch resources from the cluster.
	Reader engine.ClusterReader
	// Mapper provides a way to look up the resource types that are available
	// in the cluster.
	Mapper meta.RESTMapper

	// StatusFunc is a function for computing status of object
	StatusFunc func(u *unstructured.Unstructured) (*status.Result, error)
	// CondMap is a map with stored jsonpath expressions per GK to compute custom status
	CondMap map[schema.GroupKind]Expression
}

var _ engine.StatusReader = &CustomResourceReader{}

// NewCustomResourceReader implements custom logic to retrieve resource's status
func NewCustomResourceReader(reader engine.ClusterReader, mapper meta.RESTMapper,
	conditions ...types.Condition) engine.StatusReader {
	condMap := make(map[schema.GroupKind]Expression)
	for _, cond := range conditions {
		condMap[cond.GroupVersionKind().GroupKind()] = Expression{
			Condition: cond.JSONPath,
			Value:     cond.Value,
		}
	}

	return &CustomResourceReader{
		Reader:     reader,
		Mapper:     mapper,
		StatusFunc: status.Compute,
		CondMap:    condMap,
	}
}

// ReadStatus will fetch the resource identified by the given identifier
// from the cluster and return an ResourceStatus that will contain
// information about the latest state of the resource, its computed status
// and information about any generated resources.
func (c *CustomResourceReader) ReadStatus(ctx context.Context, identifier object.ObjMetadata) *event.ResourceStatus {
	obj, err := c.lookupResource(ctx, identifier)
	if err != nil {
		return handleResourceStatusError(identifier, err)
	}
	return c.ReadStatusForObject(ctx, obj)
}

// ReadStatusForObject is similar to ReadStatus, but instead of looking up the
// resource based on an identifier, it will use the passed-in resource.
func (c *CustomResourceReader) ReadStatusForObject(_ context.Context,
	obj *unstructured.Unstructured) *event.ResourceStatus {
	res, err := c.StatusFunc(obj)
	if err != nil {
		return &event.ResourceStatus{
			Identifier: toIdentifier(obj),
			Status:     status.UnknownStatus,
			Error:      err,
		}
	}

	if val, ok := c.CondMap[obj.GroupVersionKind().GroupKind()]; ok && res.Status == status.CurrentStatus {
		b, err := val.Match(obj.UnstructuredContent())
		if err != nil {
			return &event.ResourceStatus{
				Identifier: toIdentifier(obj),
				Status:     status.UnknownStatus,
				Error:      err,
				Message:    fmt.Sprintf("Unable to parse jsonpath '%s' in resource: %v", val.Condition, err),
			}
		}

		if b {
			return &event.ResourceStatus{
				Identifier: toIdentifier(obj),
				Status:     res.Status,
				Resource:   obj,
				Message:    res.Message,
			}
		}
		return &event.ResourceStatus{
			Identifier: toIdentifier(obj),
			Status:     status.InProgressStatus,
			Message:    fmt.Sprintf("Resource has not reached state '%s' at jsonpath '%s' yet", val.Value, val.Condition),
		}
	}

	return &event.ResourceStatus{
		Identifier: toIdentifier(obj),
		Status:     res.Status,
		Resource:   obj,
		Message:    res.Message,
	}
}

// lookupResource looks up a resource with the given identifier. It will use the rest mapper to resolve
// the version of the GroupKind given in the identifier.
// If the resource is found, it is returned. If it is not found or something
// went wrong, the function will return an error.
func (c *CustomResourceReader) lookupResource(ctx context.Context,
	identifier object.ObjMetadata) (*unstructured.Unstructured, error) {
	groupVersionKind, err := gvk(identifier.GroupKind, c.Mapper)
	if err != nil {
		return nil, err
	}

	var u unstructured.Unstructured
	u.SetGroupVersionKind(groupVersionKind)
	key := apitypes.NamespacedName{
		Name:      identifier.Name,
		Namespace: identifier.Namespace,
	}
	err = c.Reader.Get(ctx, key, &u)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// gvk looks up the GVK from a GroupKind using the rest mapper.
func gvk(gk schema.GroupKind, mapper meta.RESTMapper) (schema.GroupVersionKind, error) {
	mapping, err := mapper.RESTMapping(gk)
	if err != nil {
		return schema.GroupVersionKind{}, err
	}
	return mapping.GroupVersionKind, nil
}

func toIdentifier(u *unstructured.Unstructured) object.ObjMetadata {
	return object.ObjMetadata{
		GroupKind: u.GroupVersionKind().GroupKind(),
		Name:      u.GetName(),
		Namespace: u.GetNamespace(),
	}
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
