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
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	clusterctlclient "sigs.k8s.io/cluster-api/cmd/clusterctl/client"
	clusterctlconfig "sigs.k8s.io/cluster-api/cmd/clusterctl/client/config"
	clog "sigs.k8s.io/cluster-api/cmd/clusterctl/log"
	"sigs.k8s.io/yaml"

	airshipv1 "opendev.org/airship/airshipctl/pkg/clusterctl/api/v1alpha1"
	"opendev.org/airship/airshipctl/pkg/log"
)

const (
	// path to file on in memory file system
	confFilePath       = "/air-clusterctl.yaml"
	dummyComponentPath = "/dummy/path/v0.3.2/components.yaml"
)

var _ Interface = &Client{}

// Interface is abstraction to Clusterctl
type Interface interface {
	Init(kubeconfigPath string) error
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
func (c *Client) Init(kubeconfigPath string) error {
	log.Print("Starting cluster-api initiation")
	c.initOptions.Kubeconfig = kubeconfigPath
	_, err := c.clusterctlClient.Init(c.initOptions)
	return err
}

// newConfig returns clusterctl config client
func newConfig(options *airshipv1.Clusterctl) (clusterctlconfig.Client, error) {
	fs := afero.NewMemMapFs()
	b := []map[string]string{}
	for _, provider := range options.Providers {
		p := map[string]string{
			"name": provider.Name,
			"type": provider.Type,
			"url":  provider.URL,
		}
		// this is a workaround as cluserctl validates if URL is empty, even though it is not
		// used anywhere outside repository factory which we override
		// TODO (kkalynovskyi) we need to create issue for this in clusterctl, and remove URL
		// validation and move it to be an error during repository interface initialization
		if !provider.IsClusterctlRepository {
			p["url"] = dummyComponentPath
		}
		b = append(b, p)
	}
	cconf := map[string][]map[string]string{
		"providers": b,
	}
	data, err := yaml.Marshal(cconf)
	if err != nil {
		return nil, err
	}
	err = afero.WriteFile(fs, confFilePath, data, 0600)
	if err != nil {
		return nil, err
	}
	// Set filesystem to global viper object, to make sure, that clusterctl config is read from
	// memory filesystem instead of real one.
	viper.SetFs(fs)
	return clusterctlconfig.New(confFilePath)
}

func newClusterctlClient(root string, options *airshipv1.Clusterctl) (clusterctlclient.Client, error) {
	cconf, err := newConfig(options)
	if err != nil {
		return nil, err
	}
	rf := RepositoryFactory{
		root:         root,
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
