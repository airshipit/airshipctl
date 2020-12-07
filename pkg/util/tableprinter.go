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

package util

import (
	"fmt"
	"io"
	"reflect"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	pe "sigs.k8s.io/cli-utils/pkg/kstatus/polling/event"
	"sigs.k8s.io/cli-utils/pkg/object"
	"sigs.k8s.io/cli-utils/pkg/print/table"
)

// Printable is an object which has Group, Version, Kind, Name and Namespace
// fields and appropriate methods to retrieve them. All K8s api objects
// implement methods above
type Printable interface {
	metav1.Object
	runtime.Object
}

var _ table.ResourceStates = &ResourceTable{}

// ResourceTable provides information about the resources that should be printed
type ResourceTable struct {
	resources  []Printable
	statusFunc PrintStatusFunction
}

// NewResourceTable creates resource status table
func NewResourceTable(object interface{}, statusFunc PrintStatusFunction) (*ResourceTable, error) {
	var resources []Printable
	value := reflect.ValueOf(object)
	switch value.Kind() {
	case reflect.Slice:
		resources = make([]Printable, value.Len())
		for i := 0; i < value.Len(); i++ {
			printable, ok := value.Index(i).Interface().(Printable)
			if !ok {
				return nil, fmt.Errorf("resource %#v is not printable", value.Index(i).Interface())
			}
			resources[i] = printable
		}
	default:
		res, ok := value.Interface().(Printable)
		if !ok {
			return nil, fmt.Errorf("resource %#v is not printable", value.Interface())
		}
		resources = []Printable{res}
	}

	return &ResourceTable{
		resources:  resources,
		statusFunc: statusFunc,
	}, nil
}

// Resources list of table rows
func (rt *ResourceTable) Resources() []table.Resource {
	result := make([]table.Resource, len(rt.resources))
	for i, obj := range rt.resources {
		result[i] = &resource{Printable: obj, statusFunc: rt.statusFunc}
	}
	return result
}

//Error returns error for resource table
func (rt *ResourceTable) Error() error {
	return fmt.Errorf("error table printing")
}

var _ table.Resource = &resource{}

type resource struct {
	Printable
	statusFunc PrintStatusFunction
}

// Identifier opf the resource
func (r *resource) Identifier() object.ObjMetadata {
	return object.ObjMetadata{
		Namespace: r.GetNamespace(),
		Name:      r.GetName(),
		GroupKind: r.GetObjectKind().GroupVersionKind().GroupKind(),
	}
}

// ResourceStatus returns resource status object
func (r *resource) ResourceStatus() *PrintResourceStatus {
	return r.statusFunc(r.Printable)
}

// SubResources list of subresources
func (r *resource) SubResources() []table.Resource {
	return nil
}

// DefaultTablePrinter returns basic table printer with 2 columns: Namespace
// and Name/Kind
func DefaultTablePrinter(out, errOut io.Writer) *table.BaseTablePrinter {
	cols := []table.ColumnDefinition{
		table.MustColumn("namespace"),
		table.MustColumn("resource"),
	}
	return &table.BaseTablePrinter{
		IOStreams: genericclioptions.IOStreams{
			Out:    out,
			ErrOut: errOut,
		},
		Columns: cols,
	}
}

// PrintResourceStatus alias for ResourceStatus
type PrintResourceStatus = pe.ResourceStatus

// PrintStatusFunction alias for status function
type PrintStatusFunction = func(Printable) *PrintResourceStatus

// DefaultStatusFunction for resource status
func DefaultStatusFunction() PrintStatusFunction {
	return func(obj Printable) *PrintResourceStatus {
		unsContent, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
		identifier := object.ObjMetadata{
			Namespace: obj.GetNamespace(),
			Name:      obj.GetName(),
			GroupKind: obj.GetObjectKind().GroupVersionKind().GroupKind(),
		}
		return &PrintResourceStatus{
			Identifier: identifier,
			Resource:   &unstructured.Unstructured{Object: unsContent},
			Error:      err,
		}
	}
}
