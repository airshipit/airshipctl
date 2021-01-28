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

package inventory

import (
	"path/filepath"

	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/inventory/baremetal"
	"opendev.org/airship/airshipctl/pkg/inventory/ifc"
)

var _ ifc.Inventory = Invetnory{}

// Invetnory implementation of the interface
type Invetnory struct {
	config.Factory
}

// NewInventory inventory constructor
func NewInventory(f config.Factory) ifc.Inventory {
	return Invetnory{
		Factory: f,
	}
}

// BaremetalInventory implementation of the interface
func (i Invetnory) BaremetalInventory() (ifc.BaremetalInventory, error) {
	cfg, err := i.Factory()
	if err != nil {
		return nil, err
	}

	mgmCfg, err := cfg.CurrentContextManagementConfig()
	if err != nil {
		return nil, err
	}

	targetPath, err := cfg.CurrentContextTargetPath()
	if err != nil {
		return nil, err
	}

	phaseDir, err := cfg.CurrentContextInventoryRepositoryName()
	if err != nil {
		return nil, err
	}

	metadata, err := cfg.CurrentContextManifestMetadata()
	if err != nil {
		return nil, err
	}

	inventoryBundle := filepath.Join(targetPath, phaseDir, metadata.Inventory.Path)

	bundle, err := document.NewBundleByPath(inventoryBundle)
	if err != nil {
		return nil, err
	}
	return baremetal.NewInventory(mgmCfg, bundle), nil
}
