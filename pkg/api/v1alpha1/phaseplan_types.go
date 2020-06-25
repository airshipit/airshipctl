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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true

// PhasePlan object represents phase execution sequence
type PhasePlan struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	PhaseGroups       []PhaseGroup `json:"phaseGroups,omitempty"`
}

// PhaseGroup represents set of phases (i.e. steps) executed sequentially.
// Phase groups are executed simultaneously
type PhaseGroup struct {
	Name   string           `json:"name,omitempty"`
	Phases []PhaseGroupStep `json:"phases,omitempty"`
}

// PhaseGroupStep represents phase (or step) within phase group
type PhaseGroupStep struct {
	Name string `json:"name,omitempty"`
}
