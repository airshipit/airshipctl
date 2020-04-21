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

package redfish

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	redfishAPI "opendev.org/airship/go-redfish/api"
	redfishClient "opendev.org/airship/go-redfish/client"

	"opendev.org/airship/airshipctl/pkg/log"
)

const (
	// ClientType is used by other packages as the identifier of the Redfish client.
	ClientType          string = "redfish"
	mediaEjectDelay            = 30 * time.Second
	systemActionRetries        = 30
	systemRebootDelay          = 2 * time.Second
)

// Client holds details about a Redfish out-of-band system required for out-of-band management.
type Client struct {
	ephemeralNodeID string
	isoPath         string
	redfishURL      url.URL
	RedfishAPI      redfishAPI.RedfishAPI
	RedfishCFG      *redfishClient.Configuration
}

// EphemeralNodeID retrieves the ephemeral node ID.
func (c *Client) EphemeralNodeID() string {
	return c.ephemeralNodeID
}

// RebootSystem power cycles a host by sending a shutdown signal followed by a power on signal.
func (c *Client) RebootSystem(ctx context.Context, systemID string) error {
	waitForPowerState := func(desiredState redfishClient.PowerState) error {
		// Check if number of retries is defined in context
		totalRetries, ok := ctx.Value("numRetries").(int)
		if !ok {
			totalRetries = systemActionRetries
		}

		for retry := 0; retry <= totalRetries; retry++ {
			system, httpResp, err := c.RedfishAPI.GetSystem(ctx, systemID)
			if err = ScreenRedfishError(httpResp, err); err != nil {
				return err
			}
			if system.PowerState == desiredState {
				return nil
			}
			time.Sleep(systemRebootDelay)
		}
		return ErrOperationRetriesExceeded{}
	}

	resetReq := redfishClient.ResetRequestBody{}

	// Send PowerOff request
	resetReq.ResetType = redfishClient.RESETTYPE_FORCE_OFF
	_, httpResp, err := c.RedfishAPI.ResetSystem(ctx, systemID, resetReq)
	if err = ScreenRedfishError(httpResp, err); err != nil {
		return err
	}

	// Check that node is powered off
	if err = waitForPowerState(redfishClient.POWERSTATE_OFF); err != nil {
		return err
	}

	// Send PowerOn request
	resetReq.ResetType = redfishClient.RESETTYPE_ON
	_, httpResp, err = c.RedfishAPI.ResetSystem(ctx, systemID, resetReq)
	if err = ScreenRedfishError(httpResp, err); err != nil {
		return err
	}

	// Check that node is powered on and return
	return waitForPowerState(redfishClient.POWERSTATE_ON)
}

// SetEphemeralBootSourceByType sets the boot source of the ephemeral node to one that's compatible with the boot
// source type.
func (c *Client) SetEphemeralBootSourceByType(ctx context.Context) error {
	_, vMediaType, err := GetVirtualMediaID(ctx, c.RedfishAPI, c.ephemeralNodeID)
	if err != nil {
		return err
	}

	// Retrieve system information, containing available boot sources
	system, _, err := c.RedfishAPI.GetSystem(ctx, c.ephemeralNodeID)
	if err != nil {
		return ErrRedfishClient{Message: fmt.Sprintf("Get System[%s] failed with err: %v", c.ephemeralNodeID, err)}
	}

	allowableValues := system.Boot.BootSourceOverrideTargetRedfishAllowableValues
	for _, bootSource := range allowableValues {
		if strings.EqualFold(string(bootSource), vMediaType) {
			/* set boot source */
			systemReq := redfishClient.ComputerSystem{}
			systemReq.Boot.BootSourceOverrideTarget = bootSource
			_, httpResp, err := c.RedfishAPI.SetSystem(ctx, c.ephemeralNodeID, systemReq)
			return ScreenRedfishError(httpResp, err)
		}
	}

	return ErrRedfishClient{Message: fmt.Sprintf("failed to set system[%s] boot source", c.ephemeralNodeID)}
}

// SetVirtualMedia injects a virtual media device to an established virtual media ID. This assumes that isoPath is
// accessible to the redfish server and virtualMedia device is either of type CD or DVD.
func (c *Client) SetVirtualMedia(ctx context.Context, isoPath string) error {
	log.Debugf("Ephemeral Node System ID: '%s'", c.ephemeralNodeID)

	managerID, err := GetManagerID(ctx, c.RedfishAPI, c.ephemeralNodeID)
	if err != nil {
		return err
	}

	log.Debugf("Ephemeral node managerID: '%s'", managerID)

	vMediaID, _, err := GetVirtualMediaID(ctx, c.RedfishAPI, c.ephemeralNodeID)
	if err != nil {
		return err
	}

	// Eject virtual media if it is already inserted
	vMediaMgr, httpResp, err := c.RedfishAPI.GetManagerVirtualMedia(ctx, managerID, vMediaID)
	if err = ScreenRedfishError(httpResp, err); err != nil {
		return err
	}

	if *vMediaMgr.Inserted == true {
		var emptyBody map[string]interface{}
		_, httpResp, err = c.RedfishAPI.EjectVirtualMedia(ctx, managerID, vMediaID, emptyBody)
		if err = ScreenRedfishError(httpResp, err); err != nil {
			return err
		}

		time.Sleep(mediaEjectDelay)
	}

	vMediaReq := redfishClient.InsertMediaRequestBody{}
	vMediaReq.Image = isoPath
	vMediaReq.Inserted = true
	_, httpResp, err = c.RedfishAPI.InsertVirtualMedia(ctx, managerID, vMediaID, vMediaReq)

	return ScreenRedfishError(httpResp, err)
}

// SystemPowerOff shuts down a host.
func (c *Client) SystemPowerOff(ctx context.Context, systemID string) error {
	resetReq := redfishClient.ResetRequestBody{}
	resetReq.ResetType = redfishClient.RESETTYPE_FORCE_OFF

	_, httpResp, err := c.RedfishAPI.ResetSystem(ctx, systemID, resetReq)

	return ScreenRedfishError(httpResp, err)
}

// SystemPowerStatus retrieves the power status of a host as a human-readable string.
func (c *Client) SystemPowerStatus(ctx context.Context, systemID string) (string, error) {
	computerSystem, httpResp, err := c.RedfishAPI.GetSystem(ctx, systemID)
	if err = ScreenRedfishError(httpResp, err); err != nil {
		return "", err
	}

	return string(computerSystem.PowerState), nil
}

// NewClient returns a client with the capability to make Redfish requests.
func NewClient(ephemeralNodeID string,
	isoPath string,
	redfishURL string,
	insecure bool,
	useProxy bool,
	username string,
	password string) (context.Context, *Client, error) {
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
		return ctx, nil, ErrRedfishMissingConfig{What: "Redfish URL"}
	}

	parsedURL, err := url.Parse(redfishURL)
	if err != nil {
		return ctx, nil, err
	}

	cfg := &redfishClient.Configuration{
		BasePath:      redfishURL,
		DefaultHeader: make(map[string]string),
		UserAgent:     headerUserAgent,
	}

	// see https://github.com/golang/go/issues/26013
	// We clone the default transport to ensure when we customize the transport
	// that we are providing it sane timeouts and other defaults that we would
	// normally get when not overriding the transport
	defaultTransportCopy := http.DefaultTransport.(*http.Transport) //nolint:errcheck
	transport := defaultTransportCopy.Clone()

	if insecure {
		transport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true, //nolint:gosec
		}
	}

	if !useProxy {
		transport.Proxy = nil
	}

	cfg.HTTPClient = &http.Client{
		Transport: transport,
	}

	c := &Client{
		ephemeralNodeID: ephemeralNodeID,
		isoPath:         isoPath,
		redfishURL:      *parsedURL,
		RedfishAPI:      redfishClient.NewAPIClient(cfg).DefaultApi,
		RedfishCFG:      cfg,
	}

	return ctx, c, nil
}
