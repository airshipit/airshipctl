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

package executors

import (
	"fmt"
	"io"
	"time"

	"opendev.org/airship/airshipctl/pkg/api/v1alpha1"
	airshipv1 "opendev.org/airship/airshipctl/pkg/api/v1alpha1"
	"opendev.org/airship/airshipctl/pkg/errors"
	"opendev.org/airship/airshipctl/pkg/events"
	"opendev.org/airship/airshipctl/pkg/inventory"
	inventoryifc "opendev.org/airship/airshipctl/pkg/inventory/ifc"
	"opendev.org/airship/airshipctl/pkg/phase/ifc"
)

// BaremetalManagerExecutor is abstraction built on top of baremetal commands of airshipctl
type BaremetalManagerExecutor struct {
	inventory inventoryifc.Inventory
	options   *airshipv1.BaremetalManager
}

// NewBaremetalExecutor constructor for baremetal executor
func NewBaremetalExecutor(cfg ifc.ExecutorConfig) (ifc.Executor, error) {
	inv, err := cfg.Helper.Inventory()
	if err != nil {
		return nil, err
	}
	options := airshipv1.DefaultBaremetalManager()
	if err := cfg.ExecutorDocument.ToAPIObject(options, airshipv1.Scheme); err != nil {
		return nil, err
	}
	return &BaremetalManagerExecutor{
		inventory: inv,
		options:   options,
	}, nil
}

// Run runs baremetal operations as executor
func (e *BaremetalManagerExecutor) Run(evtCh chan events.Event, opts ifc.RunOptions) {
	defer close(evtCh)
	commandOptions := toCommandOptions(e.inventory, e.options.Spec, opts)

	evtCh <- events.NewEvent().WithBaremetalManagerEvent(events.BaremetalManagerEvent{
		Step:          events.BaremetalManagerStart,
		HostOperation: string(e.options.Spec.Operation),
		Message: fmt.Sprintf("Starting remote operation '%s', selector to be to filter hosts %v",
			e.options.Spec.Operation, e.options.Spec.HostSelector),
	})

	op, err := e.validate()
	if err != nil {
		handleError(evtCh, err)
		return
	}
	if !opts.DryRun {
		switch e.options.Spec.Operation {
		case airshipv1.BaremetalOperationPowerOn, airshipv1.BaremetalOperationPowerOff,
			airshipv1.BaremetalOperationReboot, airshipv1.BaremetalOperationEjectVirtualMedia:
			err = commandOptions.BMHAction(op)
		case airshipv1.BaremetalOperationRemoteDirect:
			err = commandOptions.RemoteDirect()
		}
	}

	if err != nil {
		handleError(evtCh, err)
		return
	}

	evtCh <- events.NewEvent().WithBaremetalManagerEvent(events.BaremetalManagerEvent{
		Step:          events.BaremetalManagerComplete,
		HostOperation: string(e.options.Spec.Operation),
		Message: fmt.Sprintf("Successfully completed operation against host selected by selector %v",
			e.options.Spec.HostSelector),
	})
}

// Validate executor configuration and documents
func (e *BaremetalManagerExecutor) Validate() error {
	_, err := e.validate()
	return err
}

func (e *BaremetalManagerExecutor) validate() (inventoryifc.BaremetalOperation, error) {
	var result inventoryifc.BaremetalOperation
	var err error
	switch e.options.Spec.Operation {
	case airshipv1.BaremetalOperationPowerOn:
		result = inventoryifc.BaremetalOperationPowerOn
	case airshipv1.BaremetalOperationPowerOff:
		result = inventoryifc.BaremetalOperationPowerOff
	case airshipv1.BaremetalOperationEjectVirtualMedia:
		result = inventoryifc.BaremetalOperationEjectVirtualMedia
	case airshipv1.BaremetalOperationReboot:
		result = inventoryifc.BaremetalOperationReboot
	case airshipv1.BaremetalOperationRemoteDirect:
		// TODO add remote direct validation, make sure that ISO-URL is specified
		result = ""
	default:
		err = ErrUnknownExecutorAction{Action: string(e.options.Spec.Operation), ExecutorName: BMHManager}
	}
	return result, err
}

// Render baremetal hosts
func (e *BaremetalManagerExecutor) Render(w io.Writer, _ ifc.RenderOptions) error {
	// add printing of baremetal hosts here
	_, err := w.Write([]byte{})
	return err
}

func toCommandOptions(i inventoryifc.Inventory,
	spec v1alpha1.BaremetalManagerSpec,
	opts ifc.RunOptions) *inventory.CommandOptions {
	timeout := time.Duration(spec.Timeout) * time.Second
	if opts.Timeout != 0 {
		timeout = opts.Timeout
	}

	return &inventory.CommandOptions{
		Inventory: i,
		IsoURL:    spec.OperationOptions.RemoteDirect.ISOURL,
		Labels:    spec.HostSelector.LabelSelector,
		Name:      spec.HostSelector.Name,
		Namespace: spec.HostSelector.Namespace,
		Timeout:   timeout,
	}
}

// Status returns the status of the given phase
func (e *BaremetalManagerExecutor) Status() (ifc.ExecutorStatus, error) {
	return ifc.ExecutorStatus{}, errors.ErrNotImplemented{What: BMHManager}
}
