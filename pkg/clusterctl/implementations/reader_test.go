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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	clusterctlv1 "sigs.k8s.io/cluster-api/cmd/clusterctl/api/v1alpha3"
	"sigs.k8s.io/cluster-api/cmd/clusterctl/client/config"

	airshipv1 "opendev.org/airship/airshipctl/pkg/api/v1alpha1"
)

func makeValidOptions() *airshipv1.Clusterctl {
	return &airshipv1.Clusterctl{
		Providers: []*airshipv1.Provider{
			{
				Name: "metal3",
				Type: "InfrastructureProvider",
				Versions: map[string]string{
					"v0.3.1": "manifests/function/capm3/v0.3.1",
				},
			},
			{
				Name: "kubeadm",
				Type: "BootstrapProvider",
				Versions: map[string]string{
					"v0.3.3": "manifests/function/cabpk/v0.3.3",
				},
			},
			{
				Name: "cluster-api",
				Type: "InfrastructureProvider",
				Versions: map[string]string{
					"v0.3.3": "manifests/function/capi/v0.3.3",
				},
			},
			{
				Name: "kubeadm",
				Type: "ControlPlaneProvider",
				Versions: map[string]string{
					"v0.3.3": "manifests/function/cacpk/v0.3.3",
				},
			},
		},
	}
}

func TestNewReader(t *testing.T) {
	tests := []struct {
		name    string
		options *airshipv1.Clusterctl
	}{
		{
			// make sure we get no panic here
			name:    "pass empty options",
			options: &airshipv1.Clusterctl{},
		},
		{
			name:    "pass airshipctl valid config",
			options: makeValidOptions(),
		},
	}
	for _, tt := range tests {
		options := tt.options
		t.Run(tt.name, func(t *testing.T) {
			reader, err := NewAirshipReader(options)
			require.NoError(t, err)
			assert.NotNil(t, reader)
		})
	}
}

func TestGet(t *testing.T) {
	tests := []struct {
		name           string
		options        *airshipv1.Clusterctl
		key            string
		expectedErr    error
		expectedResult string
	}{
		{
			// make sure we get no panic here
			name:        "pass empty options",
			options:     &airshipv1.Clusterctl{},
			key:         "FOO",
			expectedErr: ErrValueForVariableNotSet{Variable: "FOO"},
		},
		{
			name:        "pass airshipctl valid config",
			options:     makeValidOptions(),
			key:         "providers",
			expectedErr: nil,
			expectedResult: `- name: metal3
  type: InfrastructureProvider
- name: kubeadm
  type: BootstrapProvider
- name: cluster-api
  type: InfrastructureProvider
- name: kubeadm
  type: ControlPlaneProvider
`,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			reader, err := NewAirshipReader(tt.options)
			require.NoError(t, err)
			require.NotNil(t, reader)
			value, err := reader.Get(tt.key)
			assert.Equal(t, tt.expectedErr, err)
			assert.Equal(t, tt.expectedResult, value)
		})
	}
}

func TestSetGet(t *testing.T) {
	tests := []struct {
		name        string
		setKey      string
		setGetValue string
		expectedErr error
	}{
		{
			// should return empty string
			name:        "set simple key",
			setKey:      "FOO",
			expectedErr: nil,
			setGetValue: "",
		},
		{
			name:        "set providers",
			setKey:      "providers",
			expectedErr: nil,
			setGetValue: `- name: metal3
  type: InfrastructureProvider
- name: kubeadm
  type: BootstrapProvider
- name: cluster-api
  type: InfrastructureProvider
- name: kubeadm
  type: ControlPlaneProvider
`,
		},
		{
			// set empty
			name:        "empty key",
			setKey:      "",
			setGetValue: "some key",
			expectedErr: nil,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			reader, err := NewAirshipReader(&airshipv1.Clusterctl{})
			require.NoError(t, err)
			require.NotNil(t, reader)
			reader.Set(tt.setKey, tt.setGetValue)
			result, err := reader.Get(tt.setKey)
			require.Equal(t, tt.expectedErr, err)
			assert.Equal(t, tt.setGetValue, result)
		})
	}
}

// Test verifies that options provider returns
func TestUnmarshalProviders(t *testing.T) {
	options := &airshipv1.Clusterctl{
		Providers: []*airshipv1.Provider{
			{
				Name: config.Metal3ProviderName,
				Type: string(clusterctlv1.InfrastructureProviderType),
			},
			{
				Name: config.KubeadmBootstrapProviderName,
				Type: string(clusterctlv1.BootstrapProviderType),
			},
			{
				Name: config.ClusterAPIProviderName,
				Type: string(clusterctlv1.CoreProviderType),
			},
			{
				Name: config.KubeadmControlPlaneProviderName,
				Type: string(clusterctlv1.ControlPlaneProviderType),
			},
		},
	}
	providers := []configProvider{}
	reader, err := NewAirshipReader(options)
	require.NoError(t, err)
	require.NotNil(t, reader)
	// check if we can unmarshal provider key into correct struct
	err = reader.UnmarshalKey(config.ProvidersConfigKey, &providers)
	require.NoError(t, err)
	assert.Len(t, providers, 4)
	for _, actualProvider := range providers {
		assert.NotNil(t, options.Provider(actualProvider.Name, actualProvider.Type))
	}
}

func TestUnmarshal(t *testing.T) {
	tests := []struct {
		name      string
		expectErr bool
		variables map[string]string
		getKey    string
		unmarshal interface{}
	}{
		{
			name:      "unmarshal into nil",
			getKey:    "Foo",
			expectErr: true,
		},
		{
			name:      "value doesn't exist",
			getKey:    "Foo",
			variables: map[string]string{},
			unmarshal: []configProvider{},
			expectErr: true,
		},
		{
			name:      "value doesn't exist",
			getKey:    "foo",
			expectErr: false,
			variables: map[string]string{
				"foo": "foo: bar",
			},
			unmarshal: &struct {
				Foo string `json:"foo,omitempty"`
			}{},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			reader, err := NewAirshipReader(&airshipv1.Clusterctl{})
			require.NoError(t, err)
			require.NotNil(t, reader)
			reader.variables = tt.variables
			if tt.expectErr {
				assert.Error(t, reader.UnmarshalKey(tt.getKey, tt.unmarshal))
			} else {
				assert.NoError(t, reader.UnmarshalKey(tt.getKey, tt.unmarshal))
			}
		})
	}
}

// This test is simply for test coverage of the Reader interface
func TestInit(t *testing.T) {
	reader, err := NewAirshipReader(&airshipv1.Clusterctl{})
	require.NoError(t, err)
	require.NotNil(t, reader)
	assert.NoError(t, reader.Init("anything"))
}
