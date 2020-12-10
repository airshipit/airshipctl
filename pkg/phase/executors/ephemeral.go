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
	"context"
	"io"
	"time"

	"k8s.io/apimachinery/pkg/runtime/schema"

	"opendev.org/airship/airshipctl/pkg/api/v1alpha1"
	"opendev.org/airship/airshipctl/pkg/bootstrap/ephemeral"
	"opendev.org/airship/airshipctl/pkg/container"
	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/errors"
	"opendev.org/airship/airshipctl/pkg/events"
	"opendev.org/airship/airshipctl/pkg/log"
	"opendev.org/airship/airshipctl/pkg/phase/ifc"
)

var _ ifc.Executor = &EphemeralExecutor{}

// EphemeralExecutor contains resources for ephemeral executor
type EphemeralExecutor struct {
	ExecutorBundle   document.Bundle
	ExecutorDocument document.Document

	BootConf  *v1alpha1.BootConfiguration
	Container container.Container
}

// RegisterEphemeralExecutor adds executor to phase executor registry
func RegisterEphemeralExecutor(registry map[schema.GroupVersionKind]ifc.ExecutorFactory) error {
	obj := v1alpha1.DefaultBootConfiguration()
	gvks, _, err := v1alpha1.Scheme.ObjectKinds(obj)
	if err != nil {
		return err
	}
	registry[gvks[0]] = NewEphemeralExecutor
	return nil
}

// NewEphemeralExecutor creates instance of phase executor
func NewEphemeralExecutor(cfg ifc.ExecutorConfig) (ifc.Executor, error) {
	apiObj := &v1alpha1.BootConfiguration{}

	err := cfg.ExecutorDocument.ToAPIObject(apiObj, v1alpha1.Scheme)
	if err != nil {
		return nil, err
	}

	return &EphemeralExecutor{
		ExecutorDocument: cfg.ExecutorDocument,
		BootConf:         apiObj,
	}, nil
}

// Run ephemeral as a phase runner
func (c *EphemeralExecutor) Run(evtCh chan events.Event, opts ifc.RunOptions) {
	defer close(evtCh)

	evtCh <- events.NewEvent().WithBootstrapEvent(events.BootstrapEvent{
		Operation: events.BootstrapStart,
		Message:   "Processing Ephemeral cluster operation ...",
	})

	if opts.DryRun {
		log.Print("Dryrun: bootstrap container command will be skipped")
		evtCh <- events.NewEvent().WithBootstrapEvent(events.BootstrapEvent{
			Operation: events.BootstrapDryRun,
		})
		return
	}

	if c.Container == nil {
		ctx := context.Background()
		builder, err := container.NewContainer(
			ctx,
			c.BootConf.BootstrapContainer.ContainerRuntime,
			c.BootConf.BootstrapContainer.Image)
		if err != nil {
			handleEphemeralError(evtCh, err)
			return
		}
		c.Container = builder
	}

	bootstrapOpts := ephemeral.BootstrapContainerOptions{
		Container: c.Container,
		Cfg:       c.BootConf,
		Sleep:     time.Sleep,
	}

	evtCh <- events.NewEvent().WithBootstrapEvent(events.BootstrapEvent{
		Operation: events.BootstrapValidation,
		Message:   "Verifying executor manifest document ...",
	})

	err := bootstrapOpts.VerifyInputs()
	if err != nil {
		handleEphemeralError(evtCh, err)
		return
	}

	evtCh <- events.NewEvent().WithBootstrapEvent(events.BootstrapEvent{
		Operation: events.BootstrapRun,
		Message:   "Creating and starting the Bootstrap Container ...",
	})

	err = bootstrapOpts.CreateBootstrapContainer()
	if err != nil {
		handleEphemeralError(evtCh, err)
		return
	}

	evtCh <- events.NewEvent().WithBootstrapEvent(events.BootstrapEvent{
		Operation: events.BootstrapValidation,
		Message:   "Verifying generation of kubeconfig file ...",
	})

	err = bootstrapOpts.VerifyArtifacts()
	if err != nil {
		handleEphemeralError(evtCh, err)
		return
	}

	evtCh <- events.NewEvent().WithBootstrapEvent(events.BootstrapEvent{
		Operation: events.BootstrapEnd,
		Message:   "Ephemeral cluster operation has completed successfully and artifacts verified",
	})
}

// Validate executor configuration and documents
func (c *EphemeralExecutor) Validate() error {
	return errors.ErrNotImplemented{}
}

// Render executor documents
func (c *EphemeralExecutor) Render(w io.Writer, _ ifc.RenderOptions) error {
	log.Print("Ephemeral Executor Render() will be implemented later.")
	return nil
}

func handleEphemeralError(ch chan<- events.Event, err error) {
	ch <- events.Event{
		Type: events.ErrorType,
		ErrorEvent: events.ErrorEvent{
			Error: err,
		},
	}
}
