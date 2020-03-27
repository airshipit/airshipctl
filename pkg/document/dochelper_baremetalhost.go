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
