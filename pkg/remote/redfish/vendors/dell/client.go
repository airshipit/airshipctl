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

// Package dell wraps the standard Redfish client in order to provide additional functionality required to perform
// actions on iDRAC servers.
package dell

import (
	"context"

	redfishAPI "opendev.org/airship/go-redfish/api"
	redfishClient "opendev.org/airship/go-redfish/client"

	"opendev.org/airship/airshipctl/pkg/remote/redfish"
)

const (
	// ClientType is used by other packages as the identifier of the Redfish client.
	ClientType string = "redfish-dell"
)

// Client is a wrapper around the standard airshipctl Redfish client. This allows vendor specific Redfish clients to
// override methods without duplicating the entire client.
type Client struct {
	redfish.Client
	RedfishAPI redfishAPI.RedfishAPI
	RedfishCFG *redfishClient.Configuration
}

// NewClient returns a client with the capability to make Redfish requests.
func NewClient(ephemeralNodeID string,
	isoPath string,
	redfishURL string,
	insecure bool,
	useProxy bool,
	username string,
	password string) (context.Context, *Client, error) {
	ctx, genericClient, err := redfish.NewClient(
		ephemeralNodeID, isoPath, redfishURL, insecure, useProxy, username, password)
	if err != nil {
		return ctx, nil, err
	}

	c := &Client{*genericClient, genericClient.RedfishAPI, genericClient.RedfishCFG}

	return ctx, c, nil
}
