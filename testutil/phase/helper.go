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
	"github.com/stretchr/testify/mock"

	"opendev.org/airship/airshipctl/pkg/api/v1alpha1"
	"opendev.org/airship/airshipctl/pkg/cluster/clustermap"
	"opendev.org/airship/airshipctl/pkg/document"
	inventoryifc "opendev.org/airship/airshipctl/pkg/inventory/ifc"
	"opendev.org/airship/airshipctl/pkg/phase/ifc"
)

var _ ifc.Helper = &MockHelper{}

// MockHelper mock ifc.Helper interface
type MockHelper struct {
	mock.Mock
}

// TargetPath mock
func (mh *MockHelper) TargetPath() string {
	args := mh.Called()
	return args.Get(0).(string)
}

// PhaseRepoDir mock
func (mh *MockHelper) PhaseRepoDir() string {
	args := mh.Called()
	return args.Get(0).(string)
}

// DocEntryPointPrefix mock
func (mh *MockHelper) DocEntryPointPrefix() string {
	args := mh.Called()
	return args.Get(0).(string)
}

// WorkDir mock
func (mh *MockHelper) WorkDir() string {
	args := mh.Called()
	return args.Get(0).(string)
}

// Phase mock
func (mh *MockHelper) Phase(id ifc.ID) (*v1alpha1.Phase, error) {
	args := mh.Called(id)
	val, ok := args.Get(0).(*v1alpha1.Phase)
	if !ok {
		return nil, args.Error(1)
	}
	return val, args.Error(1)
}

// Plan mock
func (mh *MockHelper) Plan(id ifc.ID) (*v1alpha1.PhasePlan, error) {
	args := mh.Called(id)
	val, ok := args.Get(0).(*v1alpha1.PhasePlan)
	if !ok {
		return nil, args.Error(1)
	}
	return val, args.Error(1)
}

// ListPhases mock
func (mh *MockHelper) ListPhases(o ifc.ListPhaseOptions) ([]*v1alpha1.Phase, error) {
	args := mh.Called(o)
	val, ok := args.Get(0).([]*v1alpha1.Phase)
	if !ok {
		return nil, args.Error(1)
	}
	return val, args.Error(1)
}

// ListPlans mock
func (mh *MockHelper) ListPlans() ([]*v1alpha1.PhasePlan, error) {
	args := mh.Called()
	val, ok := args.Get(0).([]*v1alpha1.PhasePlan)
	if !ok {
		return nil, args.Error(1)
	}
	return val, args.Error(1)
}

// ClusterMapAPIobj mock
func (mh *MockHelper) ClusterMapAPIobj() (*v1alpha1.ClusterMap, error) {
	args := mh.Called()
	val, ok := args.Get(0).(*v1alpha1.ClusterMap)
	if !ok {
		return nil, args.Error(1)
	}
	return val, args.Error(1)
}

// ClusterMap mock
func (mh *MockHelper) ClusterMap() (clustermap.ClusterMap, error) {
	args := mh.Called()
	val, ok := args.Get(0).(clustermap.ClusterMap)
	if !ok {
		return nil, args.Error(1)
	}
	return val, args.Error(1)
}

// ExecutorDoc mock
func (mh *MockHelper) ExecutorDoc(phaseID ifc.ID) (document.Document, error) {
	args := mh.Called()
	val, ok := args.Get(0).(document.Document)
	if !ok {
		return nil, args.Error(1)
	}
	return val, args.Error(1)
}

// PhaseBundleRoot mock
func (mh *MockHelper) PhaseBundleRoot() string {
	args := mh.Called()
	return args.Get(0).(string)
}

// Inventory mock
func (mh *MockHelper) Inventory() (inventoryifc.Inventory, error) {
	args := mh.Called()
	val, ok := args.Get(0).(inventoryifc.Inventory)
	if !ok {
		return nil, args.Error(1)
	}
	return val, args.Error(1)
}

// PhaseEntryPointBasePath mock
func (mh *MockHelper) PhaseEntryPointBasePath() string {
	args := mh.Called()
	return args.Get(0).(string)
}

// PhaseConfigBundle mock
func (mh *MockHelper) PhaseConfigBundle() document.Bundle {
	args := mh.Called()
	val, ok := args.Get(0).(document.Bundle)
	if !ok {
		return nil
	}
	return val
}
