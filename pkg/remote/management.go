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

// Package remote manages baremetal hosts.
package remote

import (
	"context"

	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/environment"
	"opendev.org/airship/airshipctl/pkg/remote/redfish"
)

// Client is a set of functions that clients created for out-of-band power management and control should implement. The
// functions within client are used by power management commands and remote direct functionality.
type Client interface {
	RebootSystem(context.Context) error

	SystemPowerOff(context.Context) error

	// TODO(drewwalters96): Should this be a string forever? We may want to define our own custom type, as the
	// string format will be client dependent when we add new clients.
	SystemPowerStatus(context.Context) (string, error)

	NodeID() string

	SetBootSourceByType(context.Context) error

	// TODO(drewwalters96): This function is tightly coupled to Redfish. It should be combined with the
	// SetBootSource operation and removed from the client interface.
	SetVirtualMedia(context.Context, string) error
}

// Manager orchestrates a grouping of baremetal hosts. When a manager is created using its convenience function, the
// manager contains a list of hosts ready for out-of-band management. Iterate over the Hosts property to invoke actions
// on each host.
type Manager struct {
	Config config.ManagementConfiguration
	Hosts  []baremetalHost
}

// baremetalHost is an airshipctl representation of a baremetal host, defined by a baremetal host document, that embeds
// actions an out-of-band client can perform. Once instantiated, actions can be performed on a baremetal host.
type baremetalHost struct {
	Client
	Context    context.Context
	BMCAddress string
	HostName   string
	username   string
	password   string
}

// HostSelector populates baremetal hosts within a manager when supplied with selection criteria.
type HostSelector func(*Manager, config.ManagementConfiguration, document.Bundle) error

// ByLabel adds all hosts to a manager whose documents match a supplied label selector.
func ByLabel(label string) HostSelector {
	return func(a *Manager, mgmtCfg config.ManagementConfiguration, docBundle document.Bundle) error {
		selector := document.NewSelector().ByKind(document.BareMetalHostKind).ByLabel(label)
		docs, err := docBundle.Select(selector)
		if err != nil {
			return err
		}

		if len(docs) == 0 {
			return document.ErrDocNotFound{Selector: selector}
		}

		for _, doc := range docs {
			host, err := newBaremetalHost(mgmtCfg, doc, docBundle)
			if err != nil {
				return err
			}

			a.Hosts = append(a.Hosts, host)
		}

		return nil
	}
}

// ByName adds the host to a manager whose document meets the specified name.
func ByName(name string) HostSelector {
	return func(a *Manager, mgmtCfg config.ManagementConfiguration, docBundle document.Bundle) error {
		selector := document.NewSelector().ByKind(document.BareMetalHostKind).ByName(name)
		doc, err := docBundle.SelectOne(selector)
		if err != nil {
			return err
		}

		host, err := newBaremetalHost(mgmtCfg, doc, docBundle)
		if err != nil {
			return err
		}

		a.Hosts = append(a.Hosts, host)

		return nil
	}
}

// NewManager provides a manager that exposes the capability to perform remote direct functionality and other
// out-of-band management on multiple hosts.
func NewManager(settings *environment.AirshipCTLSettings, phase string, hosts ...HostSelector) (*Manager, error) {
	configContext, err := settings.Config.GetCurrentContext()
	if err != nil {
		return nil, err
	}

	managementCfg, err := settings.Config.CurrentContextManagementConfig()
	if err != nil {
		return nil, err
	}

	clusterType := configContext.ClusterType()
	entrypoint, err := settings.Config.CurrentContextEntryPoint(clusterType, phase)
	if err != nil {
		return nil, err
	}

	docBundle, err := document.NewBundleByPath(entrypoint)
	if err != nil {
		return nil, err
	}

	manager := &Manager{
		Config: *managementCfg,
		Hosts:  []baremetalHost{},
	}

	// Each function in hosts modifies the list of hosts for the new manager based on selection criteria provided
	// by CLI arguments and airshipctl settings.
	for _, addHost := range hosts {
		if err := addHost(manager, *managementCfg, docBundle); err != nil {
			return nil, err
		}
	}

	return manager, nil
}

// newBaremetalHost creates a representation of a baremetal host that is configured to perform management actions by
// invoking its client methods (provided by the remote.Client interface).
func newBaremetalHost(mgmtCfg config.ManagementConfiguration,
	hostDoc document.Document,
	docBundle document.Bundle) (baremetalHost, error) {
	var host baremetalHost

	address, err := document.GetBMHBMCAddress(hostDoc)
	if err != nil {
		return host, err
	}

	username, password, err := document.GetBMHBMCCredentials(hostDoc, docBundle)
	if err != nil {
		return host, err
	}

	// Select the client that corresponds to the management type specified in the airshipctl config.
	switch mgmtCfg.Type {
	case redfish.ClientType:
		ctx, client, err := redfish.NewClient(
			address,
			mgmtCfg.Insecure,
			mgmtCfg.UseProxy,
			username,
			password)

		if err != nil {
			return host, err
		}

		host = baremetalHost{client, ctx, address, hostDoc.GetName(), username, password}
	default:
		return host, ErrUnknownManagementType{Type: mgmtCfg.Type}
	}

	return host, nil
}
