package implementations

import (
	"bytes"
	"path/filepath"

	"k8s.io/apimachinery/pkg/util/version"
	"sigs.k8s.io/cluster-api/cmd/clusterctl/client/repository"

	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/log"
)

const (
	// TODO Add metadata support. Clusterctl uses repository to read metadata.yaml file
	// to get clusterctl metadata which is a link between provider version and api version
	// that it supports, for example v0.3.0 --> v1alpha3
	dummyComponentPath = "components.yaml"
)

// Repository implements Repository from clusterctl project
type Repository struct {
	root           string
	versions       map[string]string
	defaultVersion string
}

var _ repository.Repository = &Repository{}

// ComponentsPath always returns same value, since it is not relevant without real filesystem
func (r *Repository) ComponentsPath() string {
	return dummyComponentPath
}

// GetVersions retrieve all versions from the repository
func (r *Repository) GetVersions() ([]string, error) {
	versions := make([]string, 0, len(r.versions))
	for availableVersion := range r.versions {
		_, err := version.ParseSemantic(availableVersion)
		if err != nil {
			// discard releases with tags that are not a valid semantic versions (the user can point explicitly to such releases)
			continue
		}
		versions = append(versions, availableVersion)
	}
	return versions, nil
}

// DefaultVersion highest version available
func (r *Repository) DefaultVersion() string {
	return r.defaultVersion
}

// RootPath not relevant without real filesystem
func (r *Repository) RootPath() string {
	return r.root
}

// GetFile returns all kubernetes resources that belong to cluster-api
// TODO Add metadata support(don't ignore filepath). Clusterctl uses repository to read
// metadata.yaml file to get clusterctl metadata which is a link between provider version
// and api version that it supports, for example v0.3.0 --> v1alpha3
func (r *Repository) GetFile(version string, _ string) ([]byte, error) {
	if version == "latest" {
		// default should be latest
		version = r.defaultVersion
	}
	path, ok := r.versions[version]
	if !ok {
		return nil, ErrVersionNotDefined{Version: version}
	}

	bundle, err := document.NewBundleByPath(filepath.Join(r.root, path))
	if err != nil {
		return nil, err
	}
	filteredBundle, err := bundle.SelectBundle(document.NewDeployToK8sSelector())
	if err != nil {
		return nil, err
	}
	buffer := bytes.NewBuffer([]byte{})
	err = filteredBundle.Write(buffer)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// NewRepository builds instance of repository
func NewRepository(root string, versions map[string]string) (repository.Repository, error) {
	var latestVersion *version.Version
	var latestStringVersion string
	// calculate latest version and delete versions that do not obey version syntax
	for ver := range versions {
		availableSemVersion, err := version.ParseSemantic(ver)
		if err != nil {
			// ignore and delete version if we can't parse it.
			log.Debugf(`Invalid version %s in repository versions map %q, ignoring it. Version must obey the syntax,
semantics of the "Semantic Versioning" specification (http://semver.org/)`, ver, versions)
			// delete the version so actual version list is clean
			delete(versions, ver)
			continue
		}
		if latestVersion == nil || latestVersion.LessThan(availableSemVersion) {
			latestVersion = availableSemVersion
			latestStringVersion = ver
		}
	}
	if latestStringVersion == "" {
		return nil, ErrNoVersionsAvailable{Versions: versions}
	}
	return &Repository{
		root:           root,
		versions:       versions,
		defaultVersion: latestStringVersion}, nil
}
