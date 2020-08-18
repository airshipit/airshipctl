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

package phase

import (
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	airshipv1 "opendev.org/airship/airshipctl/pkg/api/v1alpha1"
	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/environment"
	"opendev.org/airship/airshipctl/pkg/events"
	"opendev.org/airship/airshipctl/pkg/k8s/kubeconfig"
	k8sutils "opendev.org/airship/airshipctl/pkg/k8s/utils"
	"opendev.org/airship/airshipctl/pkg/log"
	"opendev.org/airship/airshipctl/pkg/phase/ifc"
	"opendev.org/airship/airshipctl/pkg/util"
)

// ExecutorRegistry returns map with executor factories
type ExecutorRegistry func() map[schema.GroupVersionKind]ifc.ExecutorFactory

// DefaultExecutorRegistry returns map with executor factories
func DefaultExecutorRegistry() map[schema.GroupVersionKind]ifc.ExecutorFactory {
	execMap := make(map[schema.GroupVersionKind]ifc.ExecutorFactory)
	// add executors here
	return execMap
}

// Cmd object to work with phase api
type Cmd struct {
	DryRun bool

	Registry ExecutorRegistry
	// Will be used to get processor based on executor action
	Processor events.EventProcessor
	*environment.AirshipCTLSettings
}

func (p *Cmd) getBundle() (document.Bundle, error) {
	tp, err := p.AirshipCTLSettings.Config.CurrentContextTargetPath()
	if err != nil {
		return nil, err
	}
	meta, err := p.Config.CurrentContextManifestMetadata()
	if err != nil {
		return nil, err
	}
	log.Debugf("Building phase bundle from path %s", tp)
	return document.NewBundleByPath(filepath.Join(tp, meta.PhaseMeta.Path))
}

func (p *Cmd) getPhaseExecutor(name string) (ifc.Executor, error) {
	phaseConfig, err := p.GetPhase(name)
	if err != nil {
		return nil, err
	}
	return p.GetExecutor(phaseConfig)
}

// GetPhase returns particular phase object identified by name
func (p *Cmd) GetPhase(name string) (*airshipv1.Phase, error) {
	bundle, err := p.getBundle()
	if err != nil {
		return nil, err
	}
	phaseConfig := &airshipv1.Phase{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	selector, err := document.NewSelector().ByObject(phaseConfig, airshipv1.Scheme)
	if err != nil {
		return nil, err
	}
	doc, err := bundle.SelectOne(selector)
	if err != nil {
		return nil, err
	}

	if err = doc.ToAPIObject(phaseConfig, airshipv1.Scheme); err != nil {
		return nil, err
	}
	return phaseConfig, nil
}

// GetExecutor referenced in a phase configuration
func (p *Cmd) GetExecutor(phase *airshipv1.Phase) (ifc.Executor, error) {
	bundle, err := p.getBundle()
	if err != nil {
		return nil, err
	}
	phaseConfig := phase.Config
	// Searching executor configuration document referenced in
	// phase configuration
	refGVK := phaseConfig.ExecutorRef.GroupVersionKind()
	selector := document.NewSelector().
		ByGvk(refGVK.Group, refGVK.Version, refGVK.Kind).
		ByName(phaseConfig.ExecutorRef.Name).
		ByNamespace(phaseConfig.ExecutorRef.Namespace)
	executorDoc, err := bundle.SelectOne(selector)
	if err != nil {
		return nil, err
	}

	// Define executor configuration options
	targetPath, err := p.Config.CurrentContextTargetPath()
	if err != nil {
		return nil, err
	}
	executorDocBundle, err := document.NewBundleByPath(filepath.Join(targetPath, phaseConfig.DocumentEntryPoint))
	if err != nil {
		return nil, err
	}
	if p.Registry == nil {
		p.Registry = DefaultExecutorRegistry
	}
	// Look for executor factory defined in registry
	executorFactory, found := p.Registry()[refGVK]
	if !found {
		return nil, ErrExecutorNotFound{GVK: refGVK}
	}

	kubeConfPath := p.AirshipCTLSettings.Config.KubeConfigPath()
	homeDir := util.UserHomeDir()
	workDir := filepath.Join(homeDir, config.AirshipConfigDir)
	fs := document.NewDocumentFs()
	source := kubeconfig.FromFile(kubeConfPath, fs)
	fileOption := kubeconfig.InjectFilePath(kubeConfPath, fs)
	tempRootOption := kubeconfig.InjectTempRoot(workDir)
	kubeConfig := kubeconfig.NewKubeConfig(source, fileOption, tempRootOption)

	// TODO add function to decide on how to build kubeconfig instead of hardcoding it here,
	// when more kubeconfigs sources are available.
	return executorFactory(ifc.ExecutorConfig{
		ExecutorBundle:   executorDocBundle,
		PhaseName:        phase.Name,
		ExecutorDocument: executorDoc,
		AirshipSettings:  p.AirshipCTLSettings,
		KubeConfig:       kubeConfig,
	})
}

// Exec starts executor goroutine and processes the events
func (p *Cmd) Exec(name string) error {
	runCh := make(chan events.Event)
	processor := events.NewDefaultProcessor(k8sutils.Streams())
	go func() {
		executor, err := p.getPhaseExecutor(name)
		if err != nil {
			handleError(err, runCh)
			return
		}
		executor.Run(runCh, ifc.RunOptions{
			Debug:  p.Debug,
			DryRun: p.DryRun,
		})
	}()
	return processor.Process(runCh)
}

// Plan shows available phase names
func (p *Cmd) Plan() (map[string][]string, error) {
	bundle, err := p.getBundle()
	if err != nil {
		return nil, err
	}
	plan := &airshipv1.PhasePlan{}
	selector, err := document.NewSelector().ByObject(plan, airshipv1.Scheme)
	if err != nil {
		return nil, err
	}
	doc, err := bundle.SelectOne(selector)
	if err != nil {
		return nil, err
	}

	if err := doc.ToAPIObject(plan, airshipv1.Scheme); err != nil {
		return nil, err
	}

	result := make(map[string][]string)
	for _, phaseGroup := range plan.PhaseGroups {
		phases := make([]string, len(phaseGroup.Phases))
		for i, phase := range phaseGroup.Phases {
			phases[i] = phase.Name
		}
		result[phaseGroup.Name] = phases
	}
	return result, nil
}

func handleError(err error, ch chan events.Event) {
	ch <- events.Event{
		Type: events.ErrorType,
		ErrorEvent: events.ErrorEvent{
			Error: err,
		},
	}
	close(ch)
}
