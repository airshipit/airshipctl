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

package v1alpha1

import (
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true

// Templater plugin configuration for airship document model
type Templater struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// NOTE: controller-gen doesn't support map[string]interface{},
	// so we'll need to allow arbitrary JSON here and unmarshal it
	// in the templater

	// Values contains map with object parameters to render
	Values *v1.JSON `json:"values,omitempty"`

	// Template field is used to specify actual go-template which is going
	// to be used to render the object defined in Spec field
	Template string `json:"template,omitempty"`
}
