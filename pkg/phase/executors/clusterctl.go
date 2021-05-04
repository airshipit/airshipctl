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
	"fmt"
	"io"
	"strings"

	airshipv1 "opendev.org/airship/airshipctl/pkg/api/v1alpha1"
	"opendev.org/airship/airshipctl/pkg/cluster/clustermap"
	"opendev.org/airship/airshipctl/pkg/clusterctl/client"
	"opendev.org/airship/airshipctl/pkg/document"
	airerrors "opendev.org/airship/airshipctl/pkg/errors"
	"opendev.org/airship/airshipctl/pkg/events"
	"opendev.org/airship/airshipctl/pkg/k8s/kubeconfig"
	"opendev.org/airship/airshipctl/pkg/log"
	phaseerrors "opendev.org/airship/airshipctl/pkg/phase/errors"
	"opendev.org/airship/airshipctl/pkg/phase/executors/errors"
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
	client, err := client.NewClient(cfg.TargetPath, log.DebugEnabled(), options)
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
		handleError(evtCh, errors.ErrUnknownExecutorAction{Action: string(c.options.Action), ExecutorName: "clusterctl"})
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

	eventMsg := "clusterctl init completed successfully"

	// Use cluster name as context in kubeconfig file
	err = c.Init(kubeConfigFile, context)
	if err != nil && isAlreadyExistsError(err) {
		// log the already existed/initialized error as warning and continue
		eventMsg = fmt.Sprintf("WARNING: clusterctl is already initialized, received an error :  %s", err.Error())
	} else if err != nil {
		handleError(evtCh, err)
		return
	}
	evtCh <- events.NewEvent().WithClusterctlEvent(events.ClusterctlEvent{
		Operation: events.ClusterctlInitEnd,
		Message:   eventMsg,
	})
}

func isAlreadyExistsError(err error) bool {
	return strings.Contains(err.Error(), "there is already an instance")
}

// Validate executor configuration and documents
func (c *ClusterctlExecutor) Validate() error {
	switch c.options.Action {
	case "":
		return phaseerrors.ErrInvalidPhase{Reason: "ClusterctlExecutor.Action is empty"}
	case airshipv1.Init:
		if c.options.InitOptions.CoreProvider == "" {
			return phaseerrors.ErrInvalidPhase{Reason: "ClusterctlExecutor.InitOptions.CoreProvider is empty"}
		}
	case airshipv1.Move:
	default:
		return errors.ErrUnknownExecutorAction{Action: string(c.options.Action)}
	}
	// TODO: need to find if any other validation needs to be added
	return nil
}

// Render executor documents
func (c *ClusterctlExecutor) Render(w io.Writer, ro ifc.RenderOptions) error {
	dataAll := bytes.NewBuffer([]byte{})
	typeMap := map[string][]string{
		string(client.BootstrapProviderType):      c.options.InitOptions.BootstrapProviders,
		string(client.ControlPlaneProviderType):   c.options.InitOptions.ControlPlaneProviders,
		string(client.InfrastructureProviderType): c.options.InitOptions.InfrastructureProviders,
		string(client.CoreProviderType): (map[bool][]string{true: {c.options.InitOptions.CoreProvider},
			false: {}})[c.options.InitOptions.CoreProvider != ""],
	}
	for prvType, prvList := range typeMap {
		for _, prv := range prvList {
			res := strings.Split(prv, ":")
			if len(res) != 2 {
				return errors.ErrUnableParseProvider{
					Provider:     prv,
					ProviderType: prvType,
				}
			}
			data, err := c.Interface.Render(client.RenderOptions{
				ProviderName:    res[0],
				ProviderVersion: res[1],
				ProviderType:    prvType,
			})
			if err != nil {
				return err
			}
			dataAll.Write(data)
			dataAll.Write([]byte("\n---\n"))
		}
	}

	bundle, err := document.NewBundleFromBytes(dataAll.Bytes())
	if err != nil {
		return err
	}
	filtered, err := bundle.SelectBundle(ro.FilterSelector)
	if err != nil {
		return err
	}
	return filtered.Write(w)
}

// Status returns the status of the given phase
func (c *ClusterctlExecutor) Status() (ifc.ExecutorStatus, error) {
	return ifc.ExecutorStatus{}, airerrors.ErrNotImplemented{What: Clusterctl}
}
