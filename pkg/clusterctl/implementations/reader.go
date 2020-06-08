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

package implementations

import (
	clusterctlv1 "sigs.k8s.io/cluster-api/cmd/clusterctl/api/v1alpha3"
	"sigs.k8s.io/cluster-api/cmd/clusterctl/client/config"
	"sigs.k8s.io/yaml"

	airshipv1 "opendev.org/airship/airshipctl/pkg/api/v1alpha1"
)

var _ config.Reader = &AirshipReader{}

const (
	// TODO this must come as as ProviderConfigKey from clusterctl/client/config pkg
	// see https://github.com/kubernetes-sigs/cluster-api/blob/master/cmd/clusterctl/client/config/imagemeta_client.go#L27
	imagesConfigKey = "images"
)

// AirshipReader provides a reader implementation backed by a map
type AirshipReader struct {
	variables map[string]string
}

// configProvider is a mirror of config.Provider, re-implemented here in order to
// avoid circular dependencies between pkg/client/config and pkg/internal/test
type configProvider struct {
	Name string                    `json:"name,omitempty"`
	URL  string                    `json:"url,omitempty"`
	Type clusterctlv1.ProviderType `json:"type,omitempty"`
}

// Init implementation of clusterctl reader interface
// This is dummy method that is must be present to implement Reader interface
func (f *AirshipReader) Init(config string) error {
	return nil
}

// Get implementation of clusterctl reader interface
func (f *AirshipReader) Get(key string) (string, error) {
	// TODO handle empty keys
	if val, ok := f.variables[key]; ok {
		return val, nil
	}
	return "", ErrValueForVariableNotSet{Variable: key}
}

// Set implementation of clusterctl reader interface
func (f *AirshipReader) Set(key, value string) {
	// TODO handle empty keys
	f.variables[key] = value
}

// UnmarshalKey implementation of clusterctl reader interface
func (f *AirshipReader) UnmarshalKey(key string, rawval interface{}) error {
	data, err := f.Get(key)
	if err != nil {
		return err
	}
	return yaml.Unmarshal([]byte(data), rawval)
}

// NewAirshipReader returns airship implementation of clusterctl reader interface
func NewAirshipReader(options *airshipv1.Clusterctl) (*AirshipReader, error) {
	variables := map[string]string{}
	providers := []configProvider{}
	for _, prov := range options.Providers {
		appendProvider := configProvider{
			Name: prov.Name,
			Type: clusterctlv1.ProviderType(prov.Type),
			URL:  prov.URL,
		}
		providers = append(providers, appendProvider)
	}
	b, err := yaml.Marshal(providers)
	if err != nil {
		return nil, err
	}
	// Add providers to config
	variables[config.ProvidersConfigKey] = string(b)
	// Add empty image configuration to config
	variables[imagesConfigKey] = ""
	return &AirshipReader{variables: variables}, nil
}
