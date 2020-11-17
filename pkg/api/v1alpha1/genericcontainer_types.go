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
	"sigs.k8s.io/kustomize/kyaml/fn/runtime/runtimeutil"
)

// +kubebuilder:object:root=true

// GenericContainer provides info about generic container
type GenericContainer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// If set to will print output of RunFns to Stdout
	PrintOutput bool `json:"printOutput,omitempty"`
	// Settings for for a container
	Spec runtimeutil.FunctionSpec `json:"spec,omitempty"`
	// Config for the RunFns function in a custom format
	Config string `json:"config,omitempty"`
}

// DefaultGenericContainer can be used to safely unmarshal GenericContainer object without nil pointers
func DefaultGenericContainer() *GenericContainer {
	return &GenericContainer{
		Spec: runtimeutil.FunctionSpec{},
	}
}

// DeepCopyInto is copying the receiver, writing into out. in must be non-nil.
func (in *GenericContainer) DeepCopyInto(out *GenericContainer) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)

	out.Spec = in.Spec
	out.Spec.Container = in.Spec.Container
	out.Spec.Container.Network = in.Spec.Container.Network
	if in.Spec.Container.StorageMounts != nil {
		in, out := &in.Spec.Container.StorageMounts, &out.Spec.Container.StorageMounts
		*out = make([]runtimeutil.StorageMount, len(*in))
		copy(*out, *in)
	}
	if in.Spec.Container.Env != nil {
		in, out := &in.Spec.Container.Env, &out.Spec.Container.Env
		*out = make([]string, len(*in))
		copy(*out, *in)
	}

	out.Spec.Starlark = in.Spec.Starlark
	out.Spec.Exec = in.Spec.Exec
	if in.Spec.StorageMounts != nil {
		in, out := &in.Spec.StorageMounts, &out.Spec.StorageMounts
		*out = make([]runtimeutil.StorageMount, len(*in))
		copy(*out, *in)
	}
}
