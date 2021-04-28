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

// Label Selectors
const (
	BaseAirshipSelector       = "airshipit.org"
	EphemeralHostSelector     = BaseAirshipSelector + "/ephemeral-node in (True, true)"
	EphemeralUserDataSelector = BaseAirshipSelector + "/ephemeral-user-data in (True, true)"

	// Please note that by default every document in the manifest is to be deployed to kubernetes cluster.
	// so this selector simply checks that deploy-k8s label is not equal to false or False (string)
	DeployToK8sSelector = "airshipit.org/deploy-k8s notin (False, false)"
)

// GVKs
const (
	SecretKind        = "Secret"
	BareMetalHostKind = "BareMetalHost"

	ConfigMapKind    = "ConfigMap"
	ConfigMapVersion = "v1"

	ClusterctlMetadataKind    = "Metadata"
	ClusterctlMetadataVersion = "v1alpha3"
	ClusterctlMetadataGroup   = "clusterctl.cluster.x-k8s.io"

	// DocumentValidationGroup defines Group for document-validation container
	DocumentValidationGroup = "airshipit.org"
	// DocumentValidationVersion defines Version for document-validation container
	DocumentValidationVersion = "v1alpha1"
	// DocumentValidationKind defines Kind for document-validation container
	DocumentValidationKind = "GenericContainer"
	// DocumentValidationName defines Name for document-validation container
	DocumentValidationName = "document-validation"
)

// KustomizationFile is used for kustomization file
const KustomizationFile = "kustomization.yaml"
