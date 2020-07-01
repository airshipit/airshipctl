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
	"io"
	"strings"

	"sigs.k8s.io/kustomize/api/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	"sigs.k8s.io/kustomize/api/types"

	"opendev.org/airship/airshipctl/pkg/environment"
	utilyaml "opendev.org/airship/airshipctl/pkg/util/yaml"
)

// KustomizeBuildOptions contain the options for running a Kustomize build on a bundle
type KustomizeBuildOptions struct {
	KustomizationPath string
	LoadRestrictions  types.LoadRestrictions
}

// BundleFactory contains the objects within a bundle
type BundleFactory struct {
	KustomizeBuildOptions
	resmap.ResMap
	FileSystem
}

// Bundle interface provides the specification for a bundle implementation
type Bundle interface {
	Write(out io.Writer) error
	SetFileSystem(FileSystem) error
	GetFileSystem() FileSystem
	Select(selector Selector) ([]Document, error)
	SelectOne(selector Selector) (Document, error)
	SelectBundle(selector Selector) (Bundle, error)
	SelectByFieldValue(string, func(interface{}) bool) (Bundle, error)
	GetByGvk(string, string, string) ([]Document, error)
	GetByName(string) (Document, error)
	GetByAnnotation(annotationSelector string) ([]Document, error)
	GetByLabel(labelSelector string) ([]Document, error)
	GetAllDocuments() ([]Document, error)
	Append(Document) error
}

// NewBundleByPath helper function that returns new document.Bundle interface based on clusterType and
// phase, example: helpers.NewBunde(airConfig, "ephemeral", "initinfra")
func NewBundleByPath(rootPath string) (Bundle, error) {
	return NewBundle(NewDocumentFs(), rootPath)
}

// NewBundle is a convenience function to create a new bundle
// Over time, it will evolve to support allowing more control
// for kustomize plugins
func NewBundle(fSys FileSystem, kustomizePath string) (Bundle, error) {
	var options = KustomizeBuildOptions{
		KustomizationPath: kustomizePath,
		LoadRestrictions:  types.LoadRestrictionsRootOnly,
	}

	// init an empty bundle factory
	bundle := &BundleFactory{}

	// set the fs and build options we will use
	if err := bundle.SetFileSystem(fSys); err != nil {
		return nil, err
	}
	if err := bundle.SetKustomizeBuildOptions(options); err != nil {
		return nil, err
	}

	var o = krusty.Options{
		DoLegacyResourceSort: true, // Default and what we want
		LoadRestrictions:     options.LoadRestrictions,
		DoPrune:              false, // Default
		PluginConfig: &types.PluginConfig{
			AbsPluginHome:      environment.PluginPath(),
			PluginRestrictions: types.PluginRestrictionsNone,
		},
	}

	kustomizer := krusty.MakeKustomizer(fSys, &o)
	m, err := kustomizer.Run(kustomizePath)
	if err != nil {
		return bundle, err
	}
	err = bundle.SetKustomizeResourceMap(m)
	return bundle, err
}

// GetKustomizeResourceMap returns a Kustomize Resource Map for this bundle
func (b *BundleFactory) GetKustomizeResourceMap() resmap.ResMap {
	return b.ResMap
}

// SetKustomizeResourceMap allows us to set the populated resource map for this bundle.  In
// the future, it may modify it before saving it.
func (b *BundleFactory) SetKustomizeResourceMap(r resmap.ResMap) error {
	b.ResMap = r
	return nil
}

// GetKustomizeBuildOptions returns the build options object used to generate the resource map
// for this bundle
func (b *BundleFactory) GetKustomizeBuildOptions() KustomizeBuildOptions {
	return b.KustomizeBuildOptions
}

// SetKustomizeBuildOptions sets the build options to be used for this bundle. In
// the future, it may perform some basic validations.
func (b *BundleFactory) SetKustomizeBuildOptions(k KustomizeBuildOptions) error {
	b.KustomizeBuildOptions = k
	return nil
}

// SetFileSystem sets the filesystem that will be used by this bundle
func (b *BundleFactory) SetFileSystem(fSys FileSystem) error {
	b.FileSystem = fSys
	return nil
}

// GetFileSystem gets the filesystem that will be used by this bundle
func (b *BundleFactory) GetFileSystem() FileSystem {
	return b.FileSystem
}

// GetAllDocuments returns all documents in this bundle
func (b *BundleFactory) GetAllDocuments() ([]Document, error) {
	docSet := make([]Document, len(b.ResMap.Resources()))
	for i, res := range b.ResMap.Resources() {
		// Construct Bundle document for each resource returned
		doc, err := NewDocument(res)
		if err != nil {
			return docSet, err
		}
		docSet[i] = doc
	}
	return docSet, nil
}

// GetByName finds a document by name
func (b *BundleFactory) GetByName(name string) (Document, error) {
	return b.SelectOne(NewSelector().ByName(name))
}

// Select offers an interface to pass a Selector, built on top of kustomize Selector
// to the bundle returning Documents that match the criteria
func (b *BundleFactory) Select(selector Selector) ([]Document, error) {
	// use the kustomize select method
	resources, err := b.ResMap.Select(selector.Selector)
	if err != nil {
		return []Document{}, err
	}

	// Construct Bundle document for each resource returned
	docSet := make([]Document, len(resources))
	for i, res := range resources {
		var doc Document
		doc, err = NewDocument(res)
		if err != nil {
			return docSet, err
		}
		docSet[i] = doc
	}
	return docSet, err
}

// SelectOne serves the common use case where you expect one match
// and only one match to your selector -- in other words, you want to
// error if you didn't find any documents, and error if you found
// more than one.  This reduces code repetition that would otherwise
// be scattered around that evaluates the length of the doc set returned
// for this common case
func (b *BundleFactory) SelectOne(selector Selector) (Document, error) {
	docSet, err := b.Select(selector)
	if err != nil {
		return nil, err
	}

	// evaluate docSet for at least one document, and no more than
	// one document and raise errors as appropriate
	switch numDocsFound := len(docSet); {
	case numDocsFound == 0:
		return nil, ErrDocNotFound{Selector: selector}
	case numDocsFound > 1:
		return nil, ErrMultiDocsFound{Selector: selector}
	}
	return docSet[0], nil
}

// SelectBundle offers an interface to pass a Selector, built on top of kustomize Selector
// to the bundle returning a new Bundle that matches the criteria.  This is useful
// where you want to actually prune the underlying bundle you are working with
// rather then getting back the matching documents for scenarios like
// test cases where you want to pass in custom "filtered" bundles
// specific to the test case
func (b *BundleFactory) SelectBundle(selector Selector) (Bundle, error) {
	// use the kustomize select method
	resources, err := b.ResMap.Select(selector.Selector)
	if err != nil {
		return nil, err
	}

	// create a blank resourcemap and append the found resources
	// into the new resource map
	resourceMap := resmap.New()
	for _, res := range resources {
		if err = resourceMap.Append(res); err != nil {
			return nil, err
		}
	}

	// return a new bundle with the same options and filesystem
	// as this one but with a reduced resourceMap
	return &BundleFactory{
		KustomizeBuildOptions: b.KustomizeBuildOptions,
		ResMap:                resourceMap,
		FileSystem:            b.FileSystem,
	}, nil
}

// SelectByFieldValue returns new Bundle with filtered resource documents.
// Method iterates over all resources in the bundle. If resource has field
// (i.e. key) specified in JSON path, and the comparison function returns
// 'true' for value referenced by JSON path, then resource is added to
// resulting bundle.
// Example:
// The bundle contains 3 documents
//
//     ---
//     apiVersion: v1
//     kind: DocKind1
//     metadata:
//       name: doc1
//     spec:
//       somekey:
//         somefield: "someValue"
//     ---
//     apiVersion: v1
//     kind: DocKind2
//     metadata:
//       name: doc2
//     spec:
//       somekey:
//         somefield: "someValue"
//     ---
//     apiVersion: v1
//     kind: DocKind1
//     metadata:
//       name: doc3
//     spec:
//       somekey:
//         somefield: "someOtherValue"
//
// Execution of bundleInstance.SelectByFieldValue(
//		"spec.somekey.somefield",
//		func(v interface{}) { return v == "someValue" })
// will return a new Bundle instance containing 2 documents:
//     ---
//     apiVersion: v1
//     kind: DocKind1
//     metadata:
//       name: doc1
//     spec:
//       somekey:
//         somefield: "someValue"
//     ---
//     apiVersion: v1
//     kind: DocKind2
//     metadata:
//       name: doc2
//     spec:
//       somekey:
//         somefield: "someValue"
func (b *BundleFactory) SelectByFieldValue(path string, condition func(interface{}) bool) (Bundle, error) {
	result := &BundleFactory{
		KustomizeBuildOptions: b.KustomizeBuildOptions,
		FileSystem:            b.FileSystem,
	}
	resourceMap := resmap.New()
	for _, res := range b.Resources() {
		val, err := res.GetFieldValue(path)
		if err != nil {
			if strings.Contains(err.Error(), "no field named") {
				// this resource doesn't have the specified field - skip it
				continue
			} else {
				return nil, err
			}
		}

		if condition(val) {
			if err = resourceMap.Append(res); err != nil {
				return nil, err
			}
		}
	}

	if err := result.SetKustomizeResourceMap(resourceMap); err != nil {
		return nil, err
	}
	return result, nil
}

// GetByAnnotation is a convenience method to get documents for a particular annotation
func (b *BundleFactory) GetByAnnotation(annotationSelector string) ([]Document, error) {
	// Construct kustomize annotation selector
	selector := NewSelector().ByAnnotation(annotationSelector)
	// pass it to the selector
	return b.Select(selector)
}

// GetByLabel is a convenience method to get documents for a particular label
func (b *BundleFactory) GetByLabel(labelSelector string) ([]Document, error) {
	// Construct kustomize label selector
	selector := NewSelector().ByLabel(labelSelector)
	// pass it to the selector
	return b.Select(selector)
}

// GetByGvk is a convenience method to get documents for a particular Gvk tuple
func (b *BundleFactory) GetByGvk(group, version, kind string) ([]Document, error) {
	// Construct kustomize gvk object
	selector := NewSelector().ByGvk(group, version, kind)
	// pass it to the selector
	return b.Select(selector)
}

// Append bundle with the document, this only works with document interface implementation
// that is provided by this package
func (b *BundleFactory) Append(doc Document) error {
	yaml, err := doc.AsYAML()
	if err != nil {
		return err
	}
	res, err := resource.NewFactory(kunstruct.NewKunstructuredFactoryImpl()).FromBytes(yaml)
	if err != nil {
		return nil
	}
	return b.ResMap.Append(res)
}

// Write will write out the entire bundle resource map
func (b *BundleFactory) Write(out io.Writer) error {
	for _, res := range b.ResMap.Resources() {
		err := utilyaml.WriteOut(out, res)
		if err != nil {
			return err
		}
	}
	return nil
}
