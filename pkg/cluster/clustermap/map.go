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
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"opendev.org/airship/airshipctl/pkg/api/v1alpha1"
)

// DefaultClusterAPIObjNamespace is a default namespace used for cluster-api cluster object
const DefaultClusterAPIObjNamespace = "target-infra"

// WriteOptions has format in which we want to print the output(table/yaml/cluster name)
type WriteOptions struct {
	Format string
}

// ClusterMap interface that allows to list all clusters, find its parent, namespace,
// check if dynamic kubeconfig is enabled.
// TODO use typed cluster names
type ClusterMap interface {
	ParentCluster(string) (string, error)
	ValidateClusterMap() error
	AllClusters() []string
	ClusterKubeconfigContext(string) (string, error)
	Sources(string) ([]v1alpha1.KubeconfigSource, error)
	Write(io.Writer, WriteOptions) error
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

// Validates a clustermap has valid parent-child map structure
func (cm clusterMap) ValidateClusterMap() error {
	clusterMap := cm.AllClusters()
	for _, childCluster := range clusterMap {
		var parentClusters []string
		var currentChild string = childCluster
		for {
			currentCluster, _ := cm.apiMap.Map[currentChild]
			for _, c := range parentClusters {
				if c == currentCluster.Parent {
					// Quit on parent whos also child
					return ErrClusterCircularDependency{Parent: childCluster, Map: cm.apiMap}
				}
			}
			// Quit loop once top level of current cluster is reached
			if currentCluster.Parent == "" {
				break
			}
			parentClusters = append(parentClusters, currentCluster.Parent)
			currentChild = currentCluster.Parent
		}
	}

	// Return success if there are no conflicts
	return nil
}

// AllClusters returns all clusters in a map
func (cm clusterMap) AllClusters() []string {
	clusters := []string{}
	for k := range cm.apiMap.Map {
		clusters = append(clusters, k)
	}
	return clusters
}

// ClusterKubeconfigContext returns name of the context in kubeconfig corresponding to a given cluster
func (cm clusterMap) ClusterKubeconfigContext(clusterName string) (string, error) {
	_, exists := cm.apiMap.Map[clusterName]

	if !exists {
		return "", ErrClusterNotInMap{Map: cm.apiMap, Child: clusterName}
	}

	return clusterName, nil
}

func (cm clusterMap) Sources(clusterName string) ([]v1alpha1.KubeconfigSource, error) {
	cluster, ok := cm.apiMap.Map[clusterName]
	if !ok {
		return nil, ErrClusterNotInMap{Child: clusterName, Map: cm.apiMap}
	}
	return cluster.Sources, nil
}

// Write prints the cluster list in table/name output format
func (cm clusterMap) Write(writer io.Writer, wo WriteOptions) error {
	if wo.Format == "table" {
		w := tabwriter.NewWriter(os.Stdout, 20, 8, 1, ' ', 0)
		fmt.Fprintf(w, "NAME\tKUBECONFIG CONTEXT\tPARENT CLUSTER\n")
		for clustername, cluster := range cm.apiMap.Map {
			kubeconfig, err := cm.ClusterKubeconfigContext(clustername)
			if err != nil {
				return err
			}
			fmt.Fprintf(w, "%s\t%s\t%s\n",
				clustername, kubeconfig, cluster.Parent)
		}
		w.Flush()
	} else if wo.Format == "name" {
		clusterList := cm.AllClusters()
		for _, clusterName := range clusterList {
			if _, err := writer.Write([]byte(clusterName + "\n")); err != nil {
				return err
			}
		}
	}
	return nil
}
