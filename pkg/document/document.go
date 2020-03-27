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
	"sigs.k8s.io/kustomize/api/resource"
)

// Factory holds document data
type Factory struct {
	resource.Resource
}

// Document interface
type Document interface {
	Annotate(key, value string)
	AsYAML() ([]byte, error)
	GetAnnotations() map[string]string
	GetBool(path string) (bool, error)
	GetFloat64(path string) (float64, error)
	GetInt64(path string) (int64, error)
	GetKind() string
	GetLabels() map[string]string
	GetMap(path string) (map[string]interface{}, error)
	GetName() string
	GetNamespace() string
	GetSlice(path string) ([]interface{}, error)
	GetString(path string) (string, error)
	GetStringMap(path string) (map[string]string, error)
	GetStringSlice(path string) ([]string, error)
	Label(key, value string)
	MarshalJSON() ([]byte, error)
}

// Factory implements Document
var _ Document = &Factory{}

// Annotate document by applying annotation to it
func (d *Factory) Annotate(key string, value string) {
	annotations := d.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}
	annotations[key] = value
	d.SetAnnotations(annotations)
}

// Label document by applying label on it
func (d *Factory) Label(key string, value string) {
	labels := d.GetKustomizeResource().GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}
	labels[key] = value
	d.GetKustomizeResource().SetLabels(labels)
}

// GetNamespace returns the namespace the resource thinks it's in.
func (d *Factory) GetNamespace() string {
	r := d.GetKustomizeResource()
	return r.GetNamespace()
}

// GetString returns the string value at path.
func (d *Factory) GetString(path string) (string, error) {
	r := d.GetKustomizeResource()
	return r.GetString(path)
}

// GetStringSlice returns a string slice at path.
func (d *Factory) GetStringSlice(path string) ([]string, error) {
	r := d.GetKustomizeResource()
	return r.GetStringSlice(path)
}

// GetBool returns a bool at path.
func (d *Factory) GetBool(path string) (bool, error) {
	r := d.GetKustomizeResource()
	return r.GetBool(path)
}

// GetFloat64 returns a float64 at path.
func (d *Factory) GetFloat64(path string) (float64, error) {
	r := d.GetKustomizeResource()
	return r.GetFloat64(path)
}

// GetInt64 returns an int64 at path.
func (d *Factory) GetInt64(path string) (int64, error) {
	r := d.GetKustomizeResource()
	return r.GetInt64(path)
}

// GetSlice returns a slice at path.
func (d *Factory) GetSlice(path string) ([]interface{}, error) {
	r := d.GetKustomizeResource()
	return r.GetSlice(path)
}

// GetStringMap returns a string map at path.
func (d *Factory) GetStringMap(path string) (map[string]string, error) {
	r := d.GetKustomizeResource()
	return r.GetStringMap(path)
}

// GetMap returns a map at path.
func (d *Factory) GetMap(path string) (map[string]interface{}, error) {
	r := d.GetKustomizeResource()
	return r.GetMap(path)
}

// AsYAML returns the document as a YAML byte stream.
func (d *Factory) AsYAML() ([]byte, error) {
	r := d.GetKustomizeResource()
	return r.AsYAML()
}

// MarshalJSON returns the document as JSON.
func (d *Factory) MarshalJSON() ([]byte, error) {
	r := d.GetKustomizeResource()
	return r.MarshalJSON()
}

// GetName returns the name: field from the document.
func (d *Factory) GetName() string {
	r := d.GetKustomizeResource()
	return r.GetName()
}

// GetKind returns the Kind: field from the document.
func (d *Factory) GetKind() string {
	r := d.GetKustomizeResource()
	return r.GetKind()
}

// GetKustomizeResource returns a Kustomize Resource object for this document.
func (d *Factory) GetKustomizeResource() resource.Resource {
	return d.Resource
}

// SetKustomizeResource sets a Kustomize Resource object for this document.
func (d *Factory) SetKustomizeResource(r *resource.Resource) error {
	d.Resource = *r
	return nil
}

// NewDocument is a convenience method to construct a new Document.  Although
// an error is unlikely at this time, this provides some future proofing for
// when we want more strict airship specific validation of documents getting
// created as this would be the front door for all Kustomize->Airship
// documents - e.g. in the future all documents require an airship
// annotation X
func NewDocument(r *resource.Resource) (Document, error) {
	doc := &Factory{}
	err := doc.SetKustomizeResource(r)
	return doc, err
}
