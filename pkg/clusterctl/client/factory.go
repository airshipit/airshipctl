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
	"sigs.k8s.io/cluster-api/cmd/clusterctl/client"
	"sigs.k8s.io/cluster-api/cmd/clusterctl/client/cluster"
	"sigs.k8s.io/cluster-api/cmd/clusterctl/client/config"
	"sigs.k8s.io/cluster-api/cmd/clusterctl/client/repository"

	airshipv1 "opendev.org/airship/airshipctl/pkg/clusterctl/api/v1alpha1"
	"opendev.org/airship/airshipctl/pkg/clusterctl/implementations"
	"opendev.org/airship/airshipctl/pkg/log"
)

// RepositoryFactory returns an injection factory to work with clusterctl client
type RepositoryFactory struct {
	Options      *airshipv1.Clusterctl
	ConfigClient config.Client
}

// ClusterClientFactory returns cluster factory function for clusterctl client
func (f RepositoryFactory) ClusterClientFactory() client.ClusterClientFactory {
	return func(kubeconfig client.Kubeconfig) (cluster.Client, error) {
		o := cluster.InjectRepositoryFactory(f.repoFactoryClusterClient())
		return cluster.New(cluster.Kubeconfig{
			Path:    kubeconfig.Path,
			Context: kubeconfig.Context}, f.ConfigClient, o), nil
	}
}

// ClientRepositoryFactory returns repo factory function for clusterctl client
func (f RepositoryFactory) ClientRepositoryFactory() client.RepositoryClientFactory {
	return f.repoFactory
}

// These two functions are basically the same, but have different with signatures
func (f RepositoryFactory) repoFactoryClusterClient() cluster.RepositoryClientFactory {
	return func(provider config.Provider,
		configClient config.Client,
		options ...repository.Option,
	) (repository.Client, error) {
		return f.repoFactory(provider)
	}
}

func (f RepositoryFactory) repoFactory(provider config.Provider) (repository.Client, error) {
	name := provider.Name()
	repoType := provider.Type()
	airProv := f.Options.Provider(name, repoType)
	if airProv == nil {
		return nil, ErrProviderRepoNotFound{ProviderName: name, ProviderType: string(repoType)}
	}
	// if repository is not clusterctl type, construct an airshipctl implementation of repository interface
	if !airProv.IsClusterctlRepository {
		// Get repository version map
		versions := airProv.Versions
		if len(versions) == 0 {
			return nil, ErrProviderRepoNotFound{ProviderName: name, ProviderType: string(repoType)}
		}
		// construct a repository for this provider using root and version map
		repo, err := implementations.NewRepository(provider.URL(), versions)
		if err != nil {
			return nil, err
		}
		// inject repository into repository client
		o := repository.InjectRepository(repo)
		log.Printf("Creating arishipctl repository implementation interface for provider %s of type %s\n",
			name,
			repoType)

		repoClient, err := repository.New(provider, f.ConfigClient, o)
		if err != nil {
			return nil, err
		}
		return &implementations.RepositoryClient{
			Client:       repoClient,
			ProviderType: string(repoType),
			ProviderName: name}, nil
	}
	log.Printf("Creating clusterctl repository implementation interface for provider %s of type %s\n",
		name,
		repoType)
	// if repository is clusterctl pass, simply use default clusterctl repository interface
	return repository.New(provider, f.ConfigClient)
}
