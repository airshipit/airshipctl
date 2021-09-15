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

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/klog/v2"
	"sigs.k8s.io/cli-utils/pkg/kstatus/polling/clusterreader"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const allowedApplyErrors = 3

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
		if len(c.applyErrors) <= allowedApplyErrors {
			klog.V(2).Infof("timeout error occurred during sync: '%v', skipping", err)
			return nil
		}
	}
	return err
}
