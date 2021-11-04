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
	"opendev.org/airship/airshipctl/pkg/api/v1alpha1"
	"opendev.org/airship/airshipctl/pkg/document"
)

const (
	defaultUserDataKey      = "userData"
	defaultNetworkConfigKey = "networkData"
)

// GetCloudData reads YAML document input and generates cloud-init data for
// ephemeral node.
func GetCloudData(
	docBundle document.Bundle,
	userDataSelector v1alpha1.Selector,
	userDataKey string,
	networkConfigSelector v1alpha1.Selector,
	networkConfigKey string,
) (userData []byte, netConf []byte, err error) {
	uDataSel := document.NewSelectorFromV1Alpha1(userDataSelector)
	nwDataSel := document.NewSelectorFromV1Alpha1(networkConfigSelector)
	userDataSelectorFinal, userDataKeyFinal := applyDefaultsAndGetData(
		uDataSel,
		document.SecretKind,
		document.EphemeralUserDataSelector,
		userDataKey,
		defaultUserDataKey,
	)
	userData, err = getUserData(docBundle, userDataSelectorFinal, userDataKeyFinal)
	if err != nil {
		return nil, nil, err
	}

	netConfSelectorFinal, netConfKeyFinal := applyDefaultsAndGetData(
		nwDataSel,
		document.BareMetalHostKind,
		document.EphemeralHostSelector,
		networkConfigKey,
		defaultNetworkConfigKey,
	)
	netConf, err = getNetworkData(docBundle, netConfSelectorFinal, netConfKeyFinal)
	if err != nil {
		return nil, nil, err
	}

	return userData, netConf, err
}

func applyDefaultsAndGetData(
	docSelector document.Selector,
	defaultKind string,
	defaultLabel string,
	key string,
	keyDefault string,
) (document.Selector, string) {
	// Assign defaults if there are no user supplied overrides
	if docSelector.Kind == "" &&
		docSelector.Name == "" &&
		docSelector.AnnotationSelector == "" &&
		docSelector.LabelSelector == "" {
		docSelector = docSelector.ByKind(defaultKind).ByLabel(defaultLabel)
	}

	keyFinal := key
	if key == "" {
		keyFinal = keyDefault
	}

	return docSelector, keyFinal
}

func getUserData(
	docBundle document.Bundle,
	userDataSelector document.Selector,
	userDataKey string,
) ([]byte, error) {
	doc, err := docBundle.SelectOne(userDataSelector)
	if err != nil {
		return nil, err
	}

	data, err := document.GetSecretDataKey(doc, userDataKey)
	if err != nil {
		return nil, err
	}

	return []byte(data), nil
}

func getNetworkData(
	docBundle document.Bundle,
	netCfgSelector document.Selector,
	netCfgKey string,
) ([]byte, error) {
	// find the baremetal host indicated as the ephemeral node
	d, err := docBundle.SelectOne(netCfgSelector)
	if err != nil {
		return nil, err
	}

	// try and find these documents in our bundle
	selector, err := document.NewNetworkDataSelector(d)
	if err != nil {
		return nil, err
	}
	d, err = docBundle.SelectOne(selector)

	if err != nil {
		return nil, err
	}

	// finally, try and retrieve the data we want from the document
	netData, err := document.GetSecretDataKey(d, netCfgKey)
	if err != nil {
		return nil, err
	}

	return []byte(netData), nil
}
