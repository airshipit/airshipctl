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

package cmd

import (
	airshipv1 "opendev.org/airship/airshipctl/pkg/api/v1alpha1"
	"opendev.org/airship/airshipctl/pkg/clusterctl/client"
	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/environment"
)

// Command adds a layer to clusterctl interface with airshipctl context
type Command struct {
	kubeconfigPath    string
	kubeconfigContext string
	documentRoot      string
	client            client.Interface
	options           *airshipv1.Clusterctl
}

// NewCommand returns instance of Command
func NewCommand(rs *environment.AirshipCTLSettings) (*Command, error) {
	bundle, err := getBundle(rs.Config)
	if err != nil {
		return nil, err
	}
	if err = rs.Config.EnsureComplete(); err != nil {
		return nil, err
	}
	root, err := rs.Config.CurrentContextTargetPath()
	if err != nil {
		return nil, err
	}
	options, err := clusterctlOptions(bundle)
	if err != nil {
		return nil, err
	}
	client, err := client.NewClient(root, rs.Debug, options)
	if err != nil {
		return nil, err
	}
	kubeConfigPath := rs.Config.KubeConfigPath()

	return &Command{
		kubeconfigPath:    kubeConfigPath,
		documentRoot:      root,
		client:            client,
		options:           options,
		kubeconfigContext: rs.Config.CurrentContext,
	}, nil
}

// Init runs clusterctl init
func (c *Command) Init() error {
	return c.client.Init(c.kubeconfigPath, c.kubeconfigContext)
}

func clusterctlOptions(bundle document.Bundle) (*airshipv1.Clusterctl, error) {
	cctl := &airshipv1.Clusterctl{}
	selector, err := document.NewSelector().ByObject(cctl, airshipv1.Scheme)
	if err != nil {
		return nil, err
	}

	doc, err := bundle.SelectOne(selector)
	if err != nil {
		return nil, err
	}

	if err := doc.ToAPIObject(cctl, airshipv1.Scheme); err != nil {
		return nil, err
	}

	return cctl, nil
}

func getBundle(conf *config.Config) (document.Bundle, error) {
	path, err := conf.CurrentContextEntryPoint(config.ClusterctlPhase)
	if err != nil {
		return nil, err
	}
	return document.NewBundleByPath(path)
}

// Move runs clusterctl move
func (c *Command) Move(toKubeconfigContext string) error {
	if c.options.MoveOptions != nil {
		return c.client.Move(c.kubeconfigPath, c.kubeconfigContext,
			c.kubeconfigPath, toKubeconfigContext, c.options.MoveOptions.Namespace)
	}
	return c.client.Move(c.kubeconfigPath, c.kubeconfigContext, c.kubeconfigPath, toKubeconfigContext, "")
}
