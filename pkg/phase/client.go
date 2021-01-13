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

package phase

import (
	"io"
	"path/filepath"

	"k8s.io/apimachinery/pkg/runtime/schema"

	"opendev.org/airship/airshipctl/pkg/api/v1alpha1"
	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/events"
	"opendev.org/airship/airshipctl/pkg/k8s/kubeconfig"
	"opendev.org/airship/airshipctl/pkg/k8s/utils"
	"opendev.org/airship/airshipctl/pkg/log"
	"opendev.org/airship/airshipctl/pkg/phase/executors"
	"opendev.org/airship/airshipctl/pkg/phase/ifc"
)

// ExecutorRegistry returns map with executor factories
type ExecutorRegistry func() map[schema.GroupVersionKind]ifc.ExecutorFactory

// DefaultExecutorRegistry returns map with executor factories
func DefaultExecutorRegistry() map[schema.GroupVersionKind]ifc.ExecutorFactory {
	execMap := make(map[schema.GroupVersionKind]ifc.ExecutorFactory)

	for _, execName := range []string{executors.Clusterctl, executors.KubernetesApply,
		executors.Isogen, executors.GenericContainer, executors.Ephemeral} {
		if err := executors.RegisterExecutor(execName, execMap); err != nil {
			log.Fatal(ErrExecutorRegistration{ExecutorName: execName, Err: err})
		}
	}
	return execMap
}

var _ ifc.Phase = &phase{}

// Phase implements phase interface
type phase struct {
	helper     ifc.Helper
	apiObj     *v1alpha1.Phase
	registry   ExecutorRegistry
	processor  events.EventProcessor
	kubeconfig string
}

// Executor returns executor interface associated with the phase
func (p *phase) Executor() (ifc.Executor, error) {
	executorDoc, err := p.helper.ExecutorDoc(ifc.ID{Name: p.apiObj.Name, Namespace: p.apiObj.Namespace})
	if err != nil {
		return nil, err
	}

	var bundleFactory document.BundleFactoryFunc = func() (document.Bundle, error) {
		docRoot, bundleFactoryFuncErr := p.DocumentRoot()
		if bundleFactoryFuncErr != nil {
			return nil, bundleFactoryFuncErr
		}
		return document.NewBundleByPath(docRoot)
	}

	refGVK := p.apiObj.Config.ExecutorRef.GroupVersionKind()
	// Look for executor factory defined in registry
	executorFactory, found := p.registry()[refGVK]
	if !found {
		return nil, ErrExecutorNotFound{GVK: refGVK}
	}

	cMap, err := p.helper.ClusterMap()
	if err != nil {
		return nil, err
	}

	wd, err := p.helper.WorkDir()
	if err != nil {
		return nil, err
	}
	kubeconf := kubeconfig.NewBuilder().
		WithBundle(p.helper.PhaseBundleRoot()).
		WithClusterMap(cMap).
		WithClusterName(p.apiObj.ClusterName).
		WithPath(p.kubeconfig).
		WithTempRoot(wd).
		Build()

	return executorFactory(
		ifc.ExecutorConfig{
			ClusterMap:       cMap,
			BundleFactory:    bundleFactory,
			PhaseName:        p.apiObj.Name,
			KubeConfig:       kubeconf,
			ExecutorDocument: executorDoc,
			ClusterName:      p.apiObj.ClusterName,
			Helper:           p.helper,
		})
}

// Run runs the phase via executor
func (p *phase) Run(ro ifc.RunOptions) error {
	defer p.processor.Close()
	executor, err := p.Executor()
	if err != nil {
		return err
	}
	ch := make(chan events.Event)

	go func() {
		executor.Run(ch, ro)
	}()
	return p.processor.Process(ch)
}

// Validate makes sure that phase is properly configured
// TODO implement this
func (p *phase) Validate() error {
	return nil
}

// Render executor documents
func (p *phase) Render(w io.Writer, executorRender bool, options ifc.RenderOptions) error {
	if executorRender {
		executor, err := p.Executor()
		if err != nil {
			return err
		}
		return executor.Render(w, options)
	}

	root, err := p.DocumentRoot()
	if err != nil {
		return err
	}

	bundle, err := document.NewBundleByPath(root)
	if err != nil {
		return err
	}

	rendered, err := bundle.SelectBundle(options.FilterSelector)
	if err != nil {
		return err
	}
	return rendered.Write(w)
}

// DocumentRoot root that holds all the documents associated with the phase
func (p *phase) DocumentRoot() (string, error) {
	relativePath := p.apiObj.Config.DocumentEntryPoint
	if relativePath == "" {
		return "", ErrDocumentEntrypointNotDefined{
			PhaseName:      p.apiObj.Name,
			PhaseNamespace: p.apiObj.Namespace,
		}
	}

	phaseEntryPointBasePath := p.helper.PhaseEntryPointBasePath()
	return filepath.Join(phaseEntryPointBasePath, relativePath), nil
}

// Details returns description of the phase
// TODO implement this: add details field to api.Phase and method to executor and combine them here
// to give a clear understanding to user of what this phase is about
func (p *phase) Details() (string, error) {
	return "", nil
}

var _ ifc.Client = &client{}

type client struct {
	ifc.Helper

	registry      ExecutorRegistry
	processorFunc ProcessorFunc
	kubeconfig    string
}

// ProcessorFunc that returns processor interface
type ProcessorFunc func() events.EventProcessor

// Option allows to add various options to a phase
type Option func(*client)

// InjectProcessor is an option that allows to inject event processor into phase client
func InjectProcessor(procFunc ProcessorFunc) Option {
	return func(c *client) {
		c.processorFunc = procFunc
	}
}

// InjectRegistry is an option that allows to inject executor registry into phase client
func InjectRegistry(registry ExecutorRegistry) Option {
	return func(c *client) {
		c.registry = registry
	}
}

// InjectKubeconfigPath is an option that allows to inject path to kubeconfig into phase client
func InjectKubeconfigPath(path string) Option {
	return func(c *client) {
		c.kubeconfig = path
	}
}

// NewClient returns implementation of phase Client interface
func NewClient(helper ifc.Helper, opts ...Option) ifc.Client {
	c := &client{Helper: helper}
	for _, opt := range opts {
		opt(c)
	}
	if c.registry == nil {
		c.registry = DefaultExecutorRegistry
	}
	if c.processorFunc == nil {
		c.processorFunc = defaultProcessor
	}
	return c
}

func (c *client) PhaseByID(id ifc.ID) (ifc.Phase, error) {
	phaseObj, err := c.Phase(id)
	if err != nil {
		return nil, err
	}

	phase := &phase{
		apiObj:     phaseObj,
		helper:     c.Helper,
		processor:  c.processorFunc(),
		registry:   c.registry,
		kubeconfig: c.kubeconfig,
	}
	return phase, nil
}

func (c *client) PhaseByAPIObj(phaseObj *v1alpha1.Phase) (ifc.Phase, error) {
	phase := &phase{
		apiObj:     phaseObj,
		helper:     c.Helper,
		processor:  c.processorFunc(),
		registry:   c.registry,
		kubeconfig: c.kubeconfig,
	}
	return phase, nil
}

func defaultProcessor() events.EventProcessor {
	return events.NewDefaultProcessor(utils.Streams())
}
