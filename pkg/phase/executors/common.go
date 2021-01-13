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

package executors

import (
	"k8s.io/apimachinery/pkg/runtime/schema"

	airshipv1 "opendev.org/airship/airshipctl/pkg/api/v1alpha1"
	"opendev.org/airship/airshipctl/pkg/events"
	"opendev.org/airship/airshipctl/pkg/phase/ifc"
)

// Constants related to phase executor names
const (
	Clusterctl       = "clusterctl"
	KubernetesApply  = "kubernetes-apply"
	Isogen           = "isogen"
	GenericContainer = "generic-container"
	Ephemeral        = "ephemeral"
)

// RegisterExecutor adds executor to phase executor registry
func RegisterExecutor(executorName string, registry map[schema.GroupVersionKind]ifc.ExecutorFactory) error {
	var gvks []schema.GroupVersionKind
	var execObj ifc.ExecutorFactory
	var err error

	switch executorName {
	case Clusterctl:
		gvks, _, err = airshipv1.Scheme.ObjectKinds(&airshipv1.Clusterctl{})
		execObj = NewClusterctlExecutor
	case KubernetesApply:
		gvks, _, err = airshipv1.Scheme.ObjectKinds(&airshipv1.KubernetesApply{})
		execObj = NewKubeApplierExecutor
	case Isogen:
		gvks, _, err = airshipv1.Scheme.ObjectKinds(airshipv1.DefaultIsoConfiguration())
		execObj = NewIsogenExecutor
	case GenericContainer:
		gvks, _, err = airshipv1.Scheme.ObjectKinds(airshipv1.DefaultGenericContainer())
		execObj = NewContainerExecutor
	case Ephemeral:
		gvks, _, err = airshipv1.Scheme.ObjectKinds(airshipv1.DefaultBootConfiguration())
		execObj = NewEphemeralExecutor
	default:
		return ErrUnknownExecutorName{ExecutorName: executorName}
	}

	if err != nil {
		return err
	}
	registry[gvks[0]] = execObj
	return nil
}

func handleError(ch chan<- events.Event, err error) {
	ch <- events.NewEvent().WithErrorEvent(events.ErrorEvent{
		Error: err,
	})
}
