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

package cloudinit

import (
	"opendev.org/airship/airshipctl/pkg/document"

	"sigs.k8s.io/kustomize/api/resid"
	"sigs.k8s.io/kustomize/api/types"
)

var (
	// Initialize defaults where we expect to find user-data and
	// network config data in manifests
	userDataSelectorDefaults = types.Selector{
		Gvk:           resid.Gvk{Kind: document.SecretKind},
		LabelSelector: document.EphemeralUserDataSelector,
	}
	userDataKeyDefault            = "userData"
	networkConfigSelectorDefaults = types.Selector{
		Gvk:           resid.Gvk{Kind: document.BareMetalHostKind},
		LabelSelector: document.EphemeralHostSelector,
	}
	networkConfigKeyDefault = "networkData"
)

// GetCloudData reads YAML document input and generates cloud-init data for
// ephemeral node.
func GetCloudData(
	docBundle document.Bundle,
	userDataSelector types.Selector,
	userDataKey string,
	networkConfigSelector types.Selector,
	networkConfigKey string,
) (userData []byte, netConf []byte, err error) {
	userDataSelectorFinal, userDataKeyFinal := applyDefaultsAndGetData(
		userDataSelector,
		userDataSelectorDefaults,
		userDataKey,
		userDataKeyDefault,
	)
	userData, err = document.GetSecretData(docBundle, userDataSelectorFinal, userDataKeyFinal)
	if err != nil {
		return nil, nil, err
	}

	netConfSelectorFinal, netConfKeyFinal := applyDefaultsAndGetData(
		networkConfigSelector,
		networkConfigSelectorDefaults,
		networkConfigKey,
		networkConfigKeyDefault,
	)
	netConf, err = getNetworkData(docBundle, netConfSelectorFinal, netConfKeyFinal)
	if err != nil {
		return nil, nil, err
	}

	return userData, netConf, err
}

func applyDefaultsAndGetData(
	docSelector types.Selector,
	docSelectorDefaults types.Selector,
	key string,
	keyDefault string,
) (types.Selector, string) {
	// Assign defaults if there are no user supplied overrides
	if docSelector.Kind == "" &&
		docSelector.Name == "" &&
		docSelector.AnnotationSelector == "" &&
		docSelector.LabelSelector == "" {
		docSelector.Kind = docSelectorDefaults.Kind
		docSelector.LabelSelector = docSelectorDefaults.LabelSelector
	}

	keyFinal := key
	if key == "" {
		keyFinal = keyDefault
	}

	return docSelector, keyFinal
}

func getNetworkData(
	docBundle document.Bundle,
	netCfgSelector types.Selector,
	netCfgKey string,
) ([]byte, error) {
	// find the baremetal host indicated as the ephemeral node
	selector := document.NewSelector().ByKind(netCfgSelector.Kind).ByLabel(netCfgSelector.LabelSelector)
	d, err := docBundle.SelectOne(selector)
	if err != nil {
		return nil, err
	}

	// try and find these documents in our bundle
	selector, err = document.NewNetworkDataSelector(d)
	if err != nil {
		return nil, err
	}
	d, err = docBundle.SelectOne(selector)

	if err != nil {
		return nil, err
	}

	// finally, try and retrieve the data we want from the document
	netData, err := document.DecodeSecretData(d, netCfgKey)
	if err != nil {
		return nil, err
	}

	return netData, nil
}
