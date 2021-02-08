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
	"bytes"
	"io"
	"os"
	"path/filepath"

	"opendev.org/airship/airshipctl/pkg/api/v1alpha1"
	"opendev.org/airship/airshipctl/pkg/container"
	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/errors"
	"opendev.org/airship/airshipctl/pkg/events"
	"opendev.org/airship/airshipctl/pkg/phase/ifc"
)

var _ ifc.Executor = &ContainerExecutor{}

// ContainerExecutor contains resources for generic container executor
type ContainerExecutor struct {
	ResultsDir string

	Container        *v1alpha1.GenericContainer
	ClientFunc       container.ClientV1Alpha1FactoryFunc
	ExecutorBundle   document.Bundle
	ExecutorDocument document.Document
}

// NewContainerExecutor creates instance of phase executor
func NewContainerExecutor(cfg ifc.ExecutorConfig) (ifc.Executor, error) {
	// TODO add logic that checks if the path was not defined, and if so, we are fine
	// and bundle should be either nil or empty, consider ContinueOnEmptyInput option to container client
	bundle, err := cfg.BundleFactory()
	if err != nil {
		return nil, err
	}

	apiObj := v1alpha1.DefaultGenericContainer()
	err = cfg.ExecutorDocument.ToAPIObject(apiObj, v1alpha1.Scheme)
	if err != nil {
		return nil, err
	}

	var resultsDir string
	if apiObj.Spec.SinkOutputDir != "" {
		resultsDir = filepath.Join(cfg.Helper.PhaseEntryPointBasePath(), apiObj.Spec.SinkOutputDir)
	}

	return &ContainerExecutor{
		ResultsDir:       resultsDir,
		ExecutorBundle:   bundle,
		ExecutorDocument: cfg.ExecutorDocument,
		// TODO extend tests with proper client, make it interface
		ClientFunc: container.NewClientV1Alpha1,

		Container: apiObj,
	}, nil
}

// Run generic container as a phase runner
func (c *ContainerExecutor) Run(evtCh chan events.Event, opts ifc.RunOptions) {
	defer close(evtCh)

	evtCh <- events.NewEvent().WithGenericContainerEvent(events.GenericContainerEvent{
		Operation: events.GenericContainerStart,
		Message:   "starting generic container",
	})

	input, err := bundleReader(c.ExecutorBundle)
	if err != nil {
		// TODO move bundleFactory here, and make sure that if executorDoc is not defined, we dont fail
		handleError(evtCh, err)
		return
	}

	// TODO this logic is redundant in executor package, move it to pkg/container
	var output io.Writer
	if c.ResultsDir == "" {
		// set output only if the output if resulting directory is not defined
		output = os.Stdout
	}

	// TODO check the executor type  when dryrun is set
	if opts.DryRun {
		evtCh <- events.NewEvent().WithGenericContainerEvent(events.GenericContainerEvent{
			Operation: events.GenericContainerStop,
			Message:   "DryRun execution finished",
		})
		return
	}

	err = c.ClientFunc(c.ResultsDir, input, output, c.Container).Run()
	if err != nil {
		handleError(evtCh, err)
		return
	}

	evtCh <- events.NewEvent().WithGenericContainerEvent(events.GenericContainerEvent{
		Operation: events.GenericContainerStop,
		Message:   "execution of the generic container finished",
	})
}

// bundleReader sets input for function
func bundleReader(bundle document.Bundle) (io.Reader, error) {
	buf := &bytes.Buffer{}
	return buf, bundle.Write(buf)
}

// Validate executor configuration and documents
func (c *ContainerExecutor) Validate() error {
	return errors.ErrNotImplemented{}
}

// Render executor documents
func (c *ContainerExecutor) Render(_ io.Writer, _ ifc.RenderOptions) error {
	return errors.ErrNotImplemented{}
}
