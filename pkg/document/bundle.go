package document

import (
	"fmt"
	"io"

	"sigs.k8s.io/kustomize/v3/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/v3/k8sdeps/transformer"
	"sigs.k8s.io/kustomize/v3/k8sdeps/validator"
	"sigs.k8s.io/kustomize/v3/pkg/fs"
	"sigs.k8s.io/kustomize/v3/pkg/gvk"
	"sigs.k8s.io/kustomize/v3/pkg/loader"
	"sigs.k8s.io/kustomize/v3/pkg/plugins"
	"sigs.k8s.io/kustomize/v3/pkg/resmap"
	"sigs.k8s.io/kustomize/v3/pkg/resource"
	"sigs.k8s.io/kustomize/v3/pkg/target"
	"sigs.k8s.io/kustomize/v3/pkg/types"

	docplugins "opendev.org/airship/airshipctl/pkg/document/plugins"
	"opendev.org/airship/airshipctl/pkg/log"
	utilyaml "opendev.org/airship/airshipctl/pkg/util/yaml"
)

func init() {
	// NOTE (dukov) This is sort of a hack but it's the only way to add an
	// external 'builtin' plugin to Kustomize
	plugins.TransformerFactories[plugins.Unknown] = docplugins.NewTransformerLoader
}

// KustomizeBuildOptions contain the options for running a Kustomize build on a bundle
type KustomizeBuildOptions struct {
	KustomizationPath string
	OutputPath        string
	LoadRestrictor    loader.LoadRestrictorFunc
	OutOrder          int
}

// BundleFactory contains the objects within a bundle
type BundleFactory struct {
	KustomizeBuildOptions
	resmap.ResMap
	fs.FileSystem
}

// Bundle interface provides the specification for a bundle implementation
type Bundle interface {
	Write(out io.Writer) error
	GetKustomizeResourceMap() resmap.ResMap
	SetKustomizeResourceMap(resmap.ResMap) error
	GetKustomizeBuildOptions() KustomizeBuildOptions
	SetKustomizeBuildOptions(KustomizeBuildOptions) error
	SetFileSystem(fs.FileSystem) error
	GetFileSystem() fs.FileSystem
	Select(selector types.Selector) ([]Document, error)
	GetByGvk(string, string, string) ([]Document, error)
	GetByName(string) (Document, error)
	GetByAnnotation(string) ([]Document, error)
	GetByLabel(string) ([]Document, error)
	GetAllDocuments() ([]Document, error)
}

// NewBundle is a convenience function to create a new bundle
// Over time, it will evolve to support allowing more control
// for kustomize plugins
func NewBundle(fSys fs.FileSystem, kustomizePath string, outputPath string) (bundle Bundle, err error) {
	var options = KustomizeBuildOptions{
		KustomizationPath: kustomizePath,
		OutputPath:        outputPath,
		LoadRestrictor:    loader.RestrictionRootOnly,
		OutOrder:          0,
	}

	// init an empty bundle factory
	bundle = &BundleFactory{}

	// set the fs and build options we will use
	if err = bundle.SetFileSystem(fSys); err != nil {
		return nil, err
	}
	if err = bundle.SetKustomizeBuildOptions(options); err != nil {
		return nil, err
	}

	// boiler plate to allow us to run Kustomize build
	uf := kunstruct.NewKunstructuredFactoryImpl()
	pf := transformer.NewFactoryImpl()
	rf := resmap.NewFactory(resource.NewFactory(uf), pf)
	v := validator.NewKustValidator()

	pluginConfig := plugins.DefaultPluginConfig()
	pl := plugins.NewLoader(pluginConfig, rf)

	ldr, err := loader.NewLoader(
		bundle.GetKustomizeBuildOptions().LoadRestrictor, v, bundle.GetKustomizeBuildOptions().KustomizationPath, fSys)
	if err != nil {
		return bundle, err
	}

	defer func() {
		if e := ldr.Cleanup(); e != nil {
			log.Fatal("failed to cleanup loader ERROR: ", e)
		}
	}()

	kt, err := target.NewKustTarget(ldr, rf, pf, pl)
	if err != nil {
		return bundle, err
	}

	// build a resource map of kustomize rendered objects
	m, err := kt.MakeCustomizedResMap()
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
func (b *BundleFactory) SetFileSystem(fSys fs.FileSystem) error {
	b.FileSystem = fSys
	return nil
}

// GetFileSystem gets the filesystem that will be used by this bundle
func (b *BundleFactory) GetFileSystem() fs.FileSystem {
	return b.FileSystem
}

// GetAllDocuments returns all documents in this bundle
func (b *BundleFactory) GetAllDocuments() ([]Document, error) {
	docSet := []Document{}
	for _, res := range b.ResMap.Resources() {
		// Construct Bundle document for each resource returned
		doc, err := NewDocument(res)
		if err != nil {
			return docSet, err
		}
		docSet = append(docSet, doc)
	}
	return docSet, nil
}

// GetByName finds a document by name, error if more than one document found
// or if no documents found
func (b *BundleFactory) GetByName(name string) (Document, error) {
	resSet := []*resource.Resource{}
	for _, res := range b.ResMap.Resources() {
		if res.GetName() == name {
			resSet = append(resSet, res)
		}
	}
	// alanmeadows(TODO): improve this and other error potentials by
	// by adding strongly typed errors
	switch found := len(resSet); {
	case found == 0:
		return &DocumentFactory{}, fmt.Errorf("No documents found with name %s", name)
	case found > 1:
		return &DocumentFactory{}, fmt.Errorf("More than one document found with name %s", name)
	default:
		return NewDocument(resSet[0])
	}
}

// Select offers a direct interface to pass a Kustomize Selector to the bundle
// returning Documents that match the criteria
func (b *BundleFactory) Select(selector types.Selector) ([]Document, error) {
	// use the kustomize select method
	resources, err := b.ResMap.Select(selector)
	if err != nil {
		return []Document{}, err
	}

	// Construct Bundle document for each resource returned
	docSet := []Document{}
	for _, res := range resources {
		var doc Document
		doc, err = NewDocument(res)
		if err != nil {
			return docSet, err
		}
		docSet = append(docSet, doc)
	}
	return docSet, err
}

// GetByAnnotation is a convenience method to get documents for a particular annotation
func (b *BundleFactory) GetByAnnotation(annotation string) ([]Document, error) {
	// Construct kustomize annotation selector
	selector := types.Selector{AnnotationSelector: annotation}

	// pass it to the selector
	return b.Select(selector)
}

// GetByLabel is a convenience method to get documents for a particular label
func (b *BundleFactory) GetByLabel(label string) ([]Document, error) {
	// Construct kustomize annotation selector
	selector := types.Selector{LabelSelector: label}

	// pass it to the selector
	return b.Select(selector)
}

// GetByGvk is a convenience method to get documents for a particular Gvk tuple
func (b *BundleFactory) GetByGvk(group, version, kind string) ([]Document, error) {
	// Construct kustomize gvk object
	g := gvk.Gvk{Group: group, Version: version, Kind: kind}

	// pass it to the selector
	selector := types.Selector{Gvk: g}
	return b.Select(selector)
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
