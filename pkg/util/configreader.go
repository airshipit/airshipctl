package util

import (
	"io/ioutil"

	"sigs.k8s.io/yaml"
)

// ReadYAMLFile reads YAML-formatted configuration file and
// de-serializes it to a given object
func ReadYAMLFile(filePath string, cfg interface{}) error {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, cfg)
}
