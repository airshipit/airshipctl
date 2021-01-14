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
	"log"
	"os"
	"path/filepath"

	"sigs.k8s.io/kustomize/kyaml/fn/runtime/runtimeutil"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/runfn"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
	"sigs.k8s.io/yaml"

	"opendev.org/airship/airshipctl/pkg/api/v1alpha1"
	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/errors"
	"opendev.org/airship/airshipctl/pkg/events"
	"opendev.org/airship/airshipctl/pkg/phase/ifc"
)

var _ ifc.Executor = &ContainerExecutor{}

// ContainerExecutor contains resources for generic container executor
type ContainerExecutor struct {
	PhaseEntryPointBasePath string
	ExecutorBundle          document.Bundle
	ExecutorDocument        document.Document

	ContConf   *v1alpha1.GenericContainer
	RunFns     runfn.RunFns
	TargetPath string
}

// NewContainerExecutor creates instance of phase executor
func NewContainerExecutor(cfg ifc.ExecutorConfig) (ifc.Executor, error) {
	bundle, err := cfg.BundleFactory()
	if err != nil {
		return nil, err
	}

	apiObj := &v1alpha1.GenericContainer{
		Spec: runtimeutil.FunctionSpec{},
	}
	err = cfg.ExecutorDocument.ToAPIObject(apiObj, v1alpha1.Scheme)
	if err != nil {
		return nil, err
	}

	return &ContainerExecutor{
		PhaseEntryPointBasePath: cfg.Helper.PhaseEntryPointBasePath(),
		ExecutorBundle:          bundle,
		ExecutorDocument:        cfg.ExecutorDocument,

		ContConf: apiObj,
		RunFns: runfn.RunFns{
			Functions: []*kyaml.RNode{},
		},
		TargetPath: cfg.Helper.TargetPath(),
	}, nil
}

// Run generic container as a phase runner
func (c *ContainerExecutor) Run(evtCh chan events.Event, opts ifc.RunOptions) {
	defer close(evtCh)

	evtCh <- events.NewEvent().WithGenericContainerEvent(events.GenericContainerEvent{
		Operation: events.GenericContainerStart,
		Message:   "starting generic container",
	})

	if opts.DryRun {
		log.Print("generic container will be executed")
		evtCh <- events.NewEvent().WithGenericContainerEvent(events.GenericContainerEvent{
			Operation: events.GenericContainerStop,
			Message:   "DryRun execution finished",
		})
		return
	}

	if err := c.SetInput(); err != nil {
		handleError(evtCh, err)
		return
	}

	if err := c.PrepareFunctions(); err != nil {
		handleError(evtCh, err)
		return
	}

	c.SetMounts()

	var fnsOutputBuffer bytes.Buffer

	if c.ContConf.KustomizeSinkOutputDir != "" {
		c.RunFns.Output = &fnsOutputBuffer
	} else {
		c.RunFns.Output = os.Stdout
	}

	if err := c.RunFns.Execute(); err != nil {
		handleError(evtCh, err)
		return
	}

	if c.ContConf.KustomizeSinkOutputDir != "" {
		if err := c.WriteKustomizeSink(&fnsOutputBuffer); err != nil {
			handleError(evtCh, err)
			return
		}
	}

	evtCh <- events.NewEvent().WithGenericContainerEvent(events.GenericContainerEvent{
		Operation: events.GenericContainerStop,
		Message:   "execution of the generic container finished",
	})
}

// SetInput sets input for function
func (c *ContainerExecutor) SetInput() error {
	buf := &bytes.Buffer{}
	err := c.ExecutorBundle.Write(buf)
	if err != nil {
		return err
	}

	c.RunFns.Input = buf
	return nil
}

// PrepareFunctions prepares data for function
func (c *ContainerExecutor) PrepareFunctions() error {
	rnode, err := kyaml.Parse(c.ContConf.Config)
	if err != nil {
		return err
	}
	// Transform GenericContainer.Spec to annotation,
	// because we need to specify runFns config in annotation
	spec, err := yaml.Marshal(c.ContConf.Spec)
	if err != nil {
		return err
	}
	annotation := kyaml.SetAnnotation(runtimeutil.FunctionAnnotationKey, string(spec))
	_, err = annotation.Filter(rnode)
	if err != nil {
		return err
	}

	c.RunFns.Functions = append(c.RunFns.Functions, rnode)

	return nil
}

// SetMounts allows to set relative path for storage mounts to prevent security issues
func (c *ContainerExecutor) SetMounts() {
	if len(c.ContConf.Spec.Container.StorageMounts) == 0 {
		return
	}
	storageMounts := c.ContConf.Spec.Container.StorageMounts
	for i, mount := range storageMounts {
		storageMounts[i].Src = filepath.Join(c.TargetPath, mount.Src)
	}
	c.RunFns.StorageMounts = storageMounts
}

// WriteKustomizeSink writes output to kustomize sink
func (c *ContainerExecutor) WriteKustomizeSink(fnsOutputBuffer *bytes.Buffer) error {
	outputDirPath := filepath.Join(c.PhaseEntryPointBasePath, c.ContConf.KustomizeSinkOutputDir)
	sinkOutputs := []kio.Writer{&kio.LocalPackageWriter{PackagePath: outputDirPath}}
	err := kio.Pipeline{
		Inputs:  []kio.Reader{&kio.ByteReader{Reader: fnsOutputBuffer}},
		Outputs: sinkOutputs}.Execute()
	return err
}

// Validate executor configuration and documents
func (c *ContainerExecutor) Validate() error {
	return errors.ErrNotImplemented{}
}

// Render executor documents
func (c *ContainerExecutor) Render(_ io.Writer, _ ifc.RenderOptions) error {
	return errors.ErrNotImplemented{}
}
