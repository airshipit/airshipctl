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
	"os"
	"path/filepath"
	"strings"

	"k8s.io/apimachinery/pkg/runtime/schema"

	"opendev.org/airship/airshipctl/pkg/api/v1alpha1"
	"opendev.org/airship/airshipctl/pkg/bootstrap/isogen"
	"opendev.org/airship/airshipctl/pkg/container"
	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/errors"
	"opendev.org/airship/airshipctl/pkg/events"
	"opendev.org/airship/airshipctl/pkg/log"
	"opendev.org/airship/airshipctl/pkg/phase/ifc"
)

var _ ifc.Executor = &IsogenExecutor{}

// IsogenExecutor contains resources for isogen executor
type IsogenExecutor struct {
	ExecutorBundle   document.Bundle
	ExecutorDocument document.Document

	ImgConf *v1alpha1.IsoConfiguration
	Builder container.Container
}

// RegisterIsogenExecutor adds executor to phase executor registry
func RegisterIsogenExecutor(registry map[schema.GroupVersionKind]ifc.ExecutorFactory) error {
	obj := v1alpha1.DefaultIsoConfiguration()
	gvks, _, err := v1alpha1.Scheme.ObjectKinds(obj)
	if err != nil {
		return err
	}
	registry[gvks[0]] = NewIsogenExecutor
	return nil
}

// NewIsogenExecutor creates instance of phase executor
func NewIsogenExecutor(cfg ifc.ExecutorConfig) (ifc.Executor, error) {
	apiObj := &v1alpha1.IsoConfiguration{
		IsoContainer: &v1alpha1.IsoContainer{},
		Isogen:       &v1alpha1.Isogen{},
	}
	err := cfg.ExecutorDocument.ToAPIObject(apiObj, v1alpha1.Scheme)
	if err != nil {
		return nil, err
	}

	bundle, err := cfg.BundleFactory()
	if err != nil {
		return nil, err
	}

	return &IsogenExecutor{
		ExecutorBundle:   bundle,
		ExecutorDocument: cfg.ExecutorDocument,
		ImgConf:          apiObj,
	}, nil
}

// Run isogen as a phase runner
func (c *IsogenExecutor) Run(evtCh chan events.Event, opts ifc.RunOptions) {
	defer close(evtCh)

	if c.ExecutorBundle == nil {
		handleError(evtCh, ErrIsogenNilBundle{})
		return
	}

	evtCh <- events.NewEvent().WithIsogenEvent(events.IsogenEvent{
		Operation: events.IsogenStart,
		Message:   "starting ISO generation",
	})

	if opts.DryRun {
		log.Print("command isogen will be executed")
		evtCh <- events.NewEvent().WithIsogenEvent(events.IsogenEvent{
			Operation: events.IsogenEnd,
		})
		return
	}

	if c.Builder == nil {
		ctx := context.Background()
		builder, err := container.NewContainer(
			ctx,
			c.ImgConf.IsoContainer.ContainerRuntime,
			c.ImgConf.IsoContainer.Image)
		c.Builder = builder
		if err != nil {
			handleError(evtCh, err)
			return
		}
	}

	bootstrapOpts := isogen.BootstrapIsoOptions{
		DocBundle: c.ExecutorBundle,
		Builder:   c.Builder,
		Doc:       c.ExecutorDocument,
		Cfg:       c.ImgConf,
		Writer:    log.Writer(),
	}

	err := bootstrapOpts.CreateBootstrapIso()
	if err != nil {
		handleError(evtCh, err)
		return
	}

	evtCh <- events.NewEvent().WithIsogenEvent(events.IsogenEvent{
		Operation: events.IsogenValidation,
		Message:   "image is generated successfully, verifying artifacts",
	})
	err = verifyArtifacts(c.ImgConf)
	if err != nil {
		handleError(evtCh, err)
		return
	}

	evtCh <- events.NewEvent().WithIsogenEvent(events.IsogenEvent{
		Operation: events.IsogenEnd,
		Message:   "iso generation is complete and artifacts verified",
	})
}

func verifyArtifacts(cfg *v1alpha1.IsoConfiguration) error {
	hostVol := strings.Split(cfg.IsoContainer.Volume, ":")[0]
	outputFilePath := filepath.Join(hostVol, cfg.Isogen.OutputFileName)
	_, err := os.Stat(outputFilePath)
	return err
}

// Validate executor configuration and documents
func (c *IsogenExecutor) Validate() error {
	return errors.ErrNotImplemented{}
}

// Render executor documents
func (c *IsogenExecutor) Render(w io.Writer, _ ifc.RenderOptions) error {
	// will be implemented later
	_, err := w.Write([]byte{})
	return err
}

func handleError(ch chan<- events.Event, err error) {
	ch <- events.NewEvent().WithErrorEvent(events.ErrorEvent{
		Error: err,
	})
}
