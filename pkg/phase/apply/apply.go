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
	"fmt"
	"time"

	"sigs.k8s.io/cli-utils/pkg/common"

	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/environment"
	"opendev.org/airship/airshipctl/pkg/events"
	"opendev.org/airship/airshipctl/pkg/k8s/applier"
	"opendev.org/airship/airshipctl/pkg/k8s/utils"
	"opendev.org/airship/airshipctl/pkg/log"
)

// Options is an abstraction used to apply the phase
type Options struct {
	DryRun      bool
	Prune       bool
	PhaseName   string
	WaitTimeout time.Duration

	RootSettings *environment.AirshipCTLSettings
	Applier      *applier.Applier
	Processor    events.EventProcessor
	EventChannel chan events.Event
}

// Initialize Options with required field, such as Applier
func (o *Options) Initialize() {
	f := utils.FactoryFromKubeConfigPath(o.RootSettings.KubeConfigPath)
	streams := utils.Streams()
	o.EventChannel = make(chan events.Event)
	o.Applier = applier.NewApplier(o.EventChannel, f, streams)
	o.Processor = events.NewDefaultProcessor(streams)
}

// Run apply subcommand logic
func (o *Options) Run() error {
	ao := applier.ApplyOptions{
		DryRunStrategy: common.DryRunNone,
		Prune:          o.Prune,
		WaitTimeout:    o.WaitTimeout,
	}
	if o.DryRun {
		ao.DryRunStrategy = common.DryRunClient
	}

	globalConf := o.RootSettings.Config

	if err := globalConf.EnsureComplete(); err != nil {
		return err
	}
	clusterName, err := globalConf.CurrentContextClusterName()
	if err != nil {
		return err
	}
	clusterType, err := globalConf.CurrentContextClusterType()
	if err != nil {
		return err
	}
	ao.BundleName = fmt.Sprintf("%s-%s-%s", clusterName, clusterType, o.PhaseName)
	kustomizePath, err := globalConf.CurrentContextEntryPoint(o.PhaseName)
	if err != nil {
		return err
	}
	log.Debugf("building bundle from kustomize path %s", kustomizePath)
	b, err := document.NewBundleByPath(kustomizePath)
	if err != nil {
		return err
	}
	// Returns all documents for this phase
	bundle, err := b.SelectBundle(document.NewDeployToK8sSelector())
	if err != nil {
		return err
	}
	go o.Applier.ApplyBundle(bundle, ao)
	return o.Processor.Process(o.EventChannel)
}
