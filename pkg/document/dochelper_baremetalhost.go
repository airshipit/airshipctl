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

package document

// GetBMHNetworkData retrieves the associated network data string
// for the bmh document supplied from the bundle supplied
func GetBMHNetworkData(bmh Document, bundle Bundle) (string, error) {
	// try and find these documents in our bundle
	selector, err := NewNetworkDataSelector(bmh)
	if err != nil {
		return "", err
	}
	doc, err := bundle.SelectOne(selector)

	if err != nil {
		return "", err
	}

	networkData, err := GetSecretDataKey(doc, "networkData")
	if err != nil {
		return "", err
	}
	return networkData, nil
}

// GetBMHBMCAddress returns the bmc address for a particular the document supplied
func GetBMHBMCAddress(bmh Document) (string, error) {
	bmcAddress, err := bmh.GetString("spec.bmc.address")
	if err != nil {
		return "", err
	}
	return bmcAddress, nil
}

// GetBMHBMCCredentials returns the BMC credentials for the bmh document supplied from
// the supplied bundle
func GetBMHBMCCredentials(bmh Document, bundle Bundle) (username string, password string, err error) {
	// extract the secret document name
	bmcCredentialsName, err := bmh.GetString("spec.bmc.credentialsName")
	if err != nil {
		return "", "", err
	}

	// find the secret within the bundle supplied
	selector := NewBMCCredentialsSelector(bmcCredentialsName)
	doc, err := bundle.SelectOne(selector)
	if err != nil {
		return "", "", err
	}

	username, err = GetSecretDataKey(doc, "username")
	if err != nil {
		return "", "", err
	}
	password, err = GetSecretDataKey(doc, "password")
	if err != nil {
		return "", "", err
	}

	// extract the username and password from them
	return username, password, nil
}
