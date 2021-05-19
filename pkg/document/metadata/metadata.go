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

package metadata

import (
	"io/ioutil"

	apiv1 "opendev.org/airship/airshipctl/pkg/api/v1alpha1"
	"opendev.org/airship/airshipctl/pkg/document"
)

// Metadata defines the site specific Phase properties like
// PhasePath, DocEntryPointPrefix & InventoryPath
type Metadata struct {
	MetadataPhasePath   string
	DocEntryPointPrefix string
	InventoryPath       string
}

// Config returns Metadata with the attributes of Metadata updated
func Config(phaseBundlePath string) (Metadata, error) {
	m := Metadata{}

	data, err := ioutil.ReadFile(phaseBundlePath)
	if err != nil {
		return Metadata{}, err
	}
	bundle, err := document.NewBundleFromBytes(data)
	if err != nil {
		return Metadata{}, err
	}

	metaConfig := &apiv1.ManifestMetadata{}
	selector, err := document.NewSelector().ByObject(metaConfig, apiv1.Scheme)
	if err != nil {
		return Metadata{}, err
	}
	doc, err := bundle.SelectOne(selector)
	if err != nil {
		return Metadata{}, err
	}

	m.MetadataPhasePath, err = doc.GetString("spec.phase.path")
	if err != nil {
		return Metadata{}, err
	}
	m.DocEntryPointPrefix, err = doc.GetString("spec.phase.docEntryPointPrefix")
	if err != nil {
		return Metadata{}, err
	}
	m.InventoryPath, err = doc.GetString("spec.inventory.path")
	if err != nil {
		return Metadata{}, err
	}

	return m, nil
}
