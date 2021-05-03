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
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// GenericContainerAirshipDockerDriver is the driver name supported by airship container interface
	// we dont use strong type here for now, to avoid converting to string in the implementation
	GenericContainerAirshipDockerDriver = "docker"
	// GenericContainerTypeAirship specifies that airship type container will be used
	GenericContainerTypeAirship GenericContainerType = "airship"
	// GenericContainerTypeKrm specifies that kustomize krm function will be used
	GenericContainerTypeKrm GenericContainerType = "krm"
	// KubeConfigEnvKey uses as a key for kubeconfig env variable
	KubeConfigEnvKey = "KUBECONFIG"
	// KubeConfigPath is a path for mounted kubeconfig inside container
	KubeConfigPath = "/kubeconfig"
	// KubeConfigEnvKeyContext uses as a key for kubectl context env variable
	KubeConfigEnvKeyContext = "KCTL_CONTEXT"
	// KubeConfigEnv uses as a kubeconfig env variable
	KubeConfigEnv = KubeConfigEnvKey + "=" + KubeConfigPath

	// ValidatorPreventCleanup is an env variable that prevents validator to clean up its working directory after finish
	ValidatorPreventCleanup = "VALIDATOR_PREVENT_CLEANUP"
)

// +kubebuilder:object:root=true

// GenericContainer provides info about generic container
type GenericContainer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Holds container configuration
	Spec GenericContainerSpec `json:"spec,omitempty"`

	// Config will be passed to stdin of the container together with other objects
	// more information on easy ways to consume the config can be found here
	// https://googlecontainertools.github.io/kpt/guides/producer/functions/golang/
	Config string `json:"config,omitempty"`
	// Reference is a reference to a configuration object, that must reside in the same
	// bundle as this GenericContainer object, if specified, Config string will be
	// ignored and referenced object in ConfigRef will be used into the Config string
	// instead and passed further into the container stdin
	ConfigRef *v1.ObjectReference `json:"configRef,omitempty"`
}

// GenericContainerType specify type of the container, there are currently two types:
// airship - airship will run the container
// krm - kustomize krm function will run the container
type GenericContainerType string

// GenericContainerSpec container configuration
type GenericContainerSpec struct {
	// Supported types are "airship" and "krm"
	Type GenericContainerType `json:"type,omitempty"`

	// Airship container spec
	Airship AirshipContainerSpec `json:"airship,omitempty"`

	// KRM container function spec
	KRM KRMContainerSpec `json:"krm,omitempty"`

	// Executor will write output using kustomize sink if this parameter is specified.
	// Else it will write output to STDOUT.
	// This path relative to current site root.
	SinkOutputDir string `json:"sinkOutputDir,omitempty"`

	// HostNetwork defines network specific configuration
	HostNetwork bool `json:"hostNetwork,omitempty" yaml:"network,omitempty"`

	// Image is the container image to run
	Image string `json:"image,omitempty" yaml:"image,omitempty"`

	// EnvVars is a slice of env string that will be exposed to container
	// ["MY_VAR=my-value, "MY_VAR1=my-value1"]
	// if passed in format ["MY_ENV"] this env variable will be exported the container
	EnvVars []string `json:"envVars,omitempty"`

	// Mounts are the storage or directories to mount into the container
	StorageMounts []StorageMount `json:"mounts,omitempty" yaml:"mounts,omitempty"`
}

// AirshipContainerSpec airship container settings
type AirshipContainerSpec struct {

	// ContainerRuntime currently supported and default runtime is "docker"
	ContainerRuntime string `json:"containerRuntime,omitempty"`

	// Cmd to run inside the container, `["/my-command", "arg"]`
	Cmd []string `json:"cmd,omitempty"`

	// Privileged identifies if the container is to be run in a Privileged mode
	Privileged bool `json:"privileged,omitempty"`
}

// KRMContainerSpec defines a spec for running a function as a container
// empty for now since it has no extra fields from AirshipContainerSpec
type KRMContainerSpec struct{}

// StorageMount represents a container's mounted storage option(s)
// copy from https://github.com/kubernetes-sigs/kustomize to avoid imports in this package
type StorageMount struct {
	// Type of mount e.g. bind mount, local volume, etc.
	MountType string `json:"type,omitempty" yaml:"type,omitempty"`

	// Source for the storage to be mounted.
	// For named volumes, this is the name of the volume.
	// For anonymous volumes, this field is omitted (empty string).
	// For bind mounts, this is the path to the file or directory on the host.
	// If provided path is relative, it will be expanded to absolute one by following patterns:
	// - if starts with '~/' or contains only '~' : $HOME + Src
	// - in other cases : TargetPath + Src
	Src string `json:"src,omitempty" yaml:"src,omitempty"`

	// The path where the file or directory is mounted in the container.
	DstPath string `json:"dst,omitempty" yaml:"dst,omitempty"`

	// Mount in ReadWrite mode if it's explicitly configured
	// See https://docs.docker.com/storage/bind-mounts/#use-a-read-only-bind-mount
	ReadWriteMode bool `json:"rw,omitempty" yaml:"rw,omitempty"`
}

// DefaultGenericContainer can be used to safely unmarshal GenericContainer object without nil pointers
func DefaultGenericContainer() *GenericContainer {
	return &GenericContainer{}
}
