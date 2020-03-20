// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package redfishutils

import (
	"context"
	"net/url"

	"github.com/stretchr/testify/mock"
	redfishClient "opendev.org/airship/go-redfish/client"

	"opendev.org/airship/airshipctl/pkg/remote/redfish"
)

// MockClient is a fake Redfish client for unit testing.
type MockClient struct {
	mock.Mock
	ephemeralNodeID string
	isoPath         string
	redfishURL      url.URL
}

// EphemeralNodeID provides a stubbed method that can be mocked to test functions that use the Redfish client without
// making any Redfish API calls or requiring the appropriate Redfish client settings.
//
//     Example usage:
//         client := redfishutils.NewClient()
//         client.On("GetEphemeralNodeID").Return(<return values>)
//
//         err := client.GetEphemeralNodeID(<args>)
func (m *MockClient) EphemeralNodeID() string {
	args := m.Called()
	return args.String(0)
}

// RebootSystem provides a stubbed method that can be mocked to test functions that use the Redfish client without
// making any Redfish API calls or requiring the appropriate Redfish client settings.
//
//     Example usage:
//         client := redfishutils.NewClient()
//         client.On("RebootSystem").Return(<return values>)
//
//         err := client.RebootSystem(<args>)
func (m *MockClient) RebootSystem(ctx context.Context, systemID string) error {
	args := m.Called(ctx, systemID)
	return args.Error(0)
}

// SetEphemeralBootSourceByType provides a stubbed method that can be mocked to test functions that use the
// Redfish client without making any Redfish API calls or requiring the appropriate Redfish client settings.
//
//     Example usage:
//         client := redfishutils.NewClient()
//         client.On("SetEphemeralBootSourceByType").Return(<return values>)
//
//         err := client.setEphemeralBootSourceByType(<args>)
func (m *MockClient) SetEphemeralBootSourceByType(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// SetVirtualMedia provides a stubbed method that can be mocked to test functions that use the
// Redfish client without making any Redfish API calls or requiring the appropriate Redfish client settings.
//
//     Example usage:
//         client := redfishutils.NewClient()
//         client.On("SetVirtualMedia").Return(<return values>)
//
//         err := client.SetVirtualMedia(<args>)
func (m *MockClient) SetVirtualMedia(ctx context.Context, isoPath string) error {
	args := m.Called(ctx, isoPath)
	return args.Error(0)
}

// NewClient returns a mocked Redfish client in order to test functions that use the Redfish client without making any
// Redfish API calls.
func NewClient(ephemeralNodeID string, isoPath string, redfishURL string, insecure bool,
	proxy bool, username string, password string) (context.Context, *MockClient, error) {
	var ctx context.Context
	if username != "" && password != "" {
		ctx = context.WithValue(
			context.Background(),
			redfishClient.ContextBasicAuth,
			redfishClient.BasicAuth{UserName: username, Password: password},
		)
	} else {
		ctx = context.Background()
	}

	if redfishURL == "" {
		return ctx, nil, redfish.ErrRedfishMissingConfig{What: "Redfish URL"}
	}

	parsedURL, err := url.Parse(redfishURL)
	if err != nil {
		return ctx, nil, err
	}

	m := &MockClient{
		ephemeralNodeID: ephemeralNodeID,
		isoPath:         isoPath,
		redfishURL:      *parsedURL,
	}

	return ctx, m, nil
}
