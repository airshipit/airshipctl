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
	"path/filepath"

	"opendev.org/airship/airshipctl/pkg/api/v1alpha1"
	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/errors"
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

	clusterMap *v1alpha1.ClusterMap
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
func (b *Builder) WithClusterMap(cMap *v1alpha1.ClusterMap) *Builder {
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

// Build builds a kubeconfig interface to be used
func (b *Builder) Build() Interface {
	switch {
	case b.path != "":
		fs := document.NewDocumentFs()
		return NewKubeConfig(FromFile(b.path, fs), InjectFilePath(b.path, fs), InjectTempRoot(b.root))
	case b.fromParent():
		// TODO add method that would get kubeconfig from parent cluster and glue it together
		// with parent kubeconfig if needed
		return NewKubeConfig(func() ([]byte, error) {
			return nil, errors.ErrNotImplemented{}
		})
	case b.bundlePath != "":
		return NewKubeConfig(FromBundle(b.bundlePath), InjectTempRoot(b.root))
	default:
		fs := document.NewDocumentFs()
		// return default path to kubeconfig file in airship workdir
		path := filepath.Join(util.UserHomeDir(), config.AirshipConfigDir, KubeconfigDefaultFileName)
		return NewKubeConfig(FromFile(path, fs), InjectFilePath(path, fs), InjectTempRoot(b.root))
	}
}

// fromParent checks if we should get kubeconfig from parent cluster secret
func (b *Builder) fromParent() bool {
	if b.clusterMap == nil {
		return false
	}
	currentCluster, exists := b.clusterMap.Map[b.clusterName]
	if !exists {
		log.Debugf("cluster %s is not defined in cluster map %v", b.clusterName, b.clusterMap)
		return false
	}
	// Check if DynamicKubeConfig is enabled, if so that means, we should get kubeconfig
	// for this cluster from its parent
	if currentCluster.Parent == "" || !currentCluster.DynamicKubeConfig {
		log.Debugf("dynamic kubeconfig or parent cluster is not set for cluster %s", b.clusterName)
		return false
	}
	return true
}
