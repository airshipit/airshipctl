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

package remote

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/environment"
	alog "opendev.org/airship/airshipctl/pkg/log"
	"opendev.org/airship/airshipctl/pkg/remote/redfish"
)

// Adapter bridges the gap between out-of-band clients. It can hold any type of OOB client, e.g. Redfish.
type Adapter struct {
	OOBClient    Client
	Context      context.Context
	remoteConfig *config.RemoteDirect
	remoteURL    string
	username     string
	password     string
}

// configureClient retrieves a client for remoteDirect requests based on the RemoteType in the Airship config file.
func (a *Adapter) configureClient(remoteURL string) error {
	switch a.remoteConfig.RemoteType {
	case redfish.ClientType:
		alog.Debug("Remote type redfish")

		rfURL, err := url.Parse(remoteURL)
		if err != nil {
			return err
		}

		baseURL := fmt.Sprintf("%s://%s", rfURL.Scheme, rfURL.Host)
		schemeSplit := strings.Split(rfURL.Scheme, redfish.URLSchemeSeparator)
		if len(schemeSplit) > 1 {
			baseURL = fmt.Sprintf("%s://%s", schemeSplit[len(schemeSplit)-1], rfURL.Host)
		}

		urlPath := strings.Split(rfURL.Path, "/")
		nodeID := urlPath[len(urlPath)-1]
		if nodeID == "" {
			return redfish.ErrRedfishMissingConfig{
				What: "redfish ephemeral node id empty",
			}
		}

		if a.remoteConfig.IsoURL == "" {
			return redfish.ErrRedfishMissingConfig{
				What: "redfish ephemeral node iso Path empty",
			}
		}

		a.Context, a.OOBClient, err = redfish.NewClient(
			nodeID,
			a.remoteConfig.IsoURL,
			baseURL,
			a.remoteConfig.Insecure,
			a.remoteConfig.UseProxy,
			a.username,
			a.password)
		if err != nil {
			alog.Debugf("redfish remotedirect client creation failed")
			return err
		}
	default:
		return NewRemoteDirectErrorf("invalid remote type")
	}

	return nil
}

// initializeAdapter retrieves the remote direct configuration defined in the Airship configuration file.
func (a *Adapter) initializeAdapter(settings *environment.AirshipCTLSettings) error {
	cfg := settings.Config
	bootstrapSettings, err := cfg.CurrentContextBootstrapInfo()
	if err != nil {
		return err
	}

	a.remoteConfig = bootstrapSettings.RemoteDirect
	if a.remoteConfig == nil {
		return config.ErrMissingConfig{What: "RemoteDirect options not defined in bootstrap config"}
	}

	bundlePath, err := cfg.CurrentContextEntryPoint(config.Ephemeral, "")
	if err != nil {
		return err
	}

	docBundle, err := document.NewBundleByPath(bundlePath)
	if err != nil {
		return err
	}

	selector := document.NewEphemeralBMHSelector()
	doc, err := docBundle.SelectOne(selector)
	if err != nil {
		return err
	}

	a.remoteURL, err = document.GetBMHBMCAddress(doc)
	if err != nil {
		return err
	}

	a.username, a.password, err = document.GetBMHBMCCredentials(doc, docBundle)
	if err != nil {
		return err
	}

	return nil
}

// DoRemoteDirect executes remote direct based on remote type.
func (a *Adapter) DoRemoteDirect() error {
	alog.Debugf("Using Remote Endpoint: %q", a.remoteURL)

	/* Load ISO in manager's virtual media */
	err := a.OOBClient.SetVirtualMedia(a.Context, a.remoteConfig.IsoURL)
	if err != nil {
		return err
	}

	alog.Debugf("Successfully loaded virtual media: %q", a.remoteConfig.IsoURL)

	/* Set system's bootsource to selected media */
	err = a.OOBClient.SetEphemeralBootSourceByType(a.Context)
	if err != nil {
		return err
	}

	/* Reboot system */
	err = a.OOBClient.RebootSystem(a.Context, a.OOBClient.EphemeralNodeID())
	if err != nil {
		return err
	}

	alog.Debug("Restarted ephemeral host")

	return nil
}

// NewAdapter provides an adapter that exposes the capability to perform remote direct functionality with any
// out-of-band client.
func NewAdapter(settings *environment.AirshipCTLSettings) (*Adapter, error) {
	a := &Adapter{}
	a.Context = context.Background()
	err := a.initializeAdapter(settings)
	if err != nil {
		return a, err
	}

	if err := a.configureClient(a.remoteURL); err != nil {
		return a, err
	}

	return a, nil
}
