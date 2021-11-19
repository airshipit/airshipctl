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

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	v1 "k8s.io/client-go/applyconfigurations/core/v1"
	"k8s.io/klog/v2"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"sigs.k8s.io/cli-utils/cmd/flagutils"
	"sigs.k8s.io/cli-utils/cmd/printers"
	"sigs.k8s.io/cli-utils/pkg/apply"
	"sigs.k8s.io/cli-utils/pkg/common"
	"sigs.k8s.io/cli-utils/pkg/errors"
	"sigs.k8s.io/cli-utils/pkg/inventory"
	"sigs.k8s.io/cli-utils/pkg/util/factory"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"

	"opendev.org/airship/airshipctl/krm-functions/applier/image/poller"
	"opendev.org/airship/airshipctl/krm-functions/applier/image/types"
)

const (
	airshipNamespace = "airshipit"
)

// Config is an extension of ApplyConfig struct with added streams
type Config struct {
	*types.ApplyConfig
	Streams      genericclioptions.IOStreams
}

func factoryFromKubeConfig(path, context string) cmdutil.Factory {
	kf := genericclioptions.NewConfigFlags(true).WithDeprecatedPasswordFlag()
	kf.KubeConfig = &path
	kf.Context = &context
	return cmdutil.NewFactory(cmdutil.NewMatchVersionFlags(&factory.CachingRESTClientGetter{Delegate: kf}))
}

func appendInventoryInfo(obj []*unstructured.Unstructured, name string) (*unstructured.Unstructured, []*unstructured.Unstructured, error) {
	namespace := fmt.Sprintf("%s-%s", airshipNamespace, name)
	cmObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(v1.ConfigMap(fmt.Sprintf("inventory-%s",
		common.RandomStr()), namespace).WithLabels(map[string]string{common.InventoryLabel: name}))
	if err != nil {
		return nil, nil, err
	}
	namespaceObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(v1.Namespace(namespace))
	if err != nil {
		return nil, nil, err
	}

	return &unstructured.Unstructured{Object: cmObj}, append(obj, &unstructured.Unstructured{Object: namespaceObj}), nil
}

// Run prepares config, applier and performs apply process
func (c *Config) Run(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
	if c.Debug {
		if err := flag.Set("v", "2"); err != nil {
			klog.V(2).Infof("unable to set debug flag: %v\n", err)
		}
	}

	var objs []*unstructured.Unstructured
	for _, node := range nodes {
		m, err := node.Map()
		if err != nil {
			return nil, err
		}
		objs = append(objs, &unstructured.Unstructured{Object: m})
	}

	f := factoryFromKubeConfig(c.Kubeconfig, c.Context)
	statusPoller, err := poller.NewStatusPoller(f, c.WaitOptions.Conditions...)
	if err != nil {
		return nil, err
	}

	invFactory := inventory.ClusterInventoryClientFactory{}
	invClient, err := invFactory.NewInventoryClient(f)
	if err != nil {
		return nil, err
	}

	applier, err := apply.NewApplier(f, invClient, statusPoller)
	inv, obj, err := inventory.SplitUnstructureds(objs)
	if err != nil {
		klog.V(2).Infoln("injecting auto generated inventory object")
		inv, obj, err = appendInventoryInfo(obj, c.PhaseName)
		if err != nil {
			return nil, err
		}
	}

	opts := c.toCliOptions()
	err = printers.GetPrinter(printers.DefaultPrinter(), c.Streams).Print(
		applier.Run(context.Background(), inventory.WrapInventoryInfoObj(inv), obj, opts), opts.DryRunStrategy, true)
	klog.V(2).Infoln("applier channel closed")
	errors.CheckErr(c.Streams.ErrOut, err, "applier")
	return nil, err
}

func (c *Config) toCliOptions() apply.Options {
	dryRunStrategy := common.DryRunNone
	if c.DryRun {
		dryRunStrategy = common.DryRunClient
	}
	timeout := time.Second * time.Duration(c.WaitOptions.Timeout)
	pollInterval := time.Second * time.Duration(c.WaitOptions.PollInterval)

	var emitStatusEvents bool
	// if wait timeout is 0, we don't want to status poller to emit any events,
	// this should disable waiting for resources
	if timeout != time.Duration(0) {
		emitStatusEvents = true
	}

	inventoryPolicy, err := flagutils.ConvertInventoryPolicy(c.InventoryPolicy)
	if err != nil {
		klog.V(2).Infof("%s or force-adopt, using the default one (strict)", err.Error())
	}

	return apply.Options{
		DryRunStrategy:   dryRunStrategy,
		NoPrune:          !c.PruneOptions.Prune,
		EmitStatusEvents: emitStatusEvents,
		ReconcileTimeout: timeout,
		PollInterval:     pollInterval,
		InventoryPolicy:  inventoryPolicy,
	}
}

func main() {
	cfg := &Config{ApplyConfig: &types.ApplyConfig{}}
	cmd := command.Build(framework.SimpleProcessor{Filter: kio.FilterFunc(cfg.Run), Config: cfg.ApplyConfig},
		command.StandaloneDisabled, false)
	cfg.Streams = genericclioptions.IOStreams{In: cmd.InOrStdin(), Out: cmd.ErrOrStderr(), ErrOut: cmd.ErrOrStderr()}

	klog.InitFlags(nil)
	klog.SetOutput(cmd.ErrOrStderr())
	defer klog.Flush()

	if err := cmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr())
		os.Exit(1)
	}
}
