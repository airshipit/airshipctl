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
	"opendev.org/airship/airshipctl/pkg/bootstrap/isogen"
	clusterctl "opendev.org/airship/airshipctl/pkg/clusterctl/client"
	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/events"
	"opendev.org/airship/airshipctl/pkg/k8s/applier"
	"opendev.org/airship/airshipctl/pkg/k8s/kubeconfig"
	"opendev.org/airship/airshipctl/pkg/k8s/utils"
	"opendev.org/airship/airshipctl/pkg/log"
	"opendev.org/airship/airshipctl/pkg/phase/ifc"
)

// ExecutorRegistry returns map with executor factories
type ExecutorRegistry func() map[schema.GroupVersionKind]ifc.ExecutorFactory

// DefaultExecutorRegistry returns map with executor factories
func DefaultExecutorRegistry() map[schema.GroupVersionKind]ifc.ExecutorFactory {
	execMap := make(map[schema.GroupVersionKind]ifc.ExecutorFactory)

	if err := clusterctl.RegisterExecutor(execMap); err != nil {
		log.Fatal(ErrExecutorRegistration{ExecutorName: "clusterctl", Err: err})
	}
	if err := applier.RegisterExecutor(execMap); err != nil {
		log.Fatal(ErrExecutorRegistration{ExecutorName: "kubernetes-apply", Err: err})
	}
	if err := isogen.RegisterExecutor(execMap); err != nil {
		log.Fatal(ErrExecutorRegistration{ExecutorName: "isogen", Err: err})
	}
	return execMap
}

var _ ifc.Phase = &phase{}

// Phase implements phase interface
type phase struct {
	helper    ifc.Helper
	apiObj    *v1alpha1.Phase
	registry  ExecutorRegistry
	processor events.EventProcessor
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
		WithBundle(p.helper.PhaseRoot()).
		WithClusterMap(cMap).
		WithClusterName(p.apiObj.ClusterName).
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
func (p *phase) Render(w io.Writer, options ifc.RenderOptions) error {
	executor, err := p.Executor()
	if err != nil {
		return err
	}

	return executor.Render(w, options)
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

	targetPath := p.helper.TargetPath()
	return filepath.Join(targetPath, relativePath), nil
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
		apiObj:    phaseObj,
		helper:    c.Helper,
		processor: c.processorFunc(),
		registry:  c.registry,
	}
	return phase, nil
}

func (c *client) PhaseByAPIObj(phaseObj *v1alpha1.Phase) (ifc.Phase, error) {
	phase := &phase{
		apiObj:    phaseObj,
		helper:    c.Helper,
		processor: c.processorFunc(),
		registry:  c.registry,
	}
	return phase, nil
}

func defaultProcessor() events.EventProcessor {
	return events.NewDefaultProcessor(utils.Streams())
}
