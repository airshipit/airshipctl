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
	clusterctlclient "sigs.k8s.io/cluster-api/cmd/clusterctl/client"
	clusterctlconfig "sigs.k8s.io/cluster-api/cmd/clusterctl/client/config"
	clog "sigs.k8s.io/cluster-api/cmd/clusterctl/log"

	airshipv1 "opendev.org/airship/airshipctl/pkg/clusterctl/api/v1alpha1"
	"opendev.org/airship/airshipctl/pkg/clusterctl/implementations"
	"opendev.org/airship/airshipctl/pkg/log"
)

var _ Interface = &Client{}

// Interface is abstraction to Clusterctl
type Interface interface {
	Init(kubeconfigPath, kubeconfigContext string) error
}

// Client Implements interface to Clusterctl
type Client struct {
	clusterctlClient clusterctlclient.Client
	initOptions      clusterctlclient.InitOptions
}

// NewClient returns instance of clusterctl client
func NewClient(root string, debug bool, options *airshipv1.Clusterctl) (Interface, error) {
	if debug {
		debugVerbosity := 5
		clog.SetLogger(clog.NewLogger(clog.WithThreshold(&debugVerbosity)))
	}
	initOptions := options.InitOptions
	var cio clusterctlclient.InitOptions
	if initOptions != nil {
		cio = clusterctlclient.InitOptions{
			BootstrapProviders:      initOptions.BootstrapProviders,
			CoreProvider:            initOptions.CoreProvider,
			InfrastructureProviders: initOptions.InfrastructureProviders,
			ControlPlaneProviders:   initOptions.ControlPlaneProviders,
		}
	}
	cclient, err := newClusterctlClient(root, options)
	if err != nil {
		return nil, err
	}
	return &Client{clusterctlClient: cclient, initOptions: cio}, nil
}

// Init implements interface to Clusterctl
func (c *Client) Init(kubeconfigPath, kubeconfigContext string) error {
	log.Print("Starting cluster-api initiation")
	c.initOptions.Kubeconfig = clusterctlclient.Kubeconfig{
		Path:    kubeconfigPath,
		Context: kubeconfigContext,
	}
	_, err := c.clusterctlClient.Init(c.initOptions)
	return err
}

// newConfig returns clusterctl config client
func newConfig(options *airshipv1.Clusterctl, root string) (clusterctlconfig.Client, error) {
	for _, provider := range options.Providers {
		// this is a workaround as cluserctl validates if URL is empty, even though it is not
		// used anywhere outside repository factory which we override
		// TODO (kkalynovskyi) we need to create issue for this in clusterctl, and remove URL
		// validation and move it to be an error during repository interface initialization
		if !provider.IsClusterctlRepository {
			provider.URL = root
		}
	}
	reader, err := implementations.NewAirshipReader(options)
	if err != nil {
		return nil, err
	}
	return clusterctlconfig.New("", clusterctlconfig.InjectReader(reader))
}

func newClusterctlClient(root string, options *airshipv1.Clusterctl) (clusterctlclient.Client, error) {
	cconf, err := newConfig(options, root)
	if err != nil {
		return nil, err
	}
	rf := RepositoryFactory{
		Options:      options,
		ConfigClient: cconf,
	}
	// option config factory
	ocf := clusterctlclient.InjectConfig(cconf)
	// option repository factory
	orf := clusterctlclient.InjectRepositoryFactory(rf.ClientRepositoryFactory())
	// options cluster client factory
	occf := clusterctlclient.InjectClusterClientFactory(rf.ClusterClientFactory())
	return clusterctlclient.New("", ocf, orf, occf)
}
