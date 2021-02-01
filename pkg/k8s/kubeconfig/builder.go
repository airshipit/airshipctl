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
	"bytes"
	"path/filepath"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"

	"opendev.org/airship/airshipctl/pkg/api/v1alpha1"
	"opendev.org/airship/airshipctl/pkg/cluster/clustermap"
	"opendev.org/airship/airshipctl/pkg/clusterctl/client"
	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/fs"
	"opendev.org/airship/airshipctl/pkg/log"
	"opendev.org/airship/airshipctl/pkg/util"
)

// KubeconfigDefaultFileName is a default name for kubeconfig
const KubeconfigDefaultFileName = "kubeconfig"

// NewBuilder returns instance of kubeconfig builder.
func NewBuilder() *Builder {
	return &Builder{}
}

// Builder is an object that allows to build a kubeconfig based on various provided sources
// such as path to kubeconfig, path to bundle that should contain kubeconfig and parent cluster
type Builder struct {
	path        string
	bundlePath  string
	clusterName string
	root        string

	clusterMap       clustermap.ClusterMap
	clusterctlClient client.Interface
	fs               fs.FileSystem
}

// WithPath allows to set path to prexisting kubeconfig
func (b *Builder) WithPath(filePath string) *Builder {
	b.path = filePath
	return b
}

// WithBundle allows to set path to bundle that should contain kubeconfig api object
func (b *Builder) WithBundle(bundlePath string) *Builder {
	b.bundlePath = bundlePath
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

// Build builds a kubeconfig interface to be used
func (b *Builder) Build() Interface {
	switch {
	case b.path != "":
		return NewKubeConfig(FromFile(b.path, b.fs), InjectFilePath(b.path, b.fs), InjectTempRoot(b.root))
	case b.fromParent():
		// TODO consider adding various drivers to source kubeconfig from
		// Also consider accumulating different kubeconfigs, and returning one single
		// large file, so that every executor has access to all parent clusters.
		return NewKubeConfig(b.buildClusterctlFromParent, InjectTempRoot(b.root), InjectFileSystem(b.fs))
	case b.bundlePath != "":
		return NewKubeConfig(FromBundle(b.bundlePath), InjectTempRoot(b.root), InjectFileSystem(b.fs))
	default:
		// return default path to kubeconfig file in airship workdir
		path := filepath.Join(util.UserHomeDir(), config.AirshipConfigDir, KubeconfigDefaultFileName)
		return NewKubeConfig(FromFile(path, b.fs), InjectFilePath(path, b.fs), InjectTempRoot(b.root))
	}
}

// fromParent checks if we should get kubeconfig from parent cluster secret
func (b *Builder) fromParent() bool {
	if b.clusterMap == nil {
		return false
	}
	return b.clusterMap.DynamicKubeConfig(b.clusterName)
}

func (b *Builder) buildClusterctlFromParent() ([]byte, error) {
	currentCluster := b.clusterName
	log.Printf("current cluster name is '%s'",
		currentCluster)
	parentCluster, err := b.clusterMap.ParentCluster(currentCluster)
	if err != nil {
		return nil, err
	}

	parentKubeconfig := b.WithClusterName(parentCluster).Build()

	f, cleanup, err := parentKubeconfig.GetFile()
	if err != nil {
		return nil, err
	}
	defer cleanup()

	parentCtx, err := b.clusterMap.ClusterKubeconfigContext(parentCluster)
	if err != nil {
		return nil, err
	}

	clusterAPIRef, err := b.clusterMap.ClusterAPIRef(currentCluster)
	if err != nil {
		return nil, err
	}

	if b.clusterctlClient == nil {
		b.clusterctlClient, err = client.NewClient("", log.DebugEnabled(), v1alpha1.DefaultClusterctl())
		if err != nil {
			return nil, err
		}
	}

	log.Printf("Getting child kubeconfig from parent, parent context '%s', parent kubeconfing '%s'",
		parentCtx, f)

	stringChild, err := b.clusterctlClient.GetKubeconfig(&client.GetKubeconfigOptions{
		ParentKubeconfigPath:    f,
		ParentKubeconfigContext: parentCtx,
		ManagedClusterNamespace: clusterAPIRef.Namespace,
		ManagedClusterName:      clusterAPIRef.Name,
	})
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer([]byte{})

	err = parentKubeconfig.Write(buf)
	if err != nil {
		return nil, err
	}

	parentObj, err := clientcmd.Load(buf.Bytes())
	if err != nil {
		return nil, err
	}

	childObj, err := clientcmd.Load([]byte(stringChild))
	if err != nil {
		return nil, err
	}

	childCtx, err := b.clusterMap.ClusterKubeconfigContext(currentCluster)
	if err != nil {
		return nil, err
	}
	log.Printf("Merging '%s' cluster kubeconfig into '%s' cluster kubeconfig",
		currentCluster, parentCluster)

	return b.mergeOneContext(childCtx, parentObj, childObj)
}

// merges two kubeconfigs,
func (b *Builder) mergeOneContext(contextOverride string, dst, src *api.Config) ([]byte, error) {
	for key, content := range src.AuthInfos {
		dst.AuthInfos[key] = content
	}

	for key, content := range src.Clusters {
		dst.Clusters[key] = content
	}

	if len(src.Contexts) != 1 {
		return nil, &ErrClusterctlKubeconfigWrongContextsCount{
			ContextCount: len(src.Contexts),
		}
	}

	for key, content := range src.Contexts {
		if contextOverride == "" {
			contextOverride = key
		}
		dst.Contexts[contextOverride] = content
	}

	return clientcmd.Write(*dst)
}
