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

package isogen

import (
	"context"
	"io"

	"k8s.io/apimachinery/pkg/runtime/schema"

	"opendev.org/airship/airshipctl/pkg/api/v1alpha1"
	"opendev.org/airship/airshipctl/pkg/container"
	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/errors"
	"opendev.org/airship/airshipctl/pkg/events"
	"opendev.org/airship/airshipctl/pkg/log"
	"opendev.org/airship/airshipctl/pkg/phase/ifc"
)

var _ ifc.Executor = &Executor{}

// Executor contains resources for isogen executor
type Executor struct {
	ExecutorBundle   document.Bundle
	ExecutorDocument document.Document

	imgConf *v1alpha1.ImageConfiguration
	builder container.Container
}

// RegisterExecutor adds executor to phase executor registry
func RegisterExecutor(registry map[schema.GroupVersionKind]ifc.ExecutorFactory) error {
	obj := &v1alpha1.ImageConfiguration{}
	gvks, _, err := v1alpha1.Scheme.ObjectKinds(obj)
	if err != nil {
		return err
	}
	registry[gvks[0]] = NewExecutor
	return nil
}

// NewExecutor creates instance of phase executor
func NewExecutor(cfg ifc.ExecutorConfig) (ifc.Executor, error) {
	apiObj := &v1alpha1.ImageConfiguration{
		Container: &v1alpha1.Container{},
		Builder:   &v1alpha1.Builder{},
	}
	err := cfg.ExecutorDocument.ToAPIObject(apiObj, v1alpha1.Scheme)
	if err != nil {
		return nil, err
	}

	return &Executor{
		ExecutorBundle:   cfg.ExecutorBundle,
		ExecutorDocument: cfg.ExecutorDocument,
		imgConf:          apiObj,
	}, nil
}

// Run isogen as a phase runner
func (c *Executor) Run(evtCh chan events.Event, opts ifc.RunOptions) {
	defer close(evtCh)

	if c.ExecutorBundle == nil {
		handleError(evtCh, ErrIsoGenNilBundle{})
		return
	}

	evtCh <- events.Event{
		Type: events.IsogenType,
		IsogenEvent: events.IsogenEvent{
			Operation: events.IsogenStart,
			Message:   "starting ISO generation",
		},
	}

	if opts.DryRun {
		log.Print("command isogen will be executed")
		evtCh <- events.Event{
			Type: events.IsogenType,
			IsogenEvent: events.IsogenEvent{
				Operation: events.IsogenEnd,
			},
		}
		return
	}

	if c.builder == nil {
		ctx := context.Background()
		builder, err := container.NewContainer(
			&ctx,
			c.imgConf.Container.ContainerRuntime,
			c.imgConf.Container.Image)
		c.builder = builder
		if err != nil {
			handleError(evtCh, err)
			return
		}
	}

	err := createBootstrapIso(c.ExecutorBundle, c.builder, c.ExecutorDocument, c.imgConf, log.DebugEnabled(), false)
	if err != nil {
		handleError(evtCh, err)
		return
	}

	evtCh <- events.Event{
		Type: events.IsogenType,
		IsogenEvent: events.IsogenEvent{
			Operation: events.IsogenValidation,
			Message:   "image is generated successfully, verifying artifacts",
		},
	}
	err = verifyArtifacts(c.imgConf)
	if err != nil {
		handleError(evtCh, err)
		return
	}

	evtCh <- events.Event{
		Type: events.IsogenType,
		IsogenEvent: events.IsogenEvent{
			Operation: events.IsogenEnd,
			Message:   "iso generation is complete and artifacts verified",
		},
	}
}

// Validate executor configuration and documents
func (c *Executor) Validate() error {
	return errors.ErrNotImplemented{}
}

// Render executor documents
func (c *Executor) Render(w io.Writer, _ ifc.RenderOptions) error {
	// will be implemented later
	_, err := w.Write([]byte{})
	return err
}

func handleError(ch chan<- events.Event, err error) {
	ch <- events.Event{
		Type: events.ErrorType,
		ErrorEvent: events.ErrorEvent{
			Error: err,
		},
	}
}
