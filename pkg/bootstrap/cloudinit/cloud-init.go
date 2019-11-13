package cloudinit

import (
	b64 "encoding/base64"

	"opendev.org/airship/airshipctl/pkg/document"
)

const (
	// TODO (dukov) This should depend on cluster api version once it is
	// fully available for Metal3. In other words:
	// - Sectet for v1alpha1
	// - KubeAdmConfig for v1alpha2
	EphemeralClusterConfKind = "Secret"
)

func decodeData(cfg document.Document, key string) ([]byte, error) {
	data, err := cfg.GetStringMap("data")
	if err != nil {
		return nil, ErrDataNotSupplied{DocName: cfg.GetName(), Key: key}
	}

	res, ok := data[key]
	if !ok {
		return nil, ErrDataNotSupplied{DocName: cfg.GetName(), Key: key}
	}

	return b64.StdEncoding.DecodeString(res)
}

// getDataFromSecret extracts data from Secret with respect to overrides
func getDataFromSecret(cfg document.Document, key string) ([]byte, error) {
	data, err := cfg.GetStringMap("stringData")
	if err != nil {
		return decodeData(cfg, key)
	}

	res, ok := data[key]
	if !ok {
		return decodeData(cfg, key)
	}
	return []byte(res), nil
}

// GetCloudData reads YAML document input and generates cloud-init data for
// node (i.e. Cluster API Machine) with bootstrap annotation.
func GetCloudData(docBundle document.Bundle, bsAnnotation string) ([]byte, []byte, error) {
	var userData []byte
	var netConf []byte
	docs, err := docBundle.GetByAnnotation(bsAnnotation)
	if err != nil {
		return nil, nil, err
	}
	var ephemeralCfg document.Document
	for _, doc := range docs {
		if doc.GetKind() == EphemeralClusterConfKind {
			ephemeralCfg = doc
			break
		}
	}
	if ephemeralCfg == nil {
		return nil, nil, document.ErrDocNotFound{
			Annotation: bsAnnotation,
			Kind:       EphemeralClusterConfKind,
		}
	}

	netConf, err = getDataFromSecret(ephemeralCfg, "netconfig")
	if err != nil {
		return nil, nil, err
	}

	userData, err = getDataFromSecret(ephemeralCfg, "userdata")
	if err != nil {
		return nil, nil, err
	}
	return userData, netConf, nil
}
