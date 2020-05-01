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

package client

import (
	"context"

	"opendev.org/airship/airshipctl/pkg/log"

	bmoapis "github.com/metal3-io/baremetal-operator/pkg/apis"
	bmh "github.com/metal3-io/baremetal-operator/pkg/apis/metal3/v1alpha1"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	clusterctlclient "sigs.k8s.io/cluster-api/cmd/clusterctl/client"
	"sigs.k8s.io/cluster-api/cmd/clusterctl/client/cluster"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func init() {
	//nolint:errcheck
	bmoapis.AddToScheme(cluster.Scheme)
}

// Move implements interface to Clusterctl
func (c *Client) Move(fromKubeconfigPath, fromKubeconfigContext,
	toKubeconfigPath, toKubeconfigContext, namespace string) error {
	ctx := context.TODO()
	var err error
	// ephemeral cluster client
	pFrom := cluster.New(cluster.Kubeconfig{
		Path:    fromKubeconfigPath,
		Context: fromKubeconfigContext}, nil).Proxy()
	cFrom, err := pFrom.NewClient()
	if err != nil {
		return errors.Wrap(err, "failed to create ephemeral cluster client")
	}
	// target cluster client
	pTo := cluster.New(cluster.Kubeconfig{
		Path:    toKubeconfigPath,
		Context: toKubeconfigContext}, nil).Proxy()
	cTo, err := pTo.NewClient()
	if err != nil {
		return errors.Wrap(err, "failed to create target cluster client")
	}
	// If namespace is empty, try to detect it.
	if namespace == "" {
		var currentNamespace string
		currentNamespace, err = pFrom.CurrentNamespace()
		if err != nil {
			return err
		}
		namespace = currentNamespace
	}
	// Pause
	err = pauseUnpauseBMHs(ctx, cFrom, namespace, true)
	if err != nil {
		return errors.Wrap(err, "failed to pause BareMetalHost objects")
	}

	// clusterctl move
	c.moveOptions = clusterctlclient.MoveOptions{
		FromKubeconfig: clusterctlclient.Kubeconfig{Path: fromKubeconfigPath, Context: fromKubeconfigContext},
		ToKubeconfig:   clusterctlclient.Kubeconfig{Path: toKubeconfigPath, Context: toKubeconfigContext},
		Namespace:      namespace,
	}
	err = c.clusterctlClient.Move(c.moveOptions)
	if err != nil {
		return errors.Wrapf(err, "error during clusterctl move")
	}
	// Update BMH Status
	err = copyBMHStatus(ctx, cFrom, cTo, namespace)
	if err != nil {
		return errors.Wrap(err, "failed to copy BareMetalHost Status")
	}
	// Unpause
	err = pauseUnpauseBMHs(ctx, cFrom, namespace, false)
	if err != nil {
		return errors.Wrap(err, "failed to unpause BareMetalHost objects")
	}
	return err
}

// copyBMHStatus will copy the BareMetalHost Status field from a specific
// cluser to a target cluster.
func copyBMHStatus(ctx context.Context, cFrom client.Client, cTo client.Client, namespace string) error {
	fromHosts, err := getBMHs(ctx, cFrom, namespace)
	if err != nil {
		return errors.Wrap(err, "failed to list BareMetalHost objects")
	}
	toHosts, err := getBMHs(ctx, cTo, namespace)
	if err != nil {
		return errors.Wrap(err, "failed to list BMH objects")
	}
	// Copy the Status field from old BMH to new BMH
	log.Debugf("Copying BareMetalHost status to target cluster")
	for _, toHost := range toHosts.Items {
		var found bool
		t := metav1.Now()
		for _, fromHost := range fromHosts.Items {
			if fromHost.Name == toHost.Name {
				toHost.Status = fromHost.Status
				found = true
				break
			}
		}
		if !found {
			return errors.Errorf("BMH with the same name %s/%s not found in the source cluster", toHost.Name, namespace)
		}
		toHost.Status.LastUpdated = &t
		err = cTo.Status().Update(ctx, &toHost)
		if err != nil {
			return errors.Wrap(err, "failed to update BareMetalHost status")
		}
	}
	return nil
}

// pauseUnpauseBMHs will add/remove the pause annotation from the
// BareMetalHost objects.
func pauseUnpauseBMHs(ctx context.Context, crClient client.Client, namespace string, pause bool) error {
	hosts, err := getBMHs(ctx, crClient, namespace)
	if err != nil {
		return errors.Wrap(err, "failed to list BMH objects")
	}
	for _, host := range hosts.Items {
		annotations := host.GetAnnotations()
		if annotations == nil {
			host.Annotations = map[string]string{}
		}
		if pause {
			log.Debugf("Pausing BareMetalHost object %s/%s", host.Name, namespace)
			host.Annotations[bmh.PausedAnnotation] = "true"
		} else {
			log.Debugf("Unpausing BareMetalHost object %s/%s", host.Name, namespace)
			delete(host.Annotations, bmh.PausedAnnotation)
		}
		if err := crClient.Update(ctx, &host); err != nil {
			return errors.Wrapf(err, "error updating BareMetalHost %q %s/%s",
				host.GroupVersionKind(), host.GetNamespace(), host.GetName())
		}
	}
	return nil
}

// getBMHs will return all BareMetalHost objects in the specified namepace.
// It also checks to see if the BareMetalHost resource is installed, if not,
// it will return false.
func getBMHs(ctx context.Context, crClient client.Client, namespace string) (bmh.BareMetalHostList, error) {
	hosts := bmh.BareMetalHostList{}
	opts := &client.ListOptions{
		Namespace: namespace,
	}
	err := crClient.List(ctx, &hosts, opts)
	return hosts, err
}
