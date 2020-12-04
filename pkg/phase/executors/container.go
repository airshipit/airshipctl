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
	"sigs.k8s.io/kustomize/kyaml/runfn"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
	"sigs.k8s.io/yaml"

	"k8s.io/apimachinery/pkg/runtime/schema"

	"opendev.org/airship/airshipctl/pkg/api/v1alpha1"
	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/errors"
	"opendev.org/airship/airshipctl/pkg/events"
	"opendev.org/airship/airshipctl/pkg/phase/ifc"
)

var _ ifc.Executor = &ContainerExecutor{}

// ContainerExecutor contains resources for generic container executor
type ContainerExecutor struct {
	ExecutorBundle   document.Bundle
	ExecutorDocument document.Document

	ContConf   *v1alpha1.GenericContainer
	RunFns     runfn.RunFns
	targetPath string
}

// RegisterContainerExecutor adds executor to phase executor registry
func RegisterContainerExecutor(registry map[schema.GroupVersionKind]ifc.ExecutorFactory) error {
	obj := v1alpha1.DefaultGenericContainer()
	gvks, _, err := v1alpha1.Scheme.ObjectKinds(obj)
	if err != nil {
		return err
	}
	registry[gvks[0]] = NewContainerExecutor
	return nil
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
		ExecutorBundle:   bundle,
		ExecutorDocument: cfg.ExecutorDocument,

		ContConf: apiObj,
		RunFns: runfn.RunFns{
			Functions: []*kyaml.RNode{},
		},
		targetPath: cfg.Helper.TargetPath(),
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

	if c.ContConf.PrintOutput {
		c.RunFns.Output = os.Stdout
	}

	if err := c.RunFns.Execute(); err != nil {
		handleError(evtCh, err)
		return
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
		storageMounts[i].Src = filepath.Join(c.targetPath, mount.Src)
	}
	c.RunFns.StorageMounts = storageMounts
}

// Validate executor configuration and documents
func (c *ContainerExecutor) Validate() error {
	return errors.ErrNotImplemented{}
}

// Render executor documents
func (c *ContainerExecutor) Render(_ io.Writer, _ ifc.RenderOptions) error {
	return errors.ErrNotImplemented{}
}
