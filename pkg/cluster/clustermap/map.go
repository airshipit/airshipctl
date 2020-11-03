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

package clustermap

import (
	"opendev.org/airship/airshipctl/pkg/api/v1alpha1"
	"opendev.org/airship/airshipctl/pkg/log"
)

// ClusterMap interface that allows to list all clusters, find its parent, namespace,
// check if dynamic kubeconfig is enabled.
// TODO use typed cluster names
type ClusterMap interface {
	ParentCluster(string) (string, error)
	AllClusters() []string
	DynamicKubeConfig(string) bool
	ClusterNamespace(string) (string, error)
	ClusterKubeconfigContext(string) (string, error)
}

// clusterMap allows to view clusters and relationship between them
type clusterMap struct {
	apiMap *v1alpha1.ClusterMap
}

var _ ClusterMap = clusterMap{}

// NewClusterMap returns ClusterMap interface
func NewClusterMap(cMap *v1alpha1.ClusterMap) ClusterMap {
	return clusterMap{apiMap: cMap}
}

// ParentCluster finds a parent cluster for provided child
func (cm clusterMap) ParentCluster(child string) (string, error) {
	currentCluster, exists := cm.apiMap.Map[child]
	if !exists {
		return "", ErrClusterNotInMap{Child: child, Map: cm.apiMap}
	}
	if currentCluster.Parent == "" {
		return "", ErrParentNotFound{Child: child, Map: cm.apiMap}
	}
	return currentCluster.Parent, nil
}

// DynamicKubeConfig check if dynamic kubeconfig is enabled for the child cluster
func (cm clusterMap) DynamicKubeConfig(child string) bool {
	childCluster, exist := cm.apiMap.Map[child]
	if !exist {
		log.Debugf("cluster %s is not defined in cluster map %v", child, cm.apiMap)
		return false
	}
	return childCluster.DynamicKubeConfig
}

// AllClusters returns all clusters in a map
func (cm clusterMap) AllClusters() []string {
	clusters := []string{}
	for k := range cm.apiMap.Map {
		clusters = append(clusters, k)
	}
	return clusters
}

// ClusterNamespace a namespace for given cluster
// TODO implement how to get namespace for cluster
func (cm clusterMap) ClusterNamespace(clusterName string) (string, error) {
	return "default", nil
}

// ClusterNamespace a namespace for given cluster
// TODO implement how to get namespace for cluster
func (cm clusterMap) ClusterKubeconfigContext(clusterName string) (string, error) {
	cluster, exists := cm.apiMap.Map[clusterName]

	if !exists {
		return "", ErrClusterNotInMap{Map: cm.apiMap, Child: clusterName}
	}

	kubeContext := cluster.KubeconfigContext
	// if kubeContext is still empty, set it to clusterName
	if kubeContext == "" {
		kubeContext = clusterName
	}

	return kubeContext, nil
}
