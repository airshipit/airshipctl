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
	"strings"
	"time"

	redfishAPI "opendev.org/airship/go-redfish/api"
	redfishClient "opendev.org/airship/go-redfish/client"

	"opendev.org/airship/airshipctl/pkg/log"
)

const (
	// ClientType is used by other packages as the identifier of the Redfish client.
	ClientType          string = "redfish"
	systemActionRetries        = 30
	systemRebootDelay          = 30 * time.Second
)

// Client holds details about a Redfish out-of-band system required for out-of-band management.
type Client struct {
	nodeID     string
	RedfishAPI redfishAPI.RedfishAPI
	RedfishCFG *redfishClient.Configuration
}

// NodeID retrieves the ephemeral node ID.
func (c *Client) NodeID() string {
	return c.nodeID
}

// RebootSystem power cycles a host by sending a shutdown signal followed by a power on signal.
func (c *Client) RebootSystem(ctx context.Context) error {
	waitForPowerState := func(desiredState redfishClient.PowerState) error {
		// Check if number of retries is defined in context
		totalRetries, ok := ctx.Value("numRetries").(int)
		if !ok {
			totalRetries = systemActionRetries
		}

		for retry := 0; retry <= totalRetries; retry++ {
			system, httpResp, err := c.RedfishAPI.GetSystem(ctx, c.nodeID)
			if err = ScreenRedfishError(httpResp, err); err != nil {
				return err
			}
			if system.PowerState == desiredState {
				return nil
			}
			time.Sleep(systemRebootDelay)
		}
		return ErrOperationRetriesExceeded{
			What:    fmt.Sprintf("reboot system %s", c.nodeID),
			Retries: totalRetries,
		}
	}

	resetReq := redfishClient.ResetRequestBody{}

	// Send PowerOff request
	resetReq.ResetType = redfishClient.RESETTYPE_FORCE_OFF
	_, httpResp, err := c.RedfishAPI.ResetSystem(ctx, c.nodeID, resetReq)
	if err = ScreenRedfishError(httpResp, err); err != nil {
		return err
	}

	// Check that node is powered off
	if err = waitForPowerState(redfishClient.POWERSTATE_OFF); err != nil {
		return err
	}

	// Send PowerOn request
	resetReq.ResetType = redfishClient.RESETTYPE_ON
	_, httpResp, err = c.RedfishAPI.ResetSystem(ctx, c.nodeID, resetReq)
	if err = ScreenRedfishError(httpResp, err); err != nil {
		return err
	}

	// Check that node is powered on and return
	return waitForPowerState(redfishClient.POWERSTATE_ON)
}

// SetBootSourceByType sets the boot source of the ephemeral node to one that's compatible with the boot
// source type.
func (c *Client) SetBootSourceByType(ctx context.Context) error {
	_, vMediaType, err := GetVirtualMediaID(ctx, c.RedfishAPI, c.nodeID)
	if err != nil {
		return err
	}

	// Retrieve system information, containing available boot sources
	system, _, err := c.RedfishAPI.GetSystem(ctx, c.nodeID)
	if err != nil {
		return ErrRedfishClient{Message: fmt.Sprintf("Get System[%s] failed with err: %v", c.nodeID, err)}
	}

	allowableValues := system.Boot.BootSourceOverrideTargetRedfishAllowableValues
	for _, bootSource := range allowableValues {
		if strings.EqualFold(string(bootSource), vMediaType) {
			/* set boot source */
			systemReq := redfishClient.ComputerSystem{}
			systemReq.Boot.BootSourceOverrideTarget = bootSource
			_, httpResp, err := c.RedfishAPI.SetSystem(ctx, c.nodeID, systemReq)
			return ScreenRedfishError(httpResp, err)
		}
	}

	return ErrRedfishClient{Message: fmt.Sprintf("failed to set system[%s] boot source", c.nodeID)}
}

// SetVirtualMedia injects a virtual media device to an established virtual media ID. This assumes that isoPath is
// accessible to the redfish server and virtualMedia device is either of type CD or DVD.
func (c *Client) SetVirtualMedia(ctx context.Context, isoPath string) error {
	waitForEjectMedia := func(managerID string, vMediaID string) error {
		// Check if number of retries is defined in context
		totalRetries, ok := ctx.Value("numRetries").(int)
		if !ok {
			totalRetries = systemActionRetries
		}

		for retry := 0; retry < totalRetries; retry++ {
			vMediaMgr, httpResp, err := c.RedfishAPI.GetManagerVirtualMedia(ctx, managerID, vMediaID)
			if err = ScreenRedfishError(httpResp, err); err != nil {
				return err
			}

			if *vMediaMgr.Inserted == false {
				log.Debugf("Successfully ejected virtual media.")
				return nil
			}
		}

		return ErrOperationRetriesExceeded{What: fmt.Sprintf("eject media %s", vMediaID), Retries: totalRetries}
	}

	log.Debugf("Setting virtual media for node: '%s'", c.nodeID)

	managerID, err := getManagerID(ctx, c.RedfishAPI, c.nodeID)
	if err != nil {
		return err
	}

	log.Debugf("Ephemeral node managerID: '%s'", managerID)

	vMediaID, _, err := GetVirtualMediaID(ctx, c.RedfishAPI, c.nodeID)
	if err != nil {
		return err
	}

	// Eject virtual media if it is already inserted
	vMediaMgr, httpResp, err := c.RedfishAPI.GetManagerVirtualMedia(ctx, managerID, vMediaID)
	if err = ScreenRedfishError(httpResp, err); err != nil {
		return err
	}

	if *vMediaMgr.Inserted == true {
		log.Debugf("Manager %s media type %s inserted. Attempting to eject.", managerID, vMediaID)

		var emptyBody map[string]interface{}
		_, httpResp, err = c.RedfishAPI.EjectVirtualMedia(ctx, managerID, vMediaID, emptyBody)
		if err = ScreenRedfishError(httpResp, err); err != nil {
			return err
		}

		if err = waitForEjectMedia(managerID, vMediaID); err != nil {
			return err
		}
	}

	vMediaReq := redfishClient.InsertMediaRequestBody{}
	vMediaReq.Image = isoPath
	vMediaReq.Inserted = true
	_, httpResp, err = c.RedfishAPI.InsertVirtualMedia(ctx, managerID, vMediaID, vMediaReq)

	return ScreenRedfishError(httpResp, err)
}

// SystemPowerOff shuts down a host.
func (c *Client) SystemPowerOff(ctx context.Context) error {
	resetReq := redfishClient.ResetRequestBody{}
	resetReq.ResetType = redfishClient.RESETTYPE_FORCE_OFF

	_, httpResp, err := c.RedfishAPI.ResetSystem(ctx, c.nodeID, resetReq)

	return ScreenRedfishError(httpResp, err)
}

// SystemPowerStatus retrieves the power status of a host as a human-readable string.
func (c *Client) SystemPowerStatus(ctx context.Context) (string, error) {
	computerSystem, httpResp, err := c.RedfishAPI.GetSystem(ctx, c.nodeID)
	if err = ScreenRedfishError(httpResp, err); err != nil {
		return "", err
	}

	return string(computerSystem.PowerState), nil
}

// NewClient returns a client with the capability to make Redfish requests.
func NewClient(redfishURL string,
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

	basePath, err := getBasePath(redfishURL)
	if err != nil {
		return ctx, nil, err
	}

	cfg := &redfishClient.Configuration{
		BasePath:      basePath,
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

	// Retrieve system ID from end of Redfish URL
	systemID := GetResourceIDFromURL(redfishURL)
	if len(systemID) == 0 {
		return ctx, nil, ErrRedfishMissingConfig{What: "management URL system ID"}
	}

	c := &Client{
		nodeID:     systemID,
		RedfishAPI: redfishClient.NewAPIClient(cfg).DefaultApi,
		RedfishCFG: cfg,
	}

	return ctx, c, nil
}
