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

package inventory

import (
	"context"
	"fmt"
	"io"
	"time"

	"opendev.org/airship/airshipctl/pkg/inventory/ifc"
	remoteifc "opendev.org/airship/airshipctl/pkg/remote/ifc"
)

// CommandOptions is used to store common variables from cmd flags for baremetal command group
type CommandOptions struct {
	All bool

	Labels    string
	Name      string
	Namespace string
	IsoURL    string
	Timeout   time.Duration

	Inventory ifc.Inventory
}

// NewOptions options constructor
func NewOptions(i ifc.Inventory) *CommandOptions {
	return &CommandOptions{
		Inventory: i,
	}
}

func (o *CommandOptions) validateBMHAction() error {
	if o.Name == "" && o.Namespace == "" && o.Labels == "" && !o.All {
		return ErrInvalidOptions{Message: `must provide atleast one of the following options: ` +
			`'name', 'namespace', 'labels' or 'all'`}
	} else if o.All && (o.Name != "" || o.Namespace != "" || o.Labels != "") {
		return ErrInvalidOptions{Message: "option 'all' can not be combined with other host selector options"}
	}
	return nil
}

func (o *CommandOptions) validateSingleHostAction() error {
	if o.Name == "" && o.Namespace == "" && o.Labels == "" {
		return ErrInvalidOptions{Message: "No options are specified, must provide atleast 'name', 'namespace' or 'labels'"}
	}
	return nil
}

// BMHAction performs an action against BaremetalHost objects
func (o *CommandOptions) BMHAction(op ifc.BaremetalOperation) error {
	if err := o.validateBMHAction(); err != nil {
		return err
	}

	bmhInventory, err := o.Inventory.BaremetalInventory()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), o.Timeout)
	defer cancel()
	return bmhInventory.RunOperation(
		ctx,
		op,
		o.selector(),
		ifc.BaremetalBatchRunOptions{})
}

// RemoteDirect perform RemoteDirect operation against single host
func (o *CommandOptions) RemoteDirect() error {
	if err := o.validateSingleHostAction(); err != nil {
		return err
	}
	host, err := o.getHost()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), o.Timeout)
	defer cancel()
	return host.RemoteDirect(ctx, o.IsoURL)
}

// PowerStatus get power status of the single host
func (o *CommandOptions) PowerStatus(w io.Writer) error {
	if err := o.validateSingleHostAction(); err != nil {
		return err
	}
	host, err := o.getHost()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), o.Timeout)
	defer cancel()
	status, err := host.SystemPowerStatus(ctx)
	if err != nil {
		return err
	}
	// TODO support different output formats
	fmt.Fprintf(w, "Host with node id '%s' has power status: '%s'\n", host.NodeID(), status)
	return nil
}

func (o *CommandOptions) getHost() (remoteifc.Client, error) {
	bmhInventory, err := o.Inventory.BaremetalInventory()
	if err != nil {
		return nil, err
	}

	return bmhInventory.SelectOne(o.selector())
}

func (o *CommandOptions) selector() ifc.BaremetalHostSelector {
	return (ifc.BaremetalHostSelector{}).
		ByLabel(o.Labels).
		ByName(o.Name).
		ByNamespace(o.Namespace)
}
