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

package applier_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"sigs.k8s.io/cli-utils/pkg/apply/event"
	"sigs.k8s.io/cli-utils/pkg/apply/poller"

	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/events"
	"opendev.org/airship/airshipctl/pkg/k8s/applier"
	"opendev.org/airship/airshipctl/testutil"
	k8stest "opendev.org/airship/airshipctl/testutil/k8sutils"
)

func TestFakeApplier(t *testing.T) {
	a := applier.NewFakeApplier(genericclioptions.IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}, k8stest.SuccessEvents(), k8stest.FakeFactory(t, []k8stest.ClientHandler{}))
	assert.NotNil(t, a)
}

// TODO Develop more coverage with tests
func TestApplierRun(t *testing.T) {
	bundle := testutil.NewTestBundle(t, "testdata/source_bundle")
	replicationController, err := bundle.SelectOne(document.NewSelector().ByKind("ReplicationController"))
	require.NoError(t, err)
	b, err := replicationController.AsYAML()
	require.NoError(t, err)
	f := k8stest.FakeFactory(t,
		[]k8stest.ClientHandler{
			&k8stest.InventoryObjectHandler{},
			&k8stest.NamespaceHandler{},
			&k8stest.GenericHandler{
				Obj:       &corev1.ReplicationController{},
				Bytes:     b,
				URLPath:   "/namespaces/%s/replicationcontrollers",
				Namespace: replicationController.GetNamespace(),
			},
		})
	defer f.Cleanup()
	s := genericclioptions.IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}
	tests := []struct {
		name           string
		driver         applier.Driver
		expectErr      bool
		expectedString string
		bundle         document.Bundle
		poller         poller.Poller
	}{
		{
			name:           "init-err",
			driver:         applier.NewFakeAdaptor().WithInitError(fmt.Errorf("init-err")),
			expectedString: "init-err",
			bundle:         bundle,
			expectErr:      true,
		},
		{
			name:           "can't reach cluster",
			expectedString: "connection refused",
			expectErr:      true,
			bundle:         bundle,
			poller:         &applier.FakePoller{},
		},
		{
			name:           "bundle failure",
			expectedString: "Can not apply nil bundle, airship.kubernetes.Client",
			expectErr:      true,
		},
		{
			name:      "success",
			expectErr: false,
			bundle:    bundle,
			driver:    applier.NewFakeAdaptor().WithEvents(k8stest.SuccessEvents()),
		},
		{
			name:      "set poller",
			expectErr: false,
			bundle:    bundle,
			driver:    applier.NewFakeAdaptor().WithEvents(k8stest.SuccessEvents()),
			poller:    &applier.FakePoller{},
		},
		{
			name:           "two configmaps present",
			expectErr:      true,
			bundle:         newBundle("testdata/two_cm_bundle", t),
			driver:         applier.NewFakeAdaptor().WithEvents(k8stest.SuccessEvents()),
			expectedString: "found more than one document",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			// create default applier
			a := applier.NewApplier(f, s)
			opts := applier.ApplyOptions{
				WaitTimeout: time.Second * 5,
				BundleName:  "test-bundle",
				DryRun:      true,
			}
			if tt.driver != nil {
				a.Driver = tt.driver
			}
			if tt.poller != nil {
				a.Poller = tt.poller
			}
			ch := a.ApplyBundle(tt.bundle, opts)
			var airEvents []events.Event
			for e := range ch {
				airEvents = append(airEvents, e)
			}
			var errs []error
			for _, e := range airEvents {
				if e.Type == events.ErrorType {
					errs = append(errs, e.ErrorEvent.Error)
				} else if e.Type == events.ApplierType && e.ApplierEvent.Type == event.ErrorType {
					errs = append(errs, e.ApplierEvent.ErrorEvent.Err)
				}
			}
			if tt.expectErr {
				t.Logf("errors are %v \n", errs)
				require.Len(t, errs, 1)
				require.NotNil(t, errs[0])
				// check if error contains  string
				assert.Contains(t, errs[0].Error(), tt.expectedString)
			} else {
				assert.Len(t, errs, 0)
			}
		})
	}
}

func newBundle(path string, t *testing.T) document.Bundle {
	t.Helper()
	b, err := document.NewBundleByPath(path)
	require.NoError(t, err)
	return b
}
