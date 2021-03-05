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

	"opendev.org/airship/airshipctl/pkg/api/v1alpha1"
)

// ErrParentNotFound returned when requested cluster is not defined or doesn't have a parent
type ErrParentNotFound struct {
	Child string
	Map   *v1alpha1.ClusterMap
}

func (e ErrParentNotFound) Error() string {
	return fmt.Sprintf("failed to find a parent for cluster %s in cluster map %v", e.Child, e.Map)
}

// ErrClusterNotInMap returned when requested cluster is not defined in cluster map
type ErrClusterNotInMap struct {
	Child string
	Map   *v1alpha1.ClusterMap
}

func (e ErrClusterNotInMap) Error() string {
	return fmt.Sprintf("cluster '%s' is not defined in cluster map %v", e.Child, e.Map)
}
