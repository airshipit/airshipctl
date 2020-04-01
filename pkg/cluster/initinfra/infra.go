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

package initinfra

import (
	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/environment"
	"opendev.org/airship/airshipctl/pkg/k8s/client"
)

// Infra is an abstraction used to initialize base infrastructure
type Infra struct {
	FileSystem   document.FileSystem
	RootSettings *environment.AirshipCTLSettings
	Client       client.Interface

	DryRun      bool
	Prune       bool
	ClusterType string
}

// NewInfra return instance of Infra
func NewInfra(rs *environment.AirshipCTLSettings) *Infra {
	// At this point AirshipCTLSettings may not be fully initialized
	infra := &Infra{RootSettings: rs}
	return infra
}

// Run intinfra subcommand logic
func (infra *Infra) Run() error {
	infra.FileSystem = document.NewDocumentFs()
	var err error
	infra.Client, err = client.NewClient(infra.RootSettings)
	if err != nil {
		return err
	}
	return infra.Deploy()
}

// Deploy method deploys documents
func (infra *Infra) Deploy() error {
	kctl := infra.Client.Kubectl()
	ao, err := kctl.ApplyOptions()
	if err != nil {
		return err
	}

	ao.SetDryRun(infra.DryRun)
	// If prune is true, set selector for purning
	if infra.Prune {
		ao.SetPrune(document.InitInfraSelector)
	}

	globalConf := infra.RootSettings.Config()
	if err = globalConf.EnsureComplete(); err != nil {
		return err
	}

	kustomizePath, err := globalConf.CurrentContextEntryPoint(infra.ClusterType, config.Initinfra)
	if err != nil {
		return err
	}

	b, err := document.NewBundleByPath(kustomizePath)
	if err != nil {
		return err
	}

	// Returns all documents for this phase
	docs, err := b.Select(document.NewDeployToK8sSelector())
	if err != nil {
		return err
	}
	if len(docs) == 0 {
		return document.ErrDocNotFound{}
	}

	return kctl.Apply(docs, ao)
}
