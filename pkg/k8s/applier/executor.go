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

package applier

import (
	"io"
	"time"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/cli-utils/pkg/common"

	airshipv1 "opendev.org/airship/airshipctl/pkg/api/v1alpha1"
	"opendev.org/airship/airshipctl/pkg/cluster/clustermap"
	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/errors"
	"opendev.org/airship/airshipctl/pkg/events"
	"opendev.org/airship/airshipctl/pkg/k8s/kubeconfig"
	"opendev.org/airship/airshipctl/pkg/k8s/utils"
	"opendev.org/airship/airshipctl/pkg/log"
	"opendev.org/airship/airshipctl/pkg/phase/ifc"
)

// ExecutorOptions provide a way to configure executor
type ExecutorOptions struct {
	BundleName  string
	ClusterName string

	ExecutorDocument document.Document
	BundleFactory    document.BundleFactoryFunc
	Kubeconfig       kubeconfig.Interface
	Helper           ifc.Helper
	ClusterMap       clustermap.ClusterMap
}

var _ ifc.Executor = &Executor{}

// RegisterExecutor adds executor to phase executor registry
func RegisterExecutor(registry map[schema.GroupVersionKind]ifc.ExecutorFactory) error {
	obj := &airshipv1.KubernetesApply{}
	gvks, _, err := airshipv1.Scheme.ObjectKinds(obj)
	if err != nil {
		return err
	}
	registry[gvks[0]] = registerExecutor
	return nil
}

// registerExecutor is here so that executor in theory can be used outside phases
func registerExecutor(cfg ifc.ExecutorConfig) (ifc.Executor, error) {
	return NewExecutor(ExecutorOptions{
		ClusterName:      cfg.ClusterName,
		BundleName:       cfg.PhaseName,
		Helper:           cfg.Helper,
		ExecutorDocument: cfg.ExecutorDocument,
		BundleFactory:    cfg.BundleFactory,
		Kubeconfig:       cfg.KubeConfig,
		ClusterMap:       cfg.ClusterMap,
	})
}

// Executor applies resources to kubernetes
type Executor struct {
	Options        ExecutorOptions
	ExecutorBundle document.Bundle

	apiObject *airshipv1.KubernetesApply
	cleanup   kubeconfig.Cleanup
}

// NewExecutor returns instance of executor
func NewExecutor(opts ExecutorOptions) (*Executor, error) {
	apiObj := &airshipv1.KubernetesApply{}
	err := opts.ExecutorDocument.ToAPIObject(apiObj, airshipv1.Scheme)
	if err != nil {
		return nil, err
	}
	bundle, err := opts.BundleFactory()
	if err != nil {
		return nil, err
	}
	return &Executor{
		ExecutorBundle: bundle,
		Options:        opts,
		apiObject:      apiObj,
	}, nil
}

// Run executor, should be performed in separate go routine
func (e *Executor) Run(ch chan events.Event, runOpts ifc.RunOptions) {
	applier, filteredBundle, err := e.prepareApplier(ch)
	if err != nil {
		handleError(ch, err)
		close(ch)
		return
	}
	defer e.cleanup()
	dryRunStrategy := common.DryRunNone
	if runOpts.DryRun {
		dryRunStrategy = common.DryRunClient
	}
	timeout := time.Second * time.Duration(e.apiObject.Config.WaitOptions.Timeout)
	if int64(runOpts.Timeout/time.Second) != 0 {
		timeout = runOpts.Timeout
	}

	log.Debugf("WaitTimeout: %v", timeout)
	applyOptions := ApplyOptions{
		DryRunStrategy: dryRunStrategy,
		Prune:          e.apiObject.Config.PruneOptions.Prune,
		BundleName:     e.Options.BundleName,
		WaitTimeout:    timeout,
	}
	applier.ApplyBundle(filteredBundle, applyOptions)
}

func (e *Executor) prepareApplier(ch chan events.Event) (*Applier, document.Bundle, error) {
	log.Debug("Getting kubeconfig context name from cluster map")
	context, err := e.Options.ClusterMap.ClusterKubeconfigContext(e.Options.ClusterName)
	if err != nil {
		return nil, nil, err
	}
	log.Debug("Getting kubeconfig file information from kubeconfig provider")
	path, cleanup, err := e.Options.Kubeconfig.GetFile()
	if err != nil {
		return nil, nil, err
	}
	log.Debug("Filtering out documents that shouldn't be applied to kubernetes from document bundle")
	bundle, err := e.ExecutorBundle.SelectBundle(document.NewDeployToK8sSelector())
	if err != nil {
		cleanup()
		return nil, nil, err
	}
	// set up cleanup only if all calls up to here were successful
	e.cleanup = cleanup
	log.Debugf("Using kubeconfig at '%s' and context '%s'", path, context)
	factory := utils.FactoryFromKubeConfig(path, context)
	return NewApplier(ch, factory), bundle, nil
}

// Validate document set
func (e *Executor) Validate() error {
	return errors.ErrNotImplemented{}
}

// Render document set
func (e *Executor) Render(w io.Writer, o ifc.RenderOptions) error {
	bundle, err := e.ExecutorBundle.SelectBundle(o.FilterSelector)
	if err != nil {
		return err
	}
	return bundle.Write(w)
}
