/*
  Licensed under the Apache License, Version 2.0 (the "License");
  you may not use this file except in compliance with the License.
  You may obtain a copy of the License at

      http://www.apache.org/licenses/LICENSE-2.0

  Unless required by applicable law or agreed to in writing, software
  distributed under the License is distributed on an "AS IS" BASIS,
  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
  See the License for the specific language governing permissions and
  limitations under the License.
*/

package config

import (
	"sigs.k8s.io/yaml"

	"opendev.org/airship/airshipctl/pkg/remote/redfish"
	redfishdell "opendev.org/airship/airshipctl/pkg/remote/redfish/vendors/dell"
)

const (
	insecureDefaultValue = false
	useProxyDefaultValue = false
)

// ManagementConfiguration defines configuration data for all remote systems within a context.
type ManagementConfiguration struct {
	// Insecure indicates whether the SSL certificate should be checked on remote management requests.
	Insecure bool `json:"insecure,omitempty"`

	// SystemActionRetries is the number of attempts to poll a host for a status.
	SystemActionRetries int `json:"systemActionRetries,omitempty"`

	// SystemRebootDelay is the number of seconds to wait between power actions (e.g. shutdown, startup).
	SystemRebootDelay int `json:"systemRebootDelay,omitempty"`

	// Type the type of out-of-band management that will be used for baremetal orchestration, e.g. redfish.
	Type string `json:"type"`

	// UseProxy indicates whether airshipctl should transmit remote management requests through a proxy server when
	// one is configured in an environment.
	UseProxy bool `json:"useproxy,omitempty"`
}

// SetType is a helper function that sets and validates the management type.
func (m *ManagementConfiguration) SetType(managementType string) error {
	prev := m.Type
	m.Type = managementType

	if err := m.Validate(); err != nil {
		m.Type = prev
		return err
	}

	return nil
}

// String converts a management configuration to a human-readable string.
func (m *ManagementConfiguration) String() string {
	yamlData, err := yaml.Marshal(&m)
	if err != nil {
		return ""
	}

	return string(yamlData)
}

// Validate validates that a management configuration is valid. Currently, this only checks the value of the management
// type as the other fields have appropriate zero values and may be omitted.
func (m *ManagementConfiguration) Validate() error {
	switch m.Type {
	case redfish.ClientType:
		m.Type = redfish.ClientType
	case redfishdell.ClientType:
		m.Type = redfishdell.ClientType
	default:
		return ErrUnknownManagementType{Type: m.Type}
	}

	return nil
}

// NewManagementConfiguration returns a management configuration with default values.
func NewManagementConfiguration() *ManagementConfiguration {
	return &ManagementConfiguration{
		Insecure:            insecureDefaultValue,
		SystemActionRetries: DefaultSystemActionRetries,
		SystemRebootDelay:   DefaultSystemRebootDelay,
		Type:                AirshipDefaultManagementType,
		UseProxy:            useProxyDefaultValue,
	}
}
