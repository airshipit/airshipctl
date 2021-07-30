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
	"io"
	"time"

	"sigs.k8s.io/cli-utils/pkg/common"
	"sigs.k8s.io/cli-utils/pkg/kstatus/polling/aggregator"
	"sigs.k8s.io/cli-utils/pkg/kstatus/polling/event"
	"sigs.k8s.io/cli-utils/pkg/kstatus/status"
	"sigs.k8s.io/cli-utils/pkg/provider"

	airshipv1 "opendev.org/airship/airshipctl/pkg/api/v1alpha1"
	"opendev.org/airship/airshipctl/pkg/cluster/clustermap"
	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/events"
	k8sapplier "opendev.org/airship/airshipctl/pkg/k8s/applier"
	"opendev.org/airship/airshipctl/pkg/k8s/kubeconfig"
	"opendev.org/airship/airshipctl/pkg/k8s/utils"
	"opendev.org/airship/airshipctl/pkg/log"
	"opendev.org/airship/airshipctl/pkg/phase/errors"
	"opendev.org/airship/airshipctl/pkg/phase/ifc"
)

var _ ifc.Executor = &KubeApplierExecutor{}

// KubeApplierExecutor applies resources to kubernetes
type KubeApplierExecutor struct {
	ExecutorBundle   document.Bundle
	ExecutorDocument document.Document
	BundleName       string

	apiObject   *airshipv1.KubernetesApply
	cleanup     kubeconfig.Cleanup
	clusterMap  clustermap.ClusterMap
	clusterName string
	kubeconfig  kubeconfig.Interface
}

// NewKubeApplierExecutor returns instance of executor
func NewKubeApplierExecutor(cfg ifc.ExecutorConfig) (ifc.Executor, error) {
	apiObj := &airshipv1.KubernetesApply{}
	err := cfg.ExecutorDocument.ToAPIObject(apiObj, airshipv1.Scheme)
	if err != nil {
		return nil, err
	}
	bundle, err := cfg.BundleFactory()
	if err != nil {
		return nil, err
	}
	return &KubeApplierExecutor{
		ExecutorBundle:   bundle,
		BundleName:       cfg.PhaseName,
		ExecutorDocument: cfg.ExecutorDocument,
		apiObject:        apiObj,
		clusterMap:       cfg.ClusterMap,
		clusterName:      cfg.ClusterName,
		kubeconfig:       cfg.KubeConfig,
		// default cleanup that does nothing
		// replaced with a meaningful cleanup while preparing kubeconfig
		cleanup: func() {},
	}, nil
}

// Run executor, should be performed in separate go routine
func (e *KubeApplierExecutor) Run(ch chan events.Event, runOpts ifc.RunOptions) {
	defer close(ch)

	applier, filteredBundle, err := e.prepareApplier(ch)
	if err != nil {
		handleError(ch, err)
		return
	}
	defer e.cleanup()

	dryRunStrategy := common.DryRunNone
	if runOpts.DryRun {
		dryRunStrategy = common.DryRunClient
	}
	timeout := time.Second * time.Duration(e.apiObject.Config.WaitOptions.Timeout)
	if runOpts.Timeout != nil {
		timeout = *runOpts.Timeout
	}

	log.Debugf("WaitTimeout: %v", timeout)
	applyOptions := k8sapplier.ApplyOptions{
		DryRunStrategy: dryRunStrategy,
		Prune:          e.apiObject.Config.PruneOptions.Prune,
		BundleName:     e.BundleName,
		WaitTimeout:    timeout,
	}
	applier.ApplyBundle(filteredBundle, applyOptions)
}

func (e *KubeApplierExecutor) prepareApplier(ch chan events.Event) (*k8sapplier.Applier, document.Bundle, error) {
	log.Debug("Filtering out documents that shouldn't be applied to kubernetes from document bundle")
	bundle, err := e.ExecutorBundle.SelectBundle(document.NewDeployToK8sSelector())
	if err != nil {
		return nil, nil, err
	}
	log.Debug("Getting kubeconfig context name from cluster map")
	context, err := e.clusterMap.ClusterKubeconfigContext(e.clusterName)
	if err != nil {
		return nil, nil, err
	}
	log.Debug("Getting kubeconfig file information from kubeconfig provider")
	path, cleanup, err := e.kubeconfig.GetFile()
	if err != nil {
		return nil, nil, err
	}
	// set up cleanup only if all calls up to here were successful
	e.cleanup = cleanup
	log.Printf("Using kubeconfig at '%s' and context '%s'", path, context)
	factory := utils.FactoryFromKubeConfig(path, context)
	return k8sapplier.NewApplier(ch, factory), bundle, nil
}

// Validate document set
func (e *KubeApplierExecutor) Validate() error {
	if e.BundleName == "" {
		return errors.ErrInvalidPhase{Reason: "k8s applier BundleName is empty"}
	}
	docs, err := e.ExecutorBundle.GetAllDocuments()
	if err != nil {
		return err
	}
	if len(docs) == 0 {
		return errors.ErrInvalidPhase{Reason: "no executor documents in the bundle"}
	}
	// TODO: need to find if any other validation needs to be added
	return nil
}

// Render document set
func (e *KubeApplierExecutor) Render(w io.Writer, o ifc.RenderOptions) error {
	bundle, err := e.ExecutorBundle.SelectBundle(o.FilterSelector)
	if err != nil {
		return err
	}
	return bundle.Write(w)
}

// Status returns the status of the given phase
func (e *KubeApplierExecutor) Status() (sts ifc.ExecutorStatus, err error) {
	var ctx string
	ctx, err = e.clusterMap.ClusterKubeconfigContext(e.clusterName)
	if err != nil {
		return sts, err
	}
	log.Debug("Getting kubeconfig file information from kubeconfig provider")
	path, cleanup, err := e.kubeconfig.GetFile()
	if err != nil {
		return sts, err
	}
	defer cleanup()

	cf := provider.NewProvider(utils.FactoryFromKubeConfig(path, ctx))
	rm, err := cf.Factory().ToRESTMapper()
	if err != nil {
		return
	}
	r := utils.DefaultManifestReaderFactory(false, e.ExecutorBundle, rm)
	infos, err := r.Read()
	if err != nil {
		return
	}

	var resSts event.ResourceStatuses

	for _, info := range infos {
		s, sErr := status.Compute(info)
		if sErr != nil {
			return
		}
		st := &event.ResourceStatus{
			Status: s.Status,
		}
		resSts = append(resSts, st)
	}
	_ = aggregator.AggregateStatus(resSts, status.CurrentStatus)
	return ifc.ExecutorStatus{}, err
}
