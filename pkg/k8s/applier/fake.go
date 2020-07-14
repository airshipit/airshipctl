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

package applier

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/resource"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	cliapply "sigs.k8s.io/cli-utils/pkg/apply"
	applyevent "sigs.k8s.io/cli-utils/pkg/apply/event"
	"sigs.k8s.io/cli-utils/pkg/apply/poller"
	"sigs.k8s.io/cli-utils/pkg/kstatus/polling"
	pollevent "sigs.k8s.io/cli-utils/pkg/kstatus/polling/event"
	"sigs.k8s.io/cli-utils/pkg/object"

	"opendev.org/airship/airshipctl/pkg/k8s/utils"
)

// FakeAdaptor is implementation of driver interface
type FakeAdaptor struct {
	initErr error
	events  []applyevent.Event
}

var _ Driver = FakeAdaptor{}

// NewFakeAdaptor returns a fake driver interface with convience methods for testing
func NewFakeAdaptor() FakeAdaptor {
	return FakeAdaptor{}
}

// Initialize implements driver
func (fa FakeAdaptor) Initialize(p poller.Poller) error {
	return fa.initErr
}

// Run implements driver
func (fa FakeAdaptor) Run(
	ctx context.Context,
	infos []*resource.Info,
	options cliapply.Options) <-chan applyevent.Event {
	ch := make(chan applyevent.Event, len(fa.events))
	defer close(ch)
	for _, e := range fa.events {
		ch <- e
	}
	return ch
}

// WithEvents set events for the adaptor
func (fa FakeAdaptor) WithEvents(events []applyevent.Event) FakeAdaptor {
	fa.events = events
	return fa
}

// WithInitError adds error to Inititialize method
func (fa FakeAdaptor) WithInitError(err error) FakeAdaptor {
	fa.initErr = err
	return fa
}

// NewFakeApplier returns applier with events you want
func NewFakeApplier(streams genericclioptions.IOStreams, events []applyevent.Event, f cmdutil.Factory) *Applier {
	return &Applier{
		Driver:                NewFakeAdaptor().WithEvents(events),
		Poller:                &FakePoller{},
		Factory:               f,
		ManifestReaderFactory: utils.DefaultManifestReaderFactory,
	}
}

// FakePoller returns a poller that sends 2 complete events for now
// TODO, make these events configurable.
type FakePoller struct {
}

// Poll implements poller interface
func (fp *FakePoller) Poll(ctx context.Context, ids []object.ObjMetadata, opts polling.Options) <-chan pollevent.Event {
	events := []pollevent.Event{
		{
			EventType: pollevent.CompletedEvent,
			Resource: &pollevent.ResourceStatus{
				Identifier: object.ObjMetadata{
					Name:      "test-rc",
					Namespace: "test",
					GroupKind: schema.GroupKind{
						Group: "v1",
						Kind:  "ReplicationController",
					},
				},
			},
		},
		{
			EventType: pollevent.CompletedEvent,
			Resource: &pollevent.ResourceStatus{
				Identifier: object.ObjMetadata{
					Name:      "airshipit-test-bundle-4bf1e4a",
					Namespace: "airshipit",
					GroupKind: schema.GroupKind{
						Group: "v1",
						Kind:  "ConfigMap",
					},
				},
			},
		},
	}
	ch := make(chan pollevent.Event, len(events))
	defer close(ch)
	for _, e := range events {
		ch <- e
	}
	return ch
}
