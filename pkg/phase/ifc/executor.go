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

package ifc

import (
	"io"
	"time"

	"opendev.org/airship/airshipctl/pkg/cluster/clustermap"
	"opendev.org/airship/airshipctl/pkg/container"
	"opendev.org/airship/airshipctl/pkg/document"
	inventoryifc "opendev.org/airship/airshipctl/pkg/inventory/ifc"
	"opendev.org/airship/airshipctl/pkg/k8s/kubeconfig"
)

// Executor interface should be implemented by each runner
type Executor interface {
	Run(RunOptions) error
	Render(io.Writer, RenderOptions) error
	Validate() error
	Status() (ExecutorStatus, error)
}

// ExecutorStatus is a struct which defines the status
type ExecutorStatus struct{}

// RunOptions holds options for run method
type RunOptions struct {
	DryRun  bool
	Timeout *time.Duration
}

// PlanRunOptions holds options for plan run method
type PlanRunOptions struct {
	RunOptions
	ResumeFromPhase string
}

// RenderOptions holds options for render method
type RenderOptions struct {
	FilterSelector document.Selector
}

// ExecutorFactory for executor instantiation
type ExecutorFactory func(config ExecutorConfig) (Executor, error)

// ExecutorConfig container to store all executor options
type ExecutorConfig struct {
	PhaseName    string
	ClusterName  string
	SinkBasePath string
	TargetPath   string

	ClusterMap        clustermap.ClusterMap
	ExecutorDocument  document.Document
	KubeConfig        kubeconfig.Interface
	BundleFactory     document.BundleFactoryFunc
	PhaseConfigBundle document.Bundle
	Inventory         inventoryifc.Inventory
	ContainerFunc     container.ClientV1Alpha1FactoryFunc
}
