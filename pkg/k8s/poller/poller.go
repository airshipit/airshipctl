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

package poller

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/cli-utils/pkg/kstatus/polling"
	"sigs.k8s.io/cli-utils/pkg/kstatus/polling/clusterreader"
	"sigs.k8s.io/cli-utils/pkg/kstatus/polling/engine"
	"sigs.k8s.io/cli-utils/pkg/kstatus/polling/event"
	"sigs.k8s.io/cli-utils/pkg/kstatus/polling/statusreaders"
	"sigs.k8s.io/cli-utils/pkg/object"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"opendev.org/airship/airshipctl/pkg/cluster"
)

// NewStatusPoller creates a new StatusPoller using the given clusterreader and mapper. The StatusPoller
// will use the client for all calls to the cluster.
func NewStatusPoller(reader client.Reader, mapper meta.RESTMapper, statusmap *cluster.StatusMap) *StatusPoller {
	return &StatusPoller{
		engine: &engine.PollerEngine{
			Reader: reader,
			Mapper: mapper,
		},
		statusmap: statusmap,
	}
}

// StatusPoller provides functionality for polling a cluster for status for a set of resources.
type StatusPoller struct {
	engine    *engine.PollerEngine
	statusmap *cluster.StatusMap
}

// Poll will create a new statusPollerRunner that will poll all the resources provided and report their status
// back on the event channel returned. The statusPollerRunner can be canceled at any time by canceling the
// context passed in.
func (s *StatusPoller) Poll(
	ctx context.Context, identifiers []object.ObjMetadata, options polling.Options) <-chan event.Event {
	return s.engine.Poll(ctx, identifiers, engine.Options{
		PollInterval:             options.PollInterval,
		ClusterReaderFactoryFunc: clusterReaderFactoryFunc(options.UseCache),
		StatusReadersFactoryFunc: s.createStatusReaders,
	})
}

// createStatusReaders creates an instance of all the statusreaders. This includes a set of statusreaders for
// a particular GroupKind, and a default engine used for all resource types that does not have
// a specific statusreaders.
// TODO: We should consider making the registration more automatic instead of having to create each of them
// here. Also, it might be worth creating them on demand.
func (s *StatusPoller) createStatusReaders(reader engine.ClusterReader, mapper meta.RESTMapper) (
	map[schema.GroupKind]engine.StatusReader, engine.StatusReader) {
	defaultStatusReader := statusreaders.NewGenericStatusReader(reader, mapper)
	replicaSetStatusReader := statusreaders.NewReplicaSetStatusReader(reader, mapper, defaultStatusReader)
	deploymentStatusReader := statusreaders.NewDeploymentResourceReader(reader, mapper, replicaSetStatusReader)
	statefulSetStatusReader := statusreaders.NewStatefulSetResourceReader(reader, mapper, defaultStatusReader)

	statusReaders := map[schema.GroupKind]engine.StatusReader{
		appsv1.SchemeGroupVersion.WithKind("Deployment").GroupKind():  deploymentStatusReader,
		appsv1.SchemeGroupVersion.WithKind("StatefulSet").GroupKind(): statefulSetStatusReader,
		appsv1.SchemeGroupVersion.WithKind("ReplicaSet").GroupKind():  replicaSetStatusReader,
	}
	for _, gk := range s.statusmap.GkMapping {
		statusReaders[gk] = s.statusmap
	}
	return statusReaders, defaultStatusReader
}

// clusterReaderFactoryFunc returns a factory function for creating an instance of a ClusterReader.
// This function is used by the StatusPoller to create a ClusterReader for each StatusPollerRunner.
// The decision for which implementation of the ClusterReader interface that should be used are
// decided here rather than based on information passed in to the factory function. Thus, the decision
// for which implementation is decided when the StatusPoller is created.
func clusterReaderFactoryFunc(useCache bool) engine.ClusterReaderFactoryFunc {
	return func(r client.Reader, mapper meta.RESTMapper, identifiers []object.ObjMetadata) (engine.ClusterReader, error) {
		if useCache {
			return clusterreader.NewCachingClusterReader(r, mapper, identifiers)
		}
		return &clusterreader.DirectClusterReader{Reader: r}, nil
	}
}
