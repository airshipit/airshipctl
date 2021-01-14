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
	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/log"
	"opendev.org/airship/airshipctl/pkg/phase"
	"opendev.org/airship/airshipctl/pkg/phase/ifc"
	remoteifc "opendev.org/airship/airshipctl/pkg/remote/ifc"
	"opendev.org/airship/airshipctl/pkg/remote/redfish"
	redfishdell "opendev.org/airship/airshipctl/pkg/remote/redfish/vendors/dell"
)

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
	remoteifc.Client
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

		var matchingHosts []baremetalHost
		for _, doc := range docs {
			host, err := newBaremetalHost(mgmtCfg, doc, docBundle)
			if err != nil {
				return err
			}

			matchingHosts = append(matchingHosts, host)
		}

		a.Hosts = reconcileHosts(a.Hosts, matchingHosts...)

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

		a.Hosts = reconcileHosts(a.Hosts, host)

		return nil
	}
}

// All adds the host to a manager whose documents match the BareMetalHostKind.
func All() HostSelector {
	return func(a *Manager, mgmtCfg config.ManagementConfiguration, docBundle document.Bundle) error {
		selector := document.NewSelector().ByKind(document.BareMetalHostKind)
		docs, err := docBundle.Select(selector)
		if err != nil {
			return err
		}

		if len(docs) == 0 {
			return document.ErrDocNotFound{Selector: selector}
		}

		var matchingHosts []baremetalHost
		for _, doc := range docs {
			host, err := newBaremetalHost(mgmtCfg, doc, docBundle)
			if err != nil {
				return err
			}

			matchingHosts = append(matchingHosts, host)
		}

		a.Hosts = reconcileHosts(a.Hosts, matchingHosts...)

		return nil
	}
}

// NewManager provides a manager that exposes the capability to perform remote direct functionality and other
// out-of-band management on multiple hosts.
func NewManager(cfg *config.Config, phaseName string, hosts ...HostSelector) (*Manager, error) {
	managementCfg, err := cfg.CurrentContextManagementConfig()
	if err != nil {
		return nil, err
	}

	if err = managementCfg.Validate(); err != nil {
		return nil, err
	}

	helper, err := phase.NewHelper(cfg)
	if err != nil {
		return nil, err
	}

	phaseClient := phase.NewClient(helper)
	phase, err := phaseClient.PhaseByID(ifc.ID{Name: phaseName})
	if err != nil {
		return nil, err
	}

	docRoot, err := phase.DocumentRoot()
	if err != nil {
		return nil, err
	}

	docBundle, err := document.NewBundleByPath(docRoot)
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

	if len(manager.Hosts) == 0 {
		return manager, ErrNoHostsFound{}
	}

	return manager, nil
}

// newBaremetalHost creates a representation of a baremetal host that is configured to perform management actions by
// invoking its client methods (provided by the remote.Client interface).
func newBaremetalHost(mgmtCfg config.ManagementConfiguration,
	hostDoc document.Document,
	docBundle document.Bundle) (baremetalHost, error) {
	address, err := document.GetBMHBMCAddress(hostDoc)
	if err != nil {
		return baremetalHost{}, err
	}

	username, password, err := document.GetBMHBMCCredentials(hostDoc, docBundle)
	if err != nil {
		return baremetalHost{}, err
	}

	var clientFactory remoteifc.ClientFactory
	// Select the client that corresponds to the management type specified in the airshipctl config.
	switch mgmtCfg.Type {
	case redfish.ClientType:
		log.Debug("Remote type: Redfish")
		clientFactory = redfish.ClientFactory
	case redfishdell.ClientType:
		clientFactory = redfishdell.ClientFactory
	default:
		return baremetalHost{}, ErrUnknownManagementType{Type: mgmtCfg.Type}
	}

	client, err := clientFactory(
		address,
		mgmtCfg.Insecure,
		mgmtCfg.UseProxy,
		username,
		password, mgmtCfg.SystemActionRetries,
		mgmtCfg.SystemRebootDelay)

	if err != nil {
		return baremetalHost{}, err
	}

	return baremetalHost{client, address, hostDoc.GetName(), username, password}, nil
}

// reconcileHosts produces the intersection of two baremetal host arrays.
func reconcileHosts(existingHosts []baremetalHost, newHosts ...baremetalHost) []baremetalHost {
	if len(existingHosts) == 0 {
		return newHosts
	}

	// Create a map of host document names for efficient filtering
	hostMap := make(map[string]bool)
	for _, host := range existingHosts {
		hostMap[host.HostName] = true
	}

	var reconciledHosts []baremetalHost
	for _, host := range newHosts {
		if _, exists := hostMap[host.HostName]; exists {
			reconciledHosts = append(reconciledHosts, host)
		}
	}

	return reconciledHosts
}
