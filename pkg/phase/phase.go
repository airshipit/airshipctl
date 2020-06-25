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

	airshipv1 "opendev.org/airship/airshipctl/pkg/api/v1alpha1"
	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/environment"
)

const (
	// PhaseDirName directory for bundle with phases
	// TODO (dukov) Remove this once repository metadata is ready
	PhaseDirName = "phases"
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
