/*
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// RepoProperties The name of each key defined at this level should identify a Helm repository.
// Each helm_repository object is required to have a "url" key that
// specifies the location of the repository.
type RepoProperties struct {
	URL string `json:"url"`
}

// RepositorySpec defines the additional properties for repository
type RepositorySpec map[string]RepoProperties

// ChartSourceRef defines the properties of the Chart SourceRef like Kind and Name
type ChartSourceRef struct {
	Kind string `json:"kind"`
	Name string `json:"name"`
}

// ChartProperties defines the properties of the chart like Chart and version
type ChartProperties struct {
	Chart     string         `json:"chart"`
	Version   string         `json:"version"`
	SourceRef ChartSourceRef `json:"sourceRef,omitempty"`
}

// ChartSpec defines the spec for charts
type ChartSpec map[string]ChartProperties

// FileProperties The name of each key defined at this level should identify a
// single file. Each file object is required to have a "url" property defined,
// and may also define a "checksum" property.
type FileProperties struct {
	URL      string `json:"url"`
	Checksum string `json:"checksum,omitempty"`
}

// AirshipctlFunctionFileMap The name of each key defined at this level should identify a
// single file. Each file object is required to have a "url" property defined,
// and may also define a "checksum" property.
type AirshipctlFunctionFileMap map[string]FileProperties

// FileSpec The name of each key defined here should refer to the airshipctl
// function in which the file will be used.
type FileSpec map[string]AirshipctlFunctionFileMap

// ImageURLSpec defines the properties of Image URL like repository and tag
type ImageURLSpec struct {
	Repository string `json:"repository"`
	Tag        string `json:"tag"`
}

// CAPIImageProperties defines the spec for CAPI images
type CAPIImageProperties struct {
	Manager     ImageURLSpec `json:"manager"`
	AuthProxy   ImageURLSpec `json:"auth_proxy"`
	IPAMManager ImageURLSpec `json:"ipam-manager,omitempty"`
}

// CAPIImageSpec defines the additional properties for CAPI Images
type CAPIImageSpec map[string]CAPIImageProperties

// ImageURL defines the URL for an image
type ImageURL struct {
	Image string `json:"image"`
}

// KubernetesResourceMap The name of each key defined at this level should identify a single
// image. Each image object is required to have an "image" property which specifies
// the full URL for the image (i.e. repository/image:tag) as a single string.
type KubernetesResourceMap map[string]ImageURL

// AirshipctlFunctionImageRepoMap The name of each key defined here should refer to the Kubernetes
// resource document into which an image will be substituted, such as a
// Deployment or DaemonSet.
type AirshipctlFunctionImageRepoMap map[string]KubernetesResourceMap

// ImageSpec The name of each key defined here should refer to the airshipctl
// function to which the collection of images belongs, such as "baremetal-operator".
type ImageSpec map[string]AirshipctlFunctionImageRepoMap

// ImageRepositorySpec defines the spec for a repository that includes repository URL,
// Name and one of Hash/Tag/SHA/Digest.
type ImageRepositorySpec struct {
	Repository string `json:"repository"`
	Hash       string `json:"hash,omitempty"`
	Tag        string `json:"tag,omitempty"`
	SHA        string `json:"sha,omitempty"`
	Digest     string `json:"digest,omitempty"`

	// Name is an optional property that is used to specify the name of
	// an image. Typically, this format is only needed for charts such as dex-aio,
	// which uses "repo", "name", and "tag" properties to declare images, rather
	// than the more commonly used "repository" and "tag". In such cases, "repository"
	// should contain only the name of the repository (e.g. "quay.io") and the "name"
	// property should contain the image name (e.g. "metal3-io/ironic").
	Name string `json:"name,omitempty"`
}

// AirshipctlFunctionImageComponentMap The name of each key defined at this level should identify a single
// image. Each image object must have a "repository" property, and must have a
// property named "tag", "hash", "sha", or "digest".
type AirshipctlFunctionImageComponentMap map[string]ImageRepositorySpec

// ImageComponentSpec The name of each key defined at this level should refer to the
// airshipctl function to which a collection of images belongs, such as
// "baremetal-operator".
type ImageComponentSpec map[string]AirshipctlFunctionImageComponentMap

// VersionsCatalogueSpec defines the default versions catalog for functions hosted in the airshipctl project
type VersionsCatalogueSpec struct {
	// helm_repositories defines Helm repositories required by HelmReleases.
	HelmRepositories RepositorySpec `json:"helm_repositories,omitempty"`

	// charts defines collections of Helm charts. i
	// The name of each key in this section should identify a specific chart, and each
	// chart object must have "chart" and "version" properties defined.
	Charts ChartSpec `json:"charts,omitempty"`

	// files defines collections of files required by airshipctl functions.
	Files FileSpec `json:"files,omitempty"`

	// capi_images defines collections of images used by cluster API.
	// The name of each key in this section should correspond to the airshipctl
	// function in which the images will be used, such as "capm3". Each capi_image
	// object must have a "manager" and "auth_proxy" object, each of which must have
	// "repository" and "tag" properties defined. capi_images may also include an
	// optional "ipam-manager" object, which must also have "repository" and "tag"
	// properties defined.
	CAPIImages CAPIImageSpec `json:"capi_images,omitempty"`

	// images defines collections of images that are declared as complete
	// URLs rather than as a collection of discrete parts, such as "repository" and
	// "tag" or "sha". This section of the catalog is organized by
	// airshipctl function -> Deployments in function -> images in Deployment.
	Images ImageSpec `json:"images,omitempty"`

	// image_components defines images that are declared using the Helm-style
	// format that breaks image URLs into discrete parts, such as "repository" and "tag".
	// Images in this section of the catalog are grouped by airshipctl function ->
	// images in function.
	ImageComponents ImageComponentSpec `json:"image_components,omitempty"`

	// Allows for the specification of the kubernetes version being used.
	Kubernetes string `json:"kubernetes,omitempty"`

	// Allows for the specification of the image repositories
	ImageRepositories map[string]ImageRepositorySpec `json:"image_repositories,omitempty"`
}

// +kubebuilder:object:root=true

// VersionsCatalogue is the Schema for the versions catalogs API
type VersionsCatalogue struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec VersionsCatalogueSpec `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true

// VersionsCatalogues contains a list of versions catalog
type VersionsCatalogues struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VersionsCatalogue `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VersionsCatalogue{}, &VersionsCatalogues{})
}
