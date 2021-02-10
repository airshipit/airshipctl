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
	"sigs.k8s.io/cluster-api/cmd/clusterctl/api/v1alpha3"
	clusterctlclient "sigs.k8s.io/cluster-api/cmd/clusterctl/client"
	clusterctlconfig "sigs.k8s.io/cluster-api/cmd/clusterctl/client/config"
	"sigs.k8s.io/cluster-api/cmd/clusterctl/client/repository"
	"sigs.k8s.io/cluster-api/cmd/clusterctl/client/yamlprocessor"
	clog "sigs.k8s.io/cluster-api/cmd/clusterctl/log"

	airshipv1 "opendev.org/airship/airshipctl/pkg/api/v1alpha1"
	"opendev.org/airship/airshipctl/pkg/clusterctl/implementations"
	"opendev.org/airship/airshipctl/pkg/log"
)

var _ Interface = &Client{}

const (
	// BootstrapProviderType is a local copy of appropriate type from cluster-api
	BootstrapProviderType = v1alpha3.BootstrapProviderType
	// CoreProviderType is a local copy of appropriate type from cluster-api
	CoreProviderType = v1alpha3.CoreProviderType
	// ControlPlaneProviderType is a local copy of appropriate type from cluster-api
	ControlPlaneProviderType = v1alpha3.ControlPlaneProviderType
	// InfrastructureProviderType is a local copy of appropriate type from cluster-api
	InfrastructureProviderType = v1alpha3.InfrastructureProviderType
)

// Interface is abstraction to Clusterctl
type Interface interface {
	Init(kubeconfigPath, kubeconfigContext string) error
	Move(fromKubeconfigPath, fromKubeconfigContext, toKubeconfigPath, toKubeconfigContext, namespace string) error
	GetKubeconfig(options *GetKubeconfigOptions) (string, error)
	Render(options RenderOptions) ([]byte, error)
}

// Client Implements interface to Clusterctl
type Client struct {
	clusterctlClient clusterctlclient.Client
	initOptions      clusterctlclient.InitOptions
	moveOptions      clusterctlclient.MoveOptions
	repoFactory      RepositoryFactory
}

// RenderOptions is used to get providers from RepoFactory for Render method
type RenderOptions struct {
	ProviderName    string
	ProviderVersion string
	ProviderType    string
}

// GetKubeconfigOptions carries all the options to retrieve kubeconfig from parent cluster
type GetKubeconfigOptions struct {
	// Path to parent kubeconfig file
	ParentKubeconfigPath string
	// Specify context within the kubeconfig file. If empty, cluster client
	// will use the current context.
	ParentKubeconfigContext string
	// Namespace is the namespace in which secret is placed.
	ManagedClusterNamespace string
	// ManagedClusterName is the name of the managed cluster.
	ManagedClusterName string
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
	cclient, rf, err := newClusterctlClient(root, options)
	if err != nil {
		return nil, err
	}
	return &Client{clusterctlClient: cclient, initOptions: cio, repoFactory: rf}, nil
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

func newClusterctlClient(root string, options *airshipv1.Clusterctl) (clusterctlclient.Client,
	RepositoryFactory, error) {
	cconf, err := newConfig(options, root)
	if err != nil {
		return nil, RepositoryFactory{}, err
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
	client, err := clusterctlclient.New("", ocf, orf, occf)
	return client, rf, err
}

// Render returns requested components as yaml
func (c *Client) Render(renderOptions RenderOptions) ([]byte, error) {
	provider, err := c.repoFactory.ConfigClient.Providers().Get(renderOptions.ProviderName,
		v1alpha3.ProviderType(renderOptions.ProviderType))
	if err != nil {
		return nil, err
	}

	crf := c.repoFactory.ClientRepositoryFactory()
	repoClient, err := crf(clusterctlclient.RepositoryClientFactoryInput{
		Provider:  provider,
		Processor: yamlprocessor.NewSimpleProcessor(),
	})
	if err != nil {
		return nil, err
	}

	components, err := repoClient.Components().Get(repository.ComponentsOptions{Version: renderOptions.ProviderVersion})
	if err != nil {
		return nil, err
	}
	return components.Yaml()
}

// GetKubeconfig is a wrapper for related cluster-api function
func (c *Client) GetKubeconfig(options *GetKubeconfigOptions) (string, error) {
	return c.clusterctlClient.GetKubeconfig(clusterctlclient.GetKubeconfigOptions{
		Kubeconfig: clusterctlclient.Kubeconfig{
			Path:    options.ParentKubeconfigPath,
			Context: options.ParentKubeconfigContext,
		},
		Namespace:           options.ManagedClusterNamespace,
		WorkloadClusterName: options.ManagedClusterName,
	})
}
