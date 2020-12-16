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
	"fmt"
	"io"
	"path/filepath"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"opendev.org/airship/airshipctl/pkg/api/v1alpha1"
	"opendev.org/airship/airshipctl/pkg/cluster/clustermap"
	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/phase/ifc"
	"opendev.org/airship/airshipctl/pkg/util"
)

// Helper provides functions built around phase bundle to filter and build documents
type Helper struct {
	phaseBundleRoot         string
	inventoryRoot           string
	targetPath              string
	phaseRepoDir            string
	phaseEntryPointBasePath string
	metadata                *config.Metadata
}

// NewHelper constructs metadata interface based on config
func NewHelper(cfg *config.Config) (ifc.Helper, error) {
	helper := &Helper{}

	var err error
	helper.targetPath, err = cfg.CurrentContextTargetPath()
	if err != nil {
		return nil, err
	}
	helper.phaseRepoDir, err = cfg.CurrentContextPhaseRepositoryDir()
	if err != nil {
		return nil, err
	}
	helper.metadata, err = cfg.CurrentContextManifestMetadata()
	if err != nil {
		return nil, err
	}
	helper.phaseBundleRoot = filepath.Join(helper.targetPath, helper.phaseRepoDir, helper.metadata.PhaseMeta.Path)
	helper.inventoryRoot = filepath.Join(helper.targetPath, helper.phaseRepoDir, helper.metadata.Inventory.Path)
	helper.phaseEntryPointBasePath = filepath.Join(helper.targetPath, helper.phaseRepoDir,
		helper.metadata.PhaseMeta.DocEntryPointPrefix)
	return helper, nil
}

// Phase returns a phase APIObject based on phase selector
func (helper *Helper) Phase(phaseID ifc.ID) (*v1alpha1.Phase, error) {
	bundle, err := document.NewBundleByPath(helper.phaseBundleRoot)
	if err != nil {
		return nil, err
	}
	phase := &v1alpha1.Phase{
		ObjectMeta: v1.ObjectMeta{
			Name:      phaseID.Name,
			Namespace: phaseID.Namespace,
		},
	}
	selector, err := document.NewSelector().ByObject(phase, v1alpha1.Scheme)
	if err != nil {
		return nil, err
	}
	doc, err := bundle.SelectOne(selector)
	if err != nil {
		return nil, err
	}
	// Overwrite phase used for selector, with a phase with default values
	phase = v1alpha1.DefaultPhase()
	if err = doc.ToAPIObject(phase, v1alpha1.Scheme); err != nil {
		return nil, err
	}
	return phase, nil
}

// Plan returns plan associated with a manifest
func (helper *Helper) Plan() (*v1alpha1.PhasePlan, error) {
	bundle, err := document.NewBundleByPath(helper.phaseBundleRoot)
	if err != nil {
		return nil, err
	}

	plan := &v1alpha1.PhasePlan{}
	selector, err := document.NewSelector().ByObject(plan, v1alpha1.Scheme)
	if err != nil {
		return nil, err
	}

	doc, err := bundle.SelectOne(selector)
	if err != nil {
		return nil, err
	}

	if err := doc.ToAPIObject(plan, v1alpha1.Scheme); err != nil {
		return nil, err
	}
	return plan, nil
}

// ListPhases returns all phases associated with manifest
func (helper *Helper) ListPhases() ([]*v1alpha1.Phase, error) {
	bundle, err := document.NewBundleByPath(helper.phaseBundleRoot)
	if err != nil {
		return nil, err
	}

	phase := &v1alpha1.Phase{}
	selector, err := document.NewSelector().ByObject(phase, v1alpha1.Scheme)
	if err != nil {
		return nil, err
	}

	docs, err := bundle.Select(selector)
	if err != nil {
		return nil, err
	}

	phases := []*v1alpha1.Phase{}
	for _, doc := range docs {
		p := v1alpha1.DefaultPhase()
		if err = doc.ToAPIObject(p, v1alpha1.Scheme); err != nil {
			return nil, err
		}
		phases = append(phases, p)
	}
	return phases, nil
}

// ClusterMapAPIobj associated with the the manifest
func (helper *Helper) ClusterMapAPIobj() (*v1alpha1.ClusterMap, error) {
	bundle, err := document.NewBundleByPath(helper.phaseBundleRoot)
	if err != nil {
		return nil, err
	}

	cMap := v1alpha1.DefaultClusterMap()
	selector, err := document.NewSelector().ByObject(cMap, v1alpha1.Scheme)
	if err != nil {
		return nil, err
	}

	doc, err := bundle.SelectOne(selector)
	if err != nil {
		return nil, err
	}

	if err = doc.ToAPIObject(cMap, v1alpha1.Scheme); err != nil {
		return nil, err
	}
	return cMap, nil
}

// ClusterMap associated with the the manifest
func (helper *Helper) ClusterMap() (clustermap.ClusterMap, error) {
	cMap, err := helper.ClusterMapAPIobj()
	if err != nil {
		return nil, err
	}
	return clustermap.NewClusterMap(cMap), nil
}

// ExecutorDoc returns executor document associated with phase
func (helper *Helper) ExecutorDoc(phaseID ifc.ID) (document.Document, error) {
	bundle, err := document.NewBundleByPath(helper.phaseBundleRoot)
	if err != nil {
		return nil, err
	}
	phaseObj, err := helper.Phase(phaseID)
	if err != nil {
		return nil, err
	}
	phaseConfig := phaseObj.Config

	if phaseConfig.ExecutorRef == nil {
		return nil, ErrExecutorRefNotDefined{PhaseName: phaseID.Name, PhaseNamespace: phaseID.Namespace}
	}

	// Searching executor configuration document referenced in
	// phase configuration
	refGVK := phaseConfig.ExecutorRef.GroupVersionKind()
	selector := document.NewSelector().
		ByGvk(refGVK.Group, refGVK.Version, refGVK.Kind).
		ByName(phaseConfig.ExecutorRef.Name).
		ByNamespace(phaseConfig.ExecutorRef.Namespace)
	return bundle.SelectOne(selector)
}

// TargetPath returns manifest root
func (helper *Helper) TargetPath() string {
	return helper.targetPath
}

// PhaseRepoDir returns the last part of the repo url
// E.g. http://dummy.org/reponame.git -> reponame
func (helper *Helper) PhaseRepoDir() string {
	return helper.phaseRepoDir
}

// DocEntryPointPrefix returns the prefix which if not empty is prepended to the
// DocumentEntryPoint field in the phase struct
// so the full entry point is DocEntryPointPrefix + DocumentEntryPoint
func (helper *Helper) DocEntryPointPrefix() string {
	return helper.metadata.PhaseMeta.DocEntryPointPrefix
}

// PhaseBundleRoot returns path to document root with phase documents
func (helper *Helper) PhaseBundleRoot() string {
	return helper.phaseBundleRoot
}

// PhaseEntryPointBasePath returns path to current site directory
func (helper *Helper) PhaseEntryPointBasePath() string {
	return helper.phaseEntryPointBasePath
}

// WorkDir return manifest root
// TODO add creation of WorkDir if it doesn't exist
func (helper *Helper) WorkDir() (string, error) {
	return filepath.Join(util.UserHomeDir(), config.AirshipConfigDir), nil
}

// PrintPlan prints plan
// TODO make this more readable in the future, and move to client
func PrintPlan(plan *v1alpha1.PhasePlan, w io.Writer) error {
	result := make(map[string][]string)
	for _, phaseGroup := range plan.PhaseGroups {
		phases := make([]string, len(phaseGroup.Phases))
		for i, phase := range phaseGroup.Phases {
			phases[i] = phase.Name
		}
		result[phaseGroup.Name] = phases
	}

	tw := util.NewTabWriter(w)
	defer tw.Flush()
	fmt.Fprintf(tw, "GROUP\tPHASE\n")
	for group, phaseList := range result {
		fmt.Fprintf(tw, "%s\t\n", group)
		for _, phase := range phaseList {
			fmt.Fprintf(tw, "\t%s\n", phase)
		}
	}
	return nil
}
