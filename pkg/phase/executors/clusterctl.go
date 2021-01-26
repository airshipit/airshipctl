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

	airshipv1 "opendev.org/airship/airshipctl/pkg/api/v1alpha1"
	"opendev.org/airship/airshipctl/pkg/cluster/clustermap"
	"opendev.org/airship/airshipctl/pkg/clusterctl/client"
	"opendev.org/airship/airshipctl/pkg/errors"
	"opendev.org/airship/airshipctl/pkg/events"
	"opendev.org/airship/airshipctl/pkg/k8s/kubeconfig"
	"opendev.org/airship/airshipctl/pkg/log"
	"opendev.org/airship/airshipctl/pkg/phase/ifc"
)

var _ ifc.Executor = &ClusterctlExecutor{}

// ClusterctlExecutor phase executor
type ClusterctlExecutor struct {
	clusterName string

	client.Interface
	clusterMap clustermap.ClusterMap
	options    *airshipv1.Clusterctl
	kubecfg    kubeconfig.Interface
}

// NewClusterctlExecutor creates instance of 'clusterctl init' phase executor
func NewClusterctlExecutor(cfg ifc.ExecutorConfig) (ifc.Executor, error) {
	options := airshipv1.DefaultClusterctl()
	if err := cfg.ExecutorDocument.ToAPIObject(options, airshipv1.Scheme); err != nil {
		return nil, err
	}
	client, err := client.NewClient(cfg.Helper.TargetPath(), log.DebugEnabled(), options)
	if err != nil {
		return nil, err
	}
	return &ClusterctlExecutor{
		clusterName: cfg.ClusterName,
		Interface:   client,
		options:     options,
		kubecfg:     cfg.KubeConfig,
		clusterMap:  cfg.ClusterMap,
	}, nil
}

// Run clusterctl init as a phase runner
func (c *ClusterctlExecutor) Run(evtCh chan events.Event, opts ifc.RunOptions) {
	defer close(evtCh)
	switch c.options.Action {
	case airshipv1.Move:
		c.move(opts, evtCh)
	case airshipv1.Init:
		c.init(opts, evtCh)
	default:
		handleError(evtCh, ErrUnknownExecutorAction{Action: string(c.options.Action), ExecutorName: "clusterctl"})
	}
}

func (c *ClusterctlExecutor) move(opts ifc.RunOptions, evtCh chan events.Event) {
	evtCh <- events.NewEvent().WithClusterctlEvent(events.ClusterctlEvent{
		Operation: events.ClusterctlMoveStart,
		Message:   "starting clusterctl move executor",
	})
	ns := c.options.MoveOptions.Namespace
	kubeConfigFile, cleanup, err := c.kubecfg.GetFile()
	if err != nil {
		handleError(evtCh, err)
		return
	}
	defer cleanup()
	fromCluster, err := c.clusterMap.ParentCluster(c.clusterName)
	if err != nil {
		handleError(evtCh, err)
		return
	}
	fromContext, err := c.clusterMap.ClusterKubeconfigContext(fromCluster)
	if err != nil {
		handleError(evtCh, err)
		return
	}
	toContext, err := c.clusterMap.ClusterKubeconfigContext(c.clusterName)
	if err != nil {
		handleError(evtCh, err)
		return
	}

	log.Print("command 'clusterctl move' is going to be executed")
	// TODO (kkalynovskyi) add more details to dry-run, for now if dry run is set we skip move command
	if !opts.DryRun {
		err = c.Move(kubeConfigFile, fromContext, kubeConfigFile, toContext, ns)
		if err != nil {
			handleError(evtCh, err)
			return
		}
	}

	evtCh <- events.NewEvent().WithClusterctlEvent(events.ClusterctlEvent{
		Operation: events.ClusterctlMoveEnd,
		Message:   "clusterctl move completed successfully",
	})
}

func (c *ClusterctlExecutor) init(opts ifc.RunOptions, evtCh chan events.Event) {
	evtCh <- events.NewEvent().WithClusterctlEvent(events.ClusterctlEvent{
		Operation: events.ClusterctlInitStart,
		Message:   "starting clusterctl init executor",
	})
	kubeConfigFile, cleanup, err := c.kubecfg.GetFile()
	if err != nil {
		handleError(evtCh, err)
		return
	}

	defer cleanup()

	if opts.DryRun {
		// TODO (dukov) add more details to dry-run
		log.Print("command 'clusterctl init' is going to be executed")
		evtCh <- events.NewEvent().WithClusterctlEvent(events.ClusterctlEvent{
			Operation: events.ClusterctlInitEnd,
			Message:   "clusterctl init dry-run completed successfully",
		})
		return
	}

	context, err := c.clusterMap.ClusterKubeconfigContext(c.clusterName)
	if err != nil {
		handleError(evtCh, err)
		return
	}

	// Use cluster name as context in kubeconfig file
	err = c.Init(kubeConfigFile, context)
	if err != nil {
		handleError(evtCh, err)
		return
	}
	evtCh <- events.NewEvent().WithClusterctlEvent(events.ClusterctlEvent{
		Operation: events.ClusterctlInitEnd,
		Message:   "clusterctl init completed successfully",
	})
}

// Validate executor configuration and documents
func (c *ClusterctlExecutor) Validate() error {
	return errors.ErrNotImplemented{}
}

// Render executor documents
func (c *ClusterctlExecutor) Render(w io.Writer, _ ifc.RenderOptions) error {
	// will be implemented later
	_, err := w.Write([]byte{})
	return err
}
