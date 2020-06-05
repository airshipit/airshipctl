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
	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/environment"
	"opendev.org/airship/airshipctl/pkg/phase/ifc"
)

const (
	// PhaseDirName directory for bundle with phases
	// TODO (dukov) Remove this once repository metadata is ready
	PhaseDirName = "phases"
)

var (
	// ExecutorRegistry contins registered runner factories
	ExecutorRegistry = make(map[schema.GroupVersionKind]ifc.ExecutorFactory)
)

// Cmd object to work with phase api
type Cmd struct {
	*environment.AirshipCTLSettings
	DryRun bool
}

func (p *Cmd) getBundle() (document.Bundle, error) {
	ccm, err := p.Config.CurrentContextManifest()
	if err != nil {
		return nil, err
	}
	return document.NewBundleByPath(filepath.Join(ccm.TargetPath, ccm.SubPath, PhaseDirName))
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
	doc, err := bundle.SelectOne(selector)
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

	// Look for executor factory defined in registry
	executorFactory, found := ExecutorRegistry[refGVK]
	if !found {
		return nil, ErrExecutorNotFound{GVK: refGVK}
	}
	return executorFactory(doc, executorDocBundle, p.AirshipCTLSettings)
}

// Exec particular phase
func (p *Cmd) Exec(name string) error {
	phaseConfig, err := p.GetPhase(name)
	if err != nil {
		return err
	}

	executor, err := p.GetExecutor(phaseConfig)
	if err != nil {
		return err
	}

	return executor.Run(p.DryRun, p.Debug)
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
