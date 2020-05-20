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

package apply

import (
	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/environment"
	"opendev.org/airship/airshipctl/pkg/k8s/client"
)

// Options is an abstraction used to apply the phase
type Options struct {
	RootSettings *environment.AirshipCTLSettings
	Client       client.Interface

	DryRun    bool
	Prune     bool
	PhaseName string
}

// NewOptions return instance of Options
func NewOptions(settings *environment.AirshipCTLSettings) *Options {
	// At this point AirshipCTLSettings may not be fully initialized
	applyOptions := &Options{RootSettings: settings}
	return applyOptions
}

// Run apply subcommand logic
func (applyOptions *Options) Run() error {
	kctl := applyOptions.Client.Kubectl()
	ao, err := kctl.ApplyOptions()
	if err != nil {
		return err
	}

	ao.SetDryRun(applyOptions.DryRun)
	// If prune is true, set selector for pruning
	if applyOptions.Prune {
		ao.SetPrune(document.ApplyPhaseSelector + applyOptions.PhaseName)
	}

	globalConf := applyOptions.RootSettings.Config

	if err = globalConf.EnsureComplete(); err != nil {
		return err
	}

	kustomizePath, err := globalConf.CurrentContextEntryPoint(applyOptions.PhaseName)
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
