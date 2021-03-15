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

package kubeconfig

import (
	"fmt"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"

	"opendev.org/airship/airshipctl/pkg/api/v1alpha1"
	"opendev.org/airship/airshipctl/pkg/cluster/clustermap"
	"opendev.org/airship/airshipctl/pkg/clusterctl/client"
	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/fs"
	"opendev.org/airship/airshipctl/pkg/log"
)

// KubeconfigDefaultFileName is a default name for kubeconfig
const KubeconfigDefaultFileName = "kubeconfig"

// NewBuilder returns instance of kubeconfig builder.
func NewBuilder() *Builder {
	return &Builder{
		siteKubeconf: emptyConfig(),
	}
}

// Builder is an object that allows to build a kubeconfig based on various provided sources
// such as path to kubeconfig, path to bundle that should contain kubeconfig and parent cluster
type Builder struct {
	clusterName string
	root        string

	bundle           document.Bundle
	clusterMap       clustermap.ClusterMap
	clusterctlClient client.Interface
	fs               fs.FileSystem
	siteKubeconf     *api.Config
}

// WithBundle allows to set document.Bundle object that should contain kubeconfig api object
func (b *Builder) WithBundle(bundle document.Bundle) *Builder {
	b.bundle = bundle
	return b
}

// WithClusterMap allows to set a parent cluster, that can be used to extract kubeconfig for target cluster
func (b *Builder) WithClusterMap(cMap clustermap.ClusterMap) *Builder {
	b.clusterMap = cMap
	return b
}

// WithClusterName allows to reach to a cluster to download kubeconfig from there
func (b *Builder) WithClusterName(clusterName string) *Builder {
	b.clusterName = clusterName
	return b
}

// WithTempRoot allows to set temp root for kubeconfig
func (b *Builder) WithTempRoot(root string) *Builder {
	b.root = root
	return b
}

// WithClusterctClient this is used if u want to inject your own clusterctl
// mostly needed for tests
func (b *Builder) WithClusterctClient(c client.Interface) *Builder {
	b.clusterctlClient = c
	return b
}

// WithFilesytem allows to set filesystem
func (b *Builder) WithFilesytem(fs fs.FileSystem) *Builder {
	b.fs = fs
	return b
}

// Build site kubeconfig, ignores, but logs, errors that happen when building individual
// kubeconfigs. We need this behavior because, some clusters may not yet be deployed
// and their kubeconfig is inaccessible yet, but will be accessible at later phases
// If builder can't build kubeconfig for specific cluster, its context will not be present
// in final kubeconfig. User of kubeconfig, will receive error stating that context doesn't exist
func (b *Builder) Build() Interface {
	return NewKubeConfig(b.build, InjectFileSystem(b.fs), InjectTempRoot(b.root))
}

func (b *Builder) build() ([]byte, error) {
	// Set current context to clustername if it was provided
	var result *api.Config
	var err error
	var kubeContext string
	if b.clusterName != "" {
		kubeContext, result, err = b.buildOne(b.clusterName)
		if err != nil {
			return nil, err
		}
	} else {
		result, err = b.builtSiteKubeconf()
		if err != nil {
			return nil, err
		}
	}
	b.siteKubeconf.CurrentContext = kubeContext
	return clientcmd.Write(*result)
}

func (b *Builder) builtSiteKubeconf() (*api.Config, error) {
	log.Debugf("Getting site kubeconfig")
	for _, clusterID := range b.clusterMap.AllClusters() {
		log.Debugf("Getting kubeconfig for cluster '%s' to build site kubeconfig", clusterID)
		// buildOne merges context into site kubeconfig
		_, _, err := b.buildOne(clusterID)
		if IsErrAllSourcesFailedErr(err) {
			log.Printf("All kubeconfig sources failed for cluster '%s', error '%v', skipping it",
				clusterID, err)
			continue
		} else if err != nil {
			return nil, err
		}
	}
	return b.siteKubeconf, nil
}

func (b *Builder) buildOne(clusterID string) (string, *api.Config, error) {
	destContext, err := b.clusterMap.ClusterKubeconfigContext(clusterID)
	if err != nil {
		return "", nil, err
	}

	// use already built kubeconfig context, to avoid doing work multiple times
	built, oneKubeconf := b.alreadyBuilt(destContext)
	if built {
		log.Printf("kubeconfig for cluster '%s' is already built, using it", clusterID)
		return destContext, oneKubeconf, nil
	}

	sources, err := b.clusterMap.Sources(clusterID)
	if err != nil {
		return "", nil, err
	}

	for _, source := range sources {
		oneKubeconf, sourceErr := b.trySource(clusterID, destContext, source)
		if sourceErr == nil {
			// Merge source context into site kubeconfig
			log.Printf("Merging kubecontext for cluster '%s', into site kubeconfig", clusterID)
			if err = mergeContextAPI(destContext, destContext, b.siteKubeconf, oneKubeconf); err != nil {
				return "", nil, err
			}
			return destContext, oneKubeconf, err
		}
		// if error, log it and ignore it. missing problem with one kubeconfig should not
		// effect other clusters, which don't depend on it. If they do depend on it, their calls
		// will fail because the context will be missing. Combitation with a log message will make
		// it clear where the problem is.
		log.Printf("received error while trying kubeconfig source for cluster '%s', source type '%s', error '%v'",
			clusterID, source.Type, sourceErr)
	}
	// return empty not nil kubeconfig without error.
	return "", nil, &ErrAllSourcesFailed{ClusterName: clusterID}
}

func (b *Builder) trySource(clusterID, dstContext string, source v1alpha1.KubeconfigSource) (*api.Config, error) {
	var getter KubeSourceFunc
	// TODO add sourceContext defaults
	var sourceContext string
	switch source.Type {
	case v1alpha1.KubeconfigSourceTypeFilesystem:
		getter = FromFile(source.FileSystem.Path, b.fs)
		sourceContext = source.FileSystem.Context
	case v1alpha1.KubeconfigSourceTypeBundle:
		getter = FromBundle(b.bundle)
		sourceContext = source.Bundle.Context
	case v1alpha1.KubeconfigSourceTypeClusterAPI:
		getter = b.fromClusterAPI(clusterID, source.ClusterAPI)
	default:
		// TODO add validation for fast fails to clustermap interface instead of this
		return nil, &ErrUnknownKubeconfigSourceType{Type: string(source.Type)}
	}
	kubeBytes, err := getter()
	if err != nil {
		return nil, err
	}
	return extractContext(dstContext, sourceContext, kubeBytes)
}

func (b *Builder) fromClusterAPI(clusterName string, ref v1alpha1.KubeconfigSourceClusterAPI) KubeSourceFunc {
	return func() ([]byte, error) {
		log.Printf("Getting kubeconfig from cluster API for cluster '%s'", clusterName)
		parentCluster, err := b.clusterMap.ParentCluster(clusterName)
		if err != nil {
			return nil, err
		}

		parentContext, parentKubeconf, err := b.buildOne(parentCluster)
		if err != nil {
			return nil, err
		}

		parentKubeconfig := NewKubeConfig(FromConfig(parentKubeconf), InjectFileSystem(b.fs))

		f, cleanup, err := parentKubeconfig.GetFile()
		if err != nil {
			return nil, err
		}
		defer cleanup()

		if b.clusterctlClient == nil {
			b.clusterctlClient, err = client.NewClient("", log.DebugEnabled(), v1alpha1.DefaultClusterctl())
			if err != nil {
				return nil, err
			}
		}

		log.Printf("Getting child kubeconfig from parent, parent context '%s', parent kubeconfing '%s'",
			parentContext, f)
		return FromSecret(b.clusterctlClient, &client.GetKubeconfigOptions{
			ParentKubeconfigPath:    f,
			ParentKubeconfigContext: parentContext,
			ManagedClusterNamespace: ref.Namespace,
			ManagedClusterName:      ref.Name,
		})()
	}
}

func (b *Builder) alreadyBuilt(clusterContext string) (bool, *api.Config) {
	kubeconfBytes, err := clientcmd.Write(*b.siteKubeconf)
	if err != nil {
		log.Debugf("Received error when converting kubeconfig to bytes, ignoring kubeconfig. Error: %v", err)
		return false, nil
	}

	// resulting and existing context names must be the same, otherwise error will be returned
	clusterKubeconfig, err := extractContext(clusterContext, clusterContext, kubeconfBytes)
	if err != nil {
		log.Debugf("Received error when extacting context, ignoring kubeconfig. Error: %v", err)
		return false, nil
	}

	return true, clusterKubeconfig
}

func extractContext(destContext, sourceContext string, src []byte) (*api.Config, error) {
	srcKubeconf, err := clientcmd.Load(src)
	if err != nil {
		return nil, err
	}
	dstKubeconf := emptyConfig()
	return dstKubeconf, mergeContextAPI(destContext, sourceContext, dstKubeconf, srcKubeconf)
}

// merges two kubeconfigs
func mergeContextAPI(destContext, sourceContext string, dst, src *api.Config) error {
	if len(src.Contexts) > 1 && sourceContext == "" {
		// When more than one context, we don't know which to choose
		return &ErrKubeconfigMergeFailed{
			Message: "kubeconfig has multiple contexts, don't know which to choose, " +
				"please specify contextName in clusterMap cluster kubeconfig source",
		}
	}

	var context *api.Context
	context, exists := src.Contexts[sourceContext]
	switch {
	case exists:
	case sourceContext == "" && len(src.Contexts) == 1:
		for _, context = range src.Contexts {
			log.Debugf("Using context '%v' to merge kubeconfig", context)
		}
	default:
		return &ErrKubeconfigMergeFailed{
			Message: fmt.Sprintf("source context '%s' does not exist in source kubeconfig", sourceContext),
		}
	}
	dst.Contexts[destContext] = context

	// TODO design logic to make authinfo keys unique, they can overlap, or human error can occur
	user, exists := src.AuthInfos[context.AuthInfo]
	if !exists {
		return &ErrKubeconfigMergeFailed{
			Message: fmt.Sprintf("user '%s' does not exist in source kubeconfig", context.AuthInfo),
		}
	}
	dst.AuthInfos[context.AuthInfo] = user

	// TODO design logic to make cluster keys unique, they can overlap, or human error can occur
	cluster, exists := src.Clusters[context.Cluster]
	if !exists {
		return &ErrKubeconfigMergeFailed{
			Message: fmt.Sprintf("cluster '%s' does not exist in source kubeconfig", context.Cluster),
		}
	}
	dst.Clusters[context.Cluster] = cluster

	return nil
}

func emptyConfig() *api.Config {
	return &api.Config{
		Contexts:  make(map[string]*api.Context),
		AuthInfos: make(map[string]*api.AuthInfo),
		Clusters:  make(map[string]*api.Cluster),
	}
}
