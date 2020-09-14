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

package client

import (
	"io"

	"k8s.io/apimachinery/pkg/runtime/schema"

	airshipv1 "opendev.org/airship/airshipctl/pkg/api/v1alpha1"
	"opendev.org/airship/airshipctl/pkg/document"
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

	Interface
	bundle  document.Bundle
	options *airshipv1.Clusterctl
	kubecfg kubeconfig.Interface
}

// RegisterExecutor adds executor to phase executor registry
func RegisterExecutor(registry map[schema.GroupVersionKind]ifc.ExecutorFactory) error {
	obj := &airshipv1.Clusterctl{}
	gvks, _, err := airshipv1.Scheme.ObjectKinds(obj)
	if err != nil {
		return err
	}
	registry[gvks[0]] = NewExecutor
	return nil
}

// NewExecutor creates instance of 'clusterctl init' phase executor
func NewExecutor(cfg ifc.ExecutorConfig) (ifc.Executor, error) {
	options := &airshipv1.Clusterctl{}
	if err := cfg.ExecutorDocument.ToAPIObject(options, airshipv1.Scheme); err != nil {
		return nil, err
	}
	client, err := NewClient(cfg.Helper.TargetPath(), log.DebugEnabled(), options)
	if err != nil {
		return nil, err
	}
	return &ClusterctlExecutor{
		clusterName: cfg.ClusterName,
		Interface:   client,
		bundle:      cfg.ExecutorBundle,
		options:     options,
		kubecfg:     cfg.KubeConfig,
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
		c.handleErr(ErrUnknownExecutorAction{Action: string(c.options.Action)}, evtCh)
	}
}

func (c *ClusterctlExecutor) move(_ ifc.RunOptions, evtCh chan events.Event) {
	c.handleErr(errors.ErrNotImplemented{}, evtCh)
}

func (c *ClusterctlExecutor) init(opts ifc.RunOptions, evtCh chan events.Event) {
	evtCh <- events.Event{
		ClusterctlEvent: events.ClusterctlEvent{
			Operation: events.ClusterctlInitStart,
		},
	}
	kubeConfigFile, cleanup, err := c.kubecfg.GetFile()
	if err != nil {
		c.handleErr(err, evtCh)
		return
	}

	defer cleanup()

	if opts.DryRun {
		// TODO (dukov) add more details to dry-run
		log.Print("command 'clusterctl init' is going to be executed")
		evtCh <- events.Event{
			ClusterctlEvent: events.ClusterctlEvent{
				Operation: events.ClusterctlInitEnd,
			},
		}
		return
	}
	err = c.Init(kubeConfigFile, "")
	if err != nil {
		c.handleErr(err, evtCh)
	}
	evtCh <- events.Event{
		ClusterctlEvent: events.ClusterctlEvent{
			Operation: events.ClusterctlInitEnd,
		},
	}
}

func (c *ClusterctlExecutor) handleErr(err error, evtCh chan events.Event) {
	evtCh <- events.Event{
		Type: events.ErrorType,
		ErrorEvent: events.ErrorEvent{
			Error: err,
		},
	}
}

// Validate executor configuration and documents
func (c *ClusterctlExecutor) Validate() error {
	return errors.ErrNotImplemented{}
}

// Render executor documents
func (c *ClusterctlExecutor) Render(_ io.Writer, _ ifc.RenderOptions) error {
	return errors.ErrNotImplemented{}
}
