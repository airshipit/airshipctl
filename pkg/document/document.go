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
	b64 "encoding/base64"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"

	"sigs.k8s.io/kustomize/api/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/api/resource"
	"sigs.k8s.io/yaml"
)

// Factory holds document data
type Factory struct {
	resource.Resource
}

// Document interface
type Document interface {
	Annotate(map[string]string)
	AsYAML() ([]byte, error)
	GetAnnotations() map[string]string
	GetBool(path string) (bool, error)
	GetFloat64(path string) (float64, error)
	GetFieldValue(path string) (interface{}, error)
	GetGroup() string
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
	GetVersion() string
	Label(map[string]string)
	MarshalJSON() ([]byte, error)
	ToObject(interface{}) error
	ToAPIObject(runtime.Object, *runtime.Scheme) error
}

// Factory implements Document
var _ Document = &Factory{}

// Annotate document by applying annotations as map
func (d *Factory) Annotate(newAnnotations map[string]string) {
	annotations := d.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}
	// override Current labels
	for key, val := range newAnnotations {
		annotations[key] = val
	}
	d.SetAnnotations(annotations)
}

// Label document by applying labels as map
func (d *Factory) Label(newLabels map[string]string) {
	labels := d.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}
	// override Current labels
	for key, val := range newLabels {
		labels[key] = val
	}
	d.SetLabels(labels)
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
	slice, err := d.GetSlice(path)
	if err != nil {
		return nil, err
	}
	lst := make([]string, len(slice))
	for i, val := range slice {
		str, ok := val.(string)
		if !ok {
			return nil, ErrBadValueFormat{Value: path, Expected: "[]string", Actual: fmt.Sprintf("%T", val)}
		}
		lst[i] = str
	}
	return lst, nil
}

// GetBool returns a bool at path.
func (d *Factory) GetBool(path string) (bool, error) {
	val, err := d.GetFieldValue(path)
	if err != nil {
		return false, err
	}
	result, ok := val.(bool)
	if !ok {
		return false, ErrBadValueFormat{Value: path, Expected: "bool", Actual: fmt.Sprintf("%T", val)}
	}
	return result, nil
}

// GetFloat64 returns a float64 at path.
func (d *Factory) GetFloat64(path string) (float64, error) {
	val, err := d.GetFieldValue(path)
	if err != nil {
		return 0, err
	}
	result, ok := val.(float64)
	if !ok {
		return 0, ErrBadValueFormat{Value: path, Expected: "float64", Actual: fmt.Sprintf("%T", val)}
	}
	return result, nil
}

// GetInt64 returns an int64 at path.
func (d *Factory) GetInt64(path string) (int64, error) {
	val, err := d.GetFieldValue(path)
	if err != nil {
		return 0, err
	}
	// TODO: kyaml implementation (which we now use) of ifc.KunstructuredFactory uses strconv.Atoi
	// to decode integer values, so it is always int not int64.
	result, ok := val.(int64)
	if !ok {
		// Let's try to cast it to int and convert to int64
		iresult, iok := val.(int)
		if !iok {
			return 0, ErrBadValueFormat{Value: path, Expected: "int64", Actual: fmt.Sprintf("%T", val)}
		}
		return int64(iresult), nil
	}
	return result, nil
}

// GetSlice returns a slice at path.
func (d *Factory) GetSlice(path string) ([]interface{}, error) {
	r := d.GetKustomizeResource()
	return r.GetSlice(path)
}

// GetStringMap returns a string map at path.
func (d *Factory) GetStringMap(path string) (map[string]string, error) {
	val, err := d.GetMap(path)
	if err != nil {
		return nil, err
	}
	strMap := make(map[string]string, len(val))
	for k, v := range val {
		str, ok := v.(string)
		if !ok {
			return nil, ErrBadValueFormat{Value: path, Expected: "map[string]string", Actual: fmt.Sprintf("%T", val)}
		}
		strMap[k] = str
	}
	return strMap, nil
}

// GetMap returns a map at path.
func (d *Factory) GetMap(path string) (map[string]interface{}, error) {
	val, err := d.GetFieldValue(path)
	if err != nil {
		return nil, err
	}
	result, ok := val.(map[string]interface{})
	if !ok {
		return nil, ErrBadValueFormat{Value: path, Expected: "map[string]interface{}", Actual: fmt.Sprintf("%T", val)}
	}
	return result, nil
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

// GetGroup returns api group from apiVersion field
func (d *Factory) GetGroup() string {
	r := d.GetKustomizeResource()
	return r.GetGvk().Group
}

// GetVersion returns api version from apiVersion field
func (d *Factory) GetVersion() string {
	r := d.GetKustomizeResource()
	return r.GetGvk().Version
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

// ToObject serializes document to object passed as an argument
func (d *Factory) ToObject(obj interface{}) error {
	docYAML, err := d.AsYAML()
	if err != nil {
		return err
	}
	return yaml.Unmarshal(docYAML, obj)
}

// ToAPIObject de-serializes a document into a runtime.Object
func (d *Factory) ToAPIObject(obj runtime.Object, scheme *runtime.Scheme) error {
	y, err := d.AsYAML()
	if err != nil {
		return err
	}

	yamlSerializer := json.NewSerializerWithOptions(
		json.DefaultMetaFactory,
		scheme,
		scheme,
		json.SerializerOptions{Yaml: true, Pretty: false, Strict: false})

	_, _, err = yamlSerializer.Decode(y, nil, obj)
	return err
}

// GetFieldValue get object at path
func (d *Factory) GetFieldValue(path string) (interface{}, error) {
	r := d.GetKustomizeResource()
	return r.GetFieldValue(path)
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

// NewDocumentFromBytes constructs document from bytes
func NewDocumentFromBytes(b []byte) (Document, error) {
	res, err := resource.NewFactory(kunstruct.NewKunstructuredFactoryImpl()).FromBytes(b)
	if err != nil {
		return nil, err
	}
	doc := &Factory{}
	err = doc.SetKustomizeResource(res)
	return doc, err
}

// DecodeSecretData returns base64-decoded secret data
func DecodeSecretData(cfg Document, key string) ([]byte, error) {
	var needsBase64Decode = false

	// TODO(alanmeadows): distinguish between missing key
	// and missing data/stringData keys in the Secret
	data, err := cfg.GetStringMap("data")
	if err == nil {
		needsBase64Decode = true
	} else {
		// we'll catch any error below
		data, err = cfg.GetStringMap("stringData")
		if err != nil {
			return nil, ErrDataNotSupplied{DocName: cfg.GetName(), Key: "data or stringData"}
		}
	}

	res, ok := data[key]
	if !ok {
		return nil, ErrDataNotSupplied{DocName: cfg.GetName(), Key: key}
	}

	if needsBase64Decode {
		return b64.StdEncoding.DecodeString(res)
	}
	return []byte(res), nil
}
