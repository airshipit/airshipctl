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
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/cli-utils/pkg/kstatus/polling"
	"sigs.k8s.io/cli-utils/pkg/kstatus/polling/clusterreader"
	"sigs.k8s.io/cli-utils/pkg/kstatus/polling/engine"
	"sigs.k8s.io/cli-utils/pkg/kstatus/polling/event"
	"sigs.k8s.io/cli-utils/pkg/kstatus/polling/statusreaders"
	"sigs.k8s.io/cli-utils/pkg/object"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"opendev.org/airship/airshipctl/pkg/log"
)

const allowedApplyErrors = 3

// NewStatusPoller creates a new StatusPoller using the given clusterreader and mapper. The StatusPoller
// will use the client for all calls to the cluster.
func NewStatusPoller(reader client.Reader, mapper meta.RESTMapper) *StatusPoller {
	return &StatusPoller{
		engine: &engine.PollerEngine{
			Reader: reader,
			Mapper: mapper,
		},
	}
}

// StatusPoller provides functionality for polling a cluster for status for a set of resources.
type StatusPoller struct {
	engine *engine.PollerEngine
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
			cr, err := clusterreader.NewCachingClusterReader(r, mapper, identifiers)
			if err != nil {
				return nil, err
			}
			return &CachingClusterReader{Cr: cr}, nil
		}
		return &clusterreader.DirectClusterReader{Reader: r}, nil
	}
}

// CachingClusterReader is wrapper for kstatus.CachingClusterReader implementation
type CachingClusterReader struct {
	Cr          *clusterreader.CachingClusterReader
	applyErrors []error
}

// Get is a wrapper for kstatus.CachingClusterReader Get method
func (c *CachingClusterReader) Get(ctx context.Context, key client.ObjectKey, obj *unstructured.Unstructured) error {
	return c.Cr.Get(ctx, key, obj)
}

// ListNamespaceScoped is a wrapper for kstatus.CachingClusterReader ListNamespaceScoped method
func (c *CachingClusterReader) ListNamespaceScoped(
	ctx context.Context,
	list *unstructured.UnstructuredList,
	namespace string,
	selector labels.Selector) error {
	return c.Cr.ListNamespaceScoped(ctx, list, namespace, selector)
}

// ListClusterScoped is a wrapper for kstatus.CachingClusterReader ListClusterScoped method
func (c *CachingClusterReader) ListClusterScoped(
	ctx context.Context,
	list *unstructured.UnstructuredList,
	selector labels.Selector) error {
	return c.Cr.ListClusterScoped(ctx, list, selector)
}

// Sync is a wrapper for kstatus.CachingClusterReader Sync method, allows to filter specific errors
func (c *CachingClusterReader) Sync(ctx context.Context) error {
	err := c.Cr.Sync(ctx)
	if err != nil && strings.Contains(err.Error(), "request timed out") {
		c.applyErrors = append(c.applyErrors, err)
		if len(c.applyErrors) < allowedApplyErrors {
			log.Printf("timeout error occurred during sync: '%v', skipping", err)
			return nil
		}
	}
	return err
}
