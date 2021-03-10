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
	goerrors "errors"
	"path/filepath"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"opendev.org/airship/airshipctl/pkg/api/v1alpha1"
	"opendev.org/airship/airshipctl/pkg/cluster/clustermap"
	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/inventory"
	inventoryifc "opendev.org/airship/airshipctl/pkg/inventory/ifc"
	"opendev.org/airship/airshipctl/pkg/log"
	"opendev.org/airship/airshipctl/pkg/phase/executors/errors"
	"opendev.org/airship/airshipctl/pkg/phase/ifc"
)

// Helper provides functions built around phase bundle to filter and build documents
type Helper struct {
	phaseBundleRoot         string
	inventoryRoot           string
	targetPath              string
	phaseRepoDir            string
	phaseEntryPointBasePath string

	inventory inventoryifc.Inventory
	metadata  *config.Metadata
	config    *config.Config
}

// NewHelper constructs metadata interface based on config
func NewHelper(cfg *config.Config) (ifc.Helper, error) {
	helper := &Helper{
		config: cfg,
	}

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
	helper.inventory = inventory.NewInventory(func() (*config.Config, error) { return cfg, nil })
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
func (helper *Helper) Plan(planID ifc.ID) (*v1alpha1.PhasePlan, error) {
	bundle, err := document.NewBundleByPath(helper.phaseBundleRoot)
	if err != nil {
		return nil, err
	}

	plan := &v1alpha1.PhasePlan{
		ObjectMeta: v1.ObjectMeta{
			Name:      planID.Name,
			Namespace: planID.Namespace,
		},
	}
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
func (helper *Helper) ListPhases(o ifc.ListPhaseOptions) ([]*v1alpha1.Phase, error) {
	bundle, err := document.NewBundleByPath(helper.phaseBundleRoot)
	if err != nil {
		return nil, err
	}

	phase := &v1alpha1.Phase{}
	selector, err := document.NewSelector().ByObject(phase, v1alpha1.Scheme)
	if err != nil {
		return nil, err
	}

	bundle, err = bundle.SelectBundle(selector)
	if err != nil {
		return nil, err
	}

	if o.ClusterName != "" {
		if bundle, err = bundle.SelectByFieldValue("metadata.clusterName", func(v interface{}) bool {
			if field, ok := v.(string); ok {
				return field == o.ClusterName
			}
			return false
		}); err != nil {
			return nil, err
		}
	}

	var docs []document.Document
	if o.PlanID.Name != "" {
		if docs, err = helper.getDocsByPhasePlan(o.PlanID, bundle); err != nil {
			return nil, err
		}
	} else if docs, err = bundle.GetAllDocuments(); err != nil {
		return nil, err
	}

	phases := make([]*v1alpha1.Phase, 0)
	for _, doc := range docs {
		p := v1alpha1.DefaultPhase()
		if err = doc.ToAPIObject(p, v1alpha1.Scheme); err != nil {
			return nil, err
		}
		phases = append(phases, p)
	}
	return phases, nil
}

func (helper *Helper) getDocsByPhasePlan(planID ifc.ID, bundle document.Bundle) ([]document.Document, error) {
	docs := make([]document.Document, 0)
	plan, filterErr := helper.Plan(planID)
	if filterErr != nil {
		return nil, filterErr
	}
	for _, phaseStep := range plan.Phases {
		p := &v1alpha1.Phase{
			ObjectMeta: v1.ObjectMeta{
				Name: phaseStep.Name,
			},
		}
		selector, filterErr := document.NewSelector().ByObject(p, v1alpha1.Scheme)
		if filterErr != nil {
			return nil, filterErr
		}

		doc, filterErr := bundle.SelectOne(selector)
		if filterErr != nil {
			if goerrors.As(filterErr, &document.ErrDocNotFound{}) {
				log.Debug(filterErr.Error())
				continue
			}
			return nil, filterErr
		}

		docs = append(docs, doc)
	}
	return docs, nil
}

// ListPlans returns all phases associated with manifest
func (helper *Helper) ListPlans() ([]*v1alpha1.PhasePlan, error) {
	bundle, err := document.NewBundleByPath(helper.phaseBundleRoot)
	if err != nil {
		return nil, err
	}

	plan := &v1alpha1.PhasePlan{}
	selector, err := document.NewSelector().ByObject(plan, v1alpha1.Scheme)
	if err != nil {
		return nil, err
	}

	docs, err := bundle.Select(selector)
	if err != nil {
		return nil, err
	}

	plans := make([]*v1alpha1.PhasePlan, len(docs))
	for i, doc := range docs {
		p := &v1alpha1.PhasePlan{}
		if err = doc.ToAPIObject(p, v1alpha1.Scheme); err != nil {
			return nil, err
		}
		plans[i] = p
	}
	return plans, nil
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
		return nil, errors.ErrExecutorRefNotDefined{PhaseName: phaseID.Name, PhaseNamespace: phaseID.Namespace}
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

// WorkDir return working directory for aisrhipctl, creates it, if doesn't exist
func (helper *Helper) WorkDir() (string, error) {
	return helper.config.WorkDir()
}

// Inventory return inventory interface
func (helper *Helper) Inventory() (inventoryifc.Inventory, error) {
	return helper.inventory, nil
}
