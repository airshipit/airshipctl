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

package baremetal

import (
	"context"

	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/errors"
	"opendev.org/airship/airshipctl/pkg/inventory/ifc"
	"opendev.org/airship/airshipctl/pkg/log"
	remoteifc "opendev.org/airship/airshipctl/pkg/remote/ifc"
	"opendev.org/airship/airshipctl/pkg/remote/redfish"
	redfishdell "opendev.org/airship/airshipctl/pkg/remote/redfish/vendors/dell"
)

// Inventory implements baremetal invenotry interface
type Inventory struct {
	mgmtCfg         config.ManagementConfiguration
	inventoryBundle document.Bundle
}

var _ ifc.BaremetalInventory = Inventory{}

// NewInventory returns inventory implementation based on BaremetalHost objects
func NewInventory(
	mgmtCfg config.ManagementConfiguration,
	inventoryBundle document.Bundle) ifc.BaremetalInventory {
	return Inventory{
		mgmtCfg:         mgmtCfg,
		inventoryBundle: inventoryBundle,
	}
}

// Select selects hosts based on given selector
func (i Inventory) Select(selector ifc.BaremetalHostSelector) ([]remoteifc.Client, error) {
	log.Debugf("Using selector %v to filter baremetal hosts", selector)
	bmhSelector := toDocumentSelector(selector)
	docs, err := i.inventoryBundle.Select(bmhSelector)
	if err != nil {
		log.Debugf("Failed to find BaremetalHosts")
		return nil, err
	}

	log.Debugf("Baremetal hosts count that matched the selector '%v' is '%d'", selector, len(docs))
	hostList := []remoteifc.Client{}
	for _, doc := range docs {
		host, err := i.newHost(doc)
		if err != nil {
			return nil, err
		}
		hostList = append(hostList, host)
	}

	return hostList, nil
}

// SelectOne selects single host based on given selector, if more than or less than one host is found
// error is returned
func (i Inventory) SelectOne(selector ifc.BaremetalHostSelector) (remoteifc.Client, error) {
	log.Debugf("Using selector %v to filter one baremetal host", selector)
	bmhSelector := toDocumentSelector(selector)

	doc, err := i.inventoryBundle.SelectOne(bmhSelector)
	if err != nil {
		return nil, err
	}

	return i.newHost(doc)
}

// RunOperation runs specified operation against the hosts that would be filtered by selector.
// Options are ignored for now, when we implement concurency, they will be used.
func (i Inventory) RunOperation(
	ctx context.Context,
	op ifc.BaremetalOperation,
	selector ifc.BaremetalHostSelector,
	_ ifc.BaremetalBatchRunOptions) error {
	return errors.ErrNotImplemented{What: "RunOperation of the baremetal inventory interface"}
}

// Host implements baremetal host interface
type Host struct {
	remoteifc.Client
}

var _ remoteifc.Client = Host{}

func (i Inventory) newHost(doc document.Document) (Host, error) {
	address, err := document.GetBMHBMCAddress(doc)
	if err != nil {
		return Host{}, err
	}

	username, password, err := document.GetBMHBMCCredentials(doc, i.inventoryBundle)
	if err != nil {
		return Host{}, err
	}

	var clientFactory remoteifc.ClientFactory
	switch i.mgmtCfg.Type {
	case redfish.ClientType:
		clientFactory = redfish.ClientFactory
	case redfishdell.ClientType:
		clientFactory = redfishdell.ClientFactory
	default:
		return Host{}, ErrRemoteDriverNotSupported{
			BMHName:      doc.GetName(),
			BMHNamespace: doc.GetNamespace(),
			RemoteType:   i.mgmtCfg.Type,
		}
	}

	client, err := clientFactory(
		address,
		i.mgmtCfg.Insecure,
		i.mgmtCfg.UseProxy,
		username,
		password,
		i.mgmtCfg.SystemActionRetries,
		i.mgmtCfg.SystemRebootDelay)
	if err != nil {
		return Host{}, err
	}
	return Host{Client: client}, nil
}

func toDocumentSelector(selector ifc.BaremetalHostSelector) document.Selector {
	return document.NewSelector().
		ByKind(document.BareMetalHostKind).
		ByLabel(selector.LabelSelector).
		ByName(selector.Name).
		ByNamespace(selector.Namespace)
}
