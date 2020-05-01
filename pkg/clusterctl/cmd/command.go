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
	"sigs.k8s.io/yaml"

	airshipv1 "opendev.org/airship/airshipctl/pkg/clusterctl/api/v1alpha1"
	"opendev.org/airship/airshipctl/pkg/clusterctl/client"
	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/environment"
)

// Command adds a layer to clusterctl interface with airshipctl context
type Command struct {
	kubeconfigPath string
	documentRoot   string
	client         client.Interface
	options        *airshipv1.Clusterctl
}

// NewCommand returns instance of Command
func NewCommand(rs *environment.AirshipCTLSettings) (*Command, error) {
	bundle, err := getBundle(rs.Config)
	if err != nil {
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
		kubeconfigPath: kubeConfigPath,
		documentRoot:   root,
		client:         client,
		options:        options,
	}, nil
}

// Init runs clusterctl init
func (c *Command) Init() error {
	return c.client.Init(c.kubeconfigPath)
}

func clusterctlOptions(bundle document.Bundle) (*airshipv1.Clusterctl, error) {
	doc, err := bundle.SelectOne(document.NewClusterctlSelector())
	if err != nil {
		return nil, err
	}
	options := &airshipv1.Clusterctl{}
	b, err := doc.AsYAML()
	if err != nil {
		return nil, err
	}
	// TODO (kkalynovskyi) instead of this, use kubernetes serializer
	err = yaml.Unmarshal(b, options)
	if err != nil {
		return nil, err
	}
	return options, nil
}

func getBundle(conf *config.Config) (document.Bundle, error) {
	path, err := conf.CurrentContextEntryPoint(config.ClusterctlPhase)
	if err != nil {
		return nil, err
	}
	return document.NewBundleByPath(path)
}
