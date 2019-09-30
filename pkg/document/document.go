package document

import (
	"sigs.k8s.io/kustomize/v3/pkg/resource"
)

// DocumentFactory holds document data
type DocumentFactory struct {
	resource.Resource
}

// Document interface
type Document interface {
	GetKustomizeResource() resource.Resource
	SetKustomizeResource(*resource.Resource) error
	AsYAML() ([]byte, error)
	MarshalJSON() ([]byte, error)
	GetName() string
	GetKind() string
	GetNamespace() string
	GetString(path string) (string, error)
	GetStringSlice(path string) ([]string, error)
	GetBool(path string) (bool, error)
	GetFloat64(path string) (float64, error)
	GetInt64(path string) (int64, error)
	GetSlice(path string) ([]interface{}, error)
	GetStringMap(path string) (map[string]string, error)
	GetMap(path string) (map[string]interface{}, error)
}

// GetNamespace returns the namespace the resource thinks it's in.
func (d *DocumentFactory) GetNamespace() string {
	namespace, _ := d.GetString("metadata.namespace")
	// if err, namespace is empty, so no need to check.
	return namespace
}

// GetString returns the string value at path.
func (d *DocumentFactory) GetString(path string) (string, error) {
	r := d.GetKustomizeResource()
	return r.GetString(path)
}

// GetStringSlice returns a string slice at path.
func (d *DocumentFactory) GetStringSlice(path string) ([]string, error) {
	r := d.GetKustomizeResource()
	return r.GetStringSlice(path)
}

// GetBool returns a bool at path.
func (d *DocumentFactory) GetBool(path string) (bool, error) {
	r := d.GetKustomizeResource()
	return r.GetBool(path)
}

// GetFloat64 returns a float64 at path.
func (d *DocumentFactory) GetFloat64(path string) (float64, error) {
	r := d.GetKustomizeResource()
	return r.GetFloat64(path)
}

// GetInt64 returns an int64 at path.
func (d *DocumentFactory) GetInt64(path string) (int64, error) {
	r := d.GetKustomizeResource()
	return r.GetInt64(path)
}

// GetSlice returns a slice at path.
func (d *DocumentFactory) GetSlice(path string) ([]interface{}, error) {
	r := d.GetKustomizeResource()
	return r.GetSlice(path)
}

// GetStringMap returns a string map at path.
func (d *DocumentFactory) GetStringMap(path string) (map[string]string, error) {
	r := d.GetKustomizeResource()
	return r.GetStringMap(path)
}

// GetMap returns a map at path.
func (d *DocumentFactory) GetMap(path string) (map[string]interface{}, error) {
	r := d.GetKustomizeResource()
	return r.GetMap(path)
}

// AsYAML returns the document as a YAML byte stream.
func (d *DocumentFactory) AsYAML() ([]byte, error) {
	r := d.GetKustomizeResource()
	return r.AsYAML()
}

// MarshalJSON returns the document as JSON.
func (d *DocumentFactory) MarshalJSON() ([]byte, error) {
	r := d.GetKustomizeResource()
	return r.MarshalJSON()
}

// GetName returns the name: field from the document.
func (d *DocumentFactory) GetName() string {
	r := d.GetKustomizeResource()
	return r.GetName()
}

// GetKind returns the Kind: field from the document.
func (d *DocumentFactory) GetKind() string {
	r := d.GetKustomizeResource()
	return r.GetKind()
}

// GetKustomizeResource returns a Kustomize Resource object for this document.
func (d *DocumentFactory) GetKustomizeResource() resource.Resource {
	return d.Resource
}

// SetKustomizeResource sets a Kustomize Resource object for this document.
func (d *DocumentFactory) SetKustomizeResource(r *resource.Resource) error {
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

	var doc Document = &DocumentFactory{}
	err := doc.SetKustomizeResource(r)
	return doc, err

}
