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
	"opendev.org/airship/airshipctl/pkg/remote/ifc"
	"opendev.org/airship/airshipctl/pkg/remote/power"
)

const (
	// ClientType is used by other packages as the identifier of the Redfish client.
	ClientType string = "redfish"
)

// Client holds details about a Redfish out-of-band system required for out-of-band management.
type Client struct {
	nodeID              string
	username            string
	password            string
	redfishURL          string
	RedfishAPI          redfishAPI.RedfishAPI
	RedfishCFG          *redfishClient.Configuration
	systemActionRetries int
	systemRebootDelay   int

	// Sleep is meant to be mocked out for tests
	Sleep func(d time.Duration)
}

// NodeID retrieves the ephemeral node ID.
func (c *Client) NodeID() string {
	return c.nodeID
}

// SystemActionRetries returns number of attempts to reach host during reboot process and ejecting virtual media
func (c *Client) SystemActionRetries() int {
	return c.systemActionRetries
}

// SystemRebootDelay returns number of seconds to wait after reboot if host isn't available
func (c *Client) SystemRebootDelay() int {
	return c.systemRebootDelay
}

// EjectVirtualMedia ejects a virtual media device attached to a host.
func (c *Client) EjectVirtualMedia(ctx context.Context) error {
	ctx = SetAuth(ctx, c.username, c.password)
	waitForEjectMedia := func(managerID string, mediaID string) error {
		for retry := 0; retry < c.systemActionRetries; retry++ {
			getMediaReq := c.RedfishAPI.GetManagerVirtualMedia(ctx, managerID, mediaID)
			vMediaMgr, httpResp, err := c.RedfishAPI.GetManagerVirtualMediaExecute(getMediaReq)
			if err = ScreenRedfishError(httpResp, err); err != nil {
				return err
			}

			if !vMediaMgr.GetInserted() {
				log.Debugf("Successfully ejected virtual media.")
				return nil
			}
		}

		return ErrOperationRetriesExceeded{What: fmt.Sprintf("eject media %s", mediaID), Retries: c.systemActionRetries}
	}

	managerID, err := getManagerID(ctx, c.RedfishAPI, c.nodeID)
	if err != nil {
		return err
	}

	listMediaReq := c.RedfishAPI.ListManagerVirtualMedia(ctx, managerID)
	mediaCollection, httpResp, err := c.RedfishAPI.ListManagerVirtualMediaExecute(listMediaReq)
	if err = ScreenRedfishError(httpResp, err); err != nil {
		return err
	}

	// Walk all virtual media devices and eject if inserted
	for _, mediaURI := range mediaCollection.Members {
		mediaID := GetResourceIDFromURL(*mediaURI.OdataId)

		getMediaReq := c.RedfishAPI.GetManagerVirtualMedia(ctx, managerID, mediaID)
		vMediaMgr, httpResp, err := c.RedfishAPI.GetManagerVirtualMediaExecute(getMediaReq)
		if err = ScreenRedfishError(httpResp, err); err != nil {
			return err
		}

		if vMediaMgr.GetInserted() {
			log.Debugf("'%s' has virtual media inserted. Attempting to eject.", vMediaMgr.Name)

			var emptyBody map[string]interface{}
			ejectMediaReq := c.RedfishAPI.EjectVirtualMedia(ctx, managerID, mediaID).Body(emptyBody)
			_, httpResp, err = c.RedfishAPI.EjectVirtualMediaExecute(ejectMediaReq)
			if err = ScreenRedfishError(httpResp, err); err != nil {
				return err
			}

			if err = waitForEjectMedia(managerID, mediaID); err != nil {
				return err
			}
		}
	}

	return nil
}

// RebootSystem power cycles a host by sending a shutdown signal followed by a power on signal.
func (c *Client) RebootSystem(ctx context.Context) error {
	log.Debugf("Rebooting node '%s': powering off.", c.nodeID)
	ctx = SetAuth(ctx, c.username, c.password)
	resetReq := redfishClient.NewResetRequestBodyWithDefaults()

	// Send PowerOff request
	resetReq.SetResetType(redfishClient.RESETTYPE_FORCE_OFF)
	resetSystemReq := c.RedfishAPI.ResetSystem(ctx, c.nodeID).ResetRequestBody(*resetReq)
	_, httpResp, err := c.RedfishAPI.ResetSystemExecute(resetSystemReq)
	if err = ScreenRedfishError(httpResp, err); err != nil {
		log.Debugf("Failed to reboot node '%s': shutdown failure.", c.nodeID)
		return err
	}

	// Check that node is powered off
	if err = c.waitForPowerState(ctx, redfishClient.POWERSTATE_OFF); err != nil {
		return err
	}

	log.Debugf("Rebooting node '%s': powering on.", c.nodeID)

	// Send PowerOn request
	resetReq.ResetType = redfishClient.RESETTYPE_ON.Ptr()
	resetSystemReq = c.RedfishAPI.ResetSystem(ctx, c.nodeID).ResetRequestBody(*resetReq)
	_, httpResp, err = c.RedfishAPI.ResetSystemExecute(resetSystemReq)
	if err = ScreenRedfishError(httpResp, err); err != nil {
		log.Debugf("Failed to reboot node '%s': startup failure.", c.nodeID)
		return err
	}

	// Check that node is powered on and return
	return c.waitForPowerState(ctx, redfishClient.POWERSTATE_ON)
}

// SetBootSourceByType sets the boot source of the ephemeral node to one that's compatible with the boot
// source type.
func (c *Client) SetBootSourceByType(ctx context.Context) error {
	ctx = SetAuth(ctx, c.username, c.password)
	_, vMediaType, err := GetVirtualMediaID(ctx, c.RedfishAPI, c.nodeID)
	if err != nil {
		return err
	}

	log.Debugf("Setting boot device to '%s'.", vMediaType)

	// Retrieve system information, containing available boot sources
	systemReq := c.RedfishAPI.GetSystem(ctx, c.nodeID)
	system, _, err := c.RedfishAPI.GetSystemExecute(systemReq)
	if err != nil {
		return ErrRedfishClient{Message: fmt.Sprintf("Get System[%s] failed with err: %v", c.nodeID, err)}
	}

	boot := system.GetBoot()
	allowableValues := boot.GetBootSourceOverrideTargetRedfishAllowableValues()
	for i := range allowableValues {
		if strings.EqualFold(string(allowableValues[i]), vMediaType) {
			/* set boot source */
			systemReq := redfishClient.ComputerSystem{}
			systemReq.SetBoot(redfishClient.Boot{
				BootSourceOverrideTarget: &allowableValues[i],
			})
			setSystemReq := c.RedfishAPI.SetSystem(ctx, c.nodeID).ComputerSystem(systemReq)
			_, httpResp, err := c.RedfishAPI.SetSystemExecute(setSystemReq)
			if err = ScreenRedfishError(httpResp, err); err != nil {
				return err
			}

			log.Debug("Successfully set boot device.")
			return nil
		}
	}

	return ErrRedfishClient{Message: fmt.Sprintf("failed to set system[%s] boot source", c.nodeID)}
}

// SetVirtualMedia injects a virtual media device to an established virtual media ID. This assumes that isoPath is
// accessible to the redfish server and virtualMedia device is either of type CD or DVD.
func (c *Client) SetVirtualMedia(ctx context.Context, isoPath string) error {
	ctx = SetAuth(ctx, c.username, c.password)
	log.Debugf("Inserting virtual media '%s'.", isoPath)
	// Eject all previously-inserted media
	if err := c.EjectVirtualMedia(ctx); err != nil {
		return err
	}

	// Retrieve the ID of a compatible media type
	vMediaID, _, err := GetVirtualMediaID(ctx, c.RedfishAPI, c.nodeID)
	if err != nil {
		return err
	}

	managerID, err := getManagerID(ctx, c.RedfishAPI, c.nodeID)
	if err != nil {
		return err
	}

	// Insert media
	vMediaReq := redfishClient.InsertMediaRequestBody{}
	vMediaReq.Image = isoPath
	vMediaReq.SetInserted(true)

	insertMediaReq := c.RedfishAPI.InsertVirtualMedia(ctx, managerID, vMediaID).InsertMediaRequestBody(vMediaReq)
	_, httpResp, err := c.RedfishAPI.InsertVirtualMediaExecute(insertMediaReq)

	if err = ScreenRedfishError(httpResp, err); err != nil {
		return err
	}

	log.Debug("Successfully set virtual media.")
	return nil
}

// SystemPowerOff shuts down a host.
func (c *Client) SystemPowerOff(ctx context.Context) error {
	ctx = SetAuth(ctx, c.username, c.password)
	resetReq := redfishClient.ResetRequestBody{}
	resetReq.ResetType = redfishClient.RESETTYPE_FORCE_OFF.Ptr()

	resetSystemReq := c.RedfishAPI.ResetSystem(ctx, c.nodeID).ResetRequestBody(resetReq)
	_, httpResp, err := c.RedfishAPI.ResetSystemExecute(resetSystemReq)

	// if already powered off, a status conflict message is returned but we are good to go
	if httpResp != nil && httpResp.StatusCode == http.StatusConflict {
		return nil
	}
	//screen for any errors
	if err = ScreenRedfishError(httpResp, err); err != nil {
		return err
	}

	return c.waitForPowerState(ctx, redfishClient.POWERSTATE_OFF)
}

// SystemPowerOn powers on a host.
func (c *Client) SystemPowerOn(ctx context.Context) error {
	ctx = SetAuth(ctx, c.username, c.password)
	resetReq := redfishClient.NewResetRequestBodyWithDefaults()
	resetReq.SetResetType(redfishClient.RESETTYPE_ON)

	resetSystemReq := c.RedfishAPI.ResetSystem(ctx, c.nodeID).ResetRequestBody(*resetReq)

	log.Printf("reset system request: %+v", resetSystemReq)

	_, httpResp, err := c.RedfishAPI.ResetSystemExecute(resetSystemReq)

	if err = ScreenRedfishError(httpResp, err); err != nil {
		return err
	}

	return c.waitForPowerState(ctx, redfishClient.POWERSTATE_ON)
}

// SystemPowerStatus retrieves the power status of a host as a human-readable string.
func (c *Client) SystemPowerStatus(ctx context.Context) (power.Status, error) {
	ctx = SetAuth(ctx, c.username, c.password)
	systemReq := c.RedfishAPI.GetSystem(ctx, c.nodeID)
	computerSystem, httpResp, err := c.RedfishAPI.GetSystemExecute(systemReq)

	if screenErr := ScreenRedfishError(httpResp, err); screenErr != nil {
		return power.StatusUnknown, screenErr
	}

	if computerSystem.PowerState == nil {
		log.Printf("csystem power: %+v", computerSystem)
		return power.StatusUnknown, nil
	}

	switch *computerSystem.PowerState {
	case redfishClient.POWERSTATE_ON:
		return power.StatusOn, nil
	case redfishClient.POWERSTATE_OFF:
		return power.StatusOff, nil
	case redfishClient.POWERSTATE_POWERING_ON:
		return power.StatusPoweringOn, nil
	case redfishClient.POWERSTATE_POWERING_OFF:
		return power.StatusPoweringOff, nil
	default:
		log.Printf("csystem power: %+v", computerSystem.PowerState)
		return power.StatusUnknown, nil
	}
}

// RemoteDirect implements remote direct interface
func (c *Client) RemoteDirect(ctx context.Context, isoURL string) error {
	return RemoteDirect(ctx, isoURL, c.redfishURL, c)
}

// RemoteDirect allows to perform remotedirect
func RemoteDirect(ctx context.Context, isoURL, redfishURL string, c ifc.Client) error {
	log.Debugf("Bootstrapping ephemeral host with ID '%s' and BMC Address '%s'.", c.NodeID(),
		redfishURL)

	powerStatus, err := c.SystemPowerStatus(ctx)
	if err != nil {
		return err
	}

	// Power off node if it is on
	if powerStatus != power.StatusOff {
		log.Debugf("Ephemeral node has power status '%s'. Attempting to power off.", powerStatus.String())
		if err = c.SystemPowerOff(ctx); err != nil {
			return err
		}
	}

	// Perform remote direct operations
	if isoURL == "" {
		return ErrRedfishMissingConfig{What: "isoURL"}
	}

	err = c.SetVirtualMedia(ctx, isoURL)
	if err != nil {
		return err
	}

	err = c.SetBootSourceByType(ctx)
	if err != nil {
		return err
	}

	err = c.SystemPowerOn(ctx)
	if err != nil {
		return err
	}

	log.Printf("Successfully bootstrapped ephemeral host '%s'.", c.NodeID())

	return nil
}

// NewClient returns a client with the capability to make Redfish requests.
func NewClient(redfishURL string,
	insecure bool,
	useProxy bool,
	username string,
	password string,
	systemActionRetries int,
	systemRebootDelay int) (*Client, error) {
	if redfishURL == "" {
		return nil, ErrRedfishMissingConfig{What: "Redfish URL"}
	}

	basePath, err := getBasePath(redfishURL)
	if err != nil {
		return nil, err
	}

	cfg := &redfishClient.Configuration{
		Servers: redfishClient.ServerConfigurations{
			{
				URL: basePath,
			},
		},
		DefaultHeader: make(map[string]string),
		UserAgent:     headerUserAgent,
	}

	// see https://github.com/golang/go/issues/26013
	// We clone the default transport to ensure when we customize the transport
	// that we are providing it sane timeouts and other defaults that we would
	// normally get when not overriding the transport
	defaultTransportCopy := http.DefaultTransport.(*http.Transport) //nolint:errcheck
	transport := defaultTransportCopy.Clone()
	transport.DisableKeepAlives = true
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
		return nil, ErrRedfishMissingConfig{What: "management URL system ID"}
	}

	c := &Client{
		nodeID:              systemID,
		RedfishAPI:          redfishClient.NewAPIClient(cfg).DefaultApi,
		RedfishCFG:          cfg,
		systemActionRetries: systemActionRetries,
		systemRebootDelay:   systemRebootDelay,
		password:            password,
		username:            username,
		redfishURL:          redfishURL,

		Sleep: func(d time.Duration) {
			time.Sleep(d)
		},
	}

	return c, nil
}

// ClientFactory is a constructor for redfish ifc.Client implementation
var ClientFactory ifc.ClientFactory = func(redfishURL string,
	insecure bool,
	useProxy bool,
	username string,
	password string,
	systemActionRetries int,
	systemRebootDelay int) (ifc.Client, error) {
	return NewClient(redfishURL, insecure, useProxy,
		username, password, systemActionRetries, systemRebootDelay)
}
