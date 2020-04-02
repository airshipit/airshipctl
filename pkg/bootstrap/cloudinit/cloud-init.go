package cloudinit

import (
	b64 "encoding/base64"

	"opendev.org/airship/airshipctl/pkg/document"
)

const (
	networkDataKey = "networkData"
	userDataKey    = "userData"
)

// GetCloudData reads YAML document input and generates cloud-init data for
// ephemeral node.
func GetCloudData(docBundle document.Bundle) (userData []byte, netConf []byte, err error) {
	userData, err = getUserData(docBundle)

	if err != nil {
		return nil, nil, err
	}

	netConf, err = getNetworkData(docBundle)

	if err != nil {
		return nil, nil, err
	}

	return userData, netConf, err
}

func getUserData(docBundle document.Bundle) ([]byte, error) {
	// find the user-data document
	selector := document.NewEphemeralCloudDataSelector()
	userDataDoc, err := docBundle.SelectOne(selector)
	if err != nil {
		return nil, err
	}

	// finally, try and retrieve the data we want from the document
	userData, err := decodeData(userDataDoc, userDataKey)
	if err != nil {
		return nil, err
	}

	return userData, nil
}

func getNetworkData(docBundle document.Bundle) ([]byte, error) {
	// find the baremetal host indicated as the ephemeral node
	selector := document.NewEphemeralBMHSelector()
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
	netData, err := decodeData(d, networkDataKey)
	if err != nil {
		return nil, err
	}

	return netData, nil
}

func decodeData(cfg document.Document, key string) ([]byte, error) {
	var needsBase64Decode = false

	// TODO(alanmeadows): distinguish between missing net-data key
	// and missing data/stringData keys in the Secret
	data, err := cfg.GetStringMap("data")
	if err == nil {
		needsBase64Decode = true
	} else {
		// we'll catch any error below
		data, err = cfg.GetStringMap("stringData")
		if err != nil {
			return nil, ErrDataNotSupplied{DocName: cfg.GetName(), Key: "data or stringData"}
		}
	}

	res, ok := data[key]
	if !ok {
		return nil, ErrDataNotSupplied{DocName: cfg.GetName(), Key: key}
	}

	if needsBase64Decode {
		return b64.StdEncoding.DecodeString(res)
	}
	return []byte(res), nil
}
