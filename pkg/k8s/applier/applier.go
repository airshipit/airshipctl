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
	"errors"
	"fmt"
	"time"

	v1 "k8s.io/api/core/v1"
	apierror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/resource"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	cliapply "sigs.k8s.io/cli-utils/pkg/apply"
	applyevent "sigs.k8s.io/cli-utils/pkg/apply/event"
	"sigs.k8s.io/cli-utils/pkg/apply/poller"
	clicommon "sigs.k8s.io/cli-utils/pkg/common"
	"sigs.k8s.io/cli-utils/pkg/inventory"
	"sigs.k8s.io/cli-utils/pkg/manifestreader"
	"sigs.k8s.io/cli-utils/pkg/provider"
	"sigs.k8s.io/yaml"

	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/events"
	"opendev.org/airship/airshipctl/pkg/k8s/utils"
	"opendev.org/airship/airshipctl/pkg/log"
)

const (
	// DefaultNamespace to store inventory objects in
	DefaultNamespace = "airshipit"
)

// Applier delivers documents to kubernetes in a declarative way
type Applier struct {
	Driver                Driver
	Factory               cmdutil.Factory
	Poller                poller.Poller
	ManifestReaderFactory utils.ManifestReaderFactory
	eventChannel          chan events.Event
}

// ReaderFactory function that returns reader factory interface
type ReaderFactory func(validate bool, bundle document.Bundle, factory cmdutil.Factory) manifestreader.ManifestReader

// NewApplier returns instance of Applier
func NewApplier(eventCh chan events.Event, f cmdutil.Factory) *Applier {
	cf := provider.NewProvider(f, inventory.WrapInventoryObj)
	return &Applier{
		Factory:               f,
		ManifestReaderFactory: utils.DefaultManifestReaderFactory,
		Driver: &Adaptor{
			CliUtilsApplier: cliapply.NewApplier(cf),
		},
		eventChannel: eventCh,
	}
}

// ApplyBundle apply bundle to kubernetes cluster
func (a *Applier) ApplyBundle(bundle document.Bundle, ao ApplyOptions) {
	defer close(a.eventChannel)
	log.Debugf("Getting infos for bundle, inventory id is %s", ao.BundleName)
	infos, err := a.getInfos(ao.BundleName, bundle, ao.DryRunStrategy)
	if err != nil {
		handleError(a.eventChannel, err)
		return
	}

	ctx := context.Background()
	ch := a.Driver.Run(ctx, infos, cliApplyOptions(ao))
	for e := range ch {
		a.eventChannel <- events.Event{
			Type:         events.ApplierType,
			ApplierEvent: e,
		}
	}
}

func (a *Applier) getInfos(
	bundleName string,
	bundle document.Bundle,
	dryRun clicommon.DryRunStrategy) ([]*resource.Info, error) {
	if bundle == nil {
		return nil, ErrNilBundle{}
	}
	selector := document.
		NewSelector().
		ByLabel(clicommon.InventoryLabel).
		ByKind(document.ConfigMapKind)
	// if we could find exactly one inventory document, we don't do anything else with it
	_, err := bundle.SelectOne(selector)
	// if we got an error, which means we could not find Config Map with inventory ID at rest
	// now we need to generate and inject one at runtime
	if err != nil && errors.As(err, &document.ErrDocNotFound{}) {
		log.Debug("Inventory Object config Map not found, auto generating Inventory object")
		invDoc, innerErr := NewInventoryDocument(bundleName)
		if innerErr != nil {
			// this should never happen
			log.Debug("Failed to create new inventory document")
			return nil, innerErr
		}
		log.Debugf("Injecting Inventory Object: %v into bundle", invDoc)
		innerErr = bundle.Append(invDoc)
		if innerErr != nil {
			log.Debug("Couldn't append bundle with inventory document")
			return nil, innerErr
		}
		log.Debugf("Making sure that inventory object namespace %s exists", invDoc.GetNamespace())
		innerErr = a.ensureNamespaceExists(invDoc.GetNamespace(), dryRun)
		if innerErr != nil {
			return nil, innerErr
		}
	} else if err != nil {
		return nil, err
	}
	if err = a.Driver.Initialize(a.Poller); err != nil {
		return nil, err
	}
	return a.ManifestReaderFactory(false, bundle, a.Factory).Read()
}

func (a *Applier) ensureNamespaceExists(name string, dryRun clicommon.DryRunStrategy) error {
	if dryRun != clicommon.DryRunNone {
		log.Printf("Dryrun strategy is specified, otherwise Namespace '%s' would have been created", name)
		return nil
	}
	clientSet, err := a.Factory.KubernetesClientSet()
	if err != nil {
		return err
	}
	nsClient := clientSet.CoreV1().Namespaces()
	_, err = nsClient.Create(&v1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	})
	if err != nil && !apierror.IsAlreadyExists(err) {
		// we got error and error doesn't say that NS already exist
		return err
	}
	return nil
}

func cliApplyOptions(ao ApplyOptions) cliapply.Options {
	var emitStatusEvents bool
	// if wait timeout is 0, we don't want to status poller to emit any events,
	// this should disable waiting for resources
	if ao.WaitTimeout != time.Duration(0) {
		emitStatusEvents = true
	}
	return cliapply.Options{
		EmitStatusEvents: emitStatusEvents,
		ReconcileTimeout: ao.WaitTimeout,
		NoPrune:          !ao.Prune,
		DryRunStrategy:   ao.DryRunStrategy,
	}
}

// Driver to cli-utils apply
type Driver interface {
	Initialize(p poller.Poller) error
	Run(ctx context.Context, infos []*resource.Info, options cliapply.Options) <-chan applyevent.Event
}

// Adaptor is implementation of driver interface
type Adaptor struct {
	CliUtilsApplier *cliapply.Applier
	Factory         cmdutil.Factory
}

// Initialize sets fake required command line flags for underlying cli-utils package
func (a *Adaptor) Initialize(p poller.Poller) error {
	err := a.CliUtilsApplier.Initialize()
	if err != nil {
		return err
	}
	// status poller needs to be set only after the Initialize method is executed
	if p != nil {
		a.CliUtilsApplier.StatusPoller = p
	}
	return nil
}

// Run perform apply operation
func (a *Adaptor) Run(ctx context.Context, infos []*resource.Info, options cliapply.Options) <-chan applyevent.Event {
	return a.CliUtilsApplier.Run(ctx, infos, options)
}

// NewInventoryDocument returns default config map with inventory Id to group up the objects
func NewInventoryDocument(inventoryID string) (document.Document, error) {
	cm := v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       document.ConfigMapKind,
			APIVersion: document.ConfigMapVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			// name here is a dummy name, cli utils won't use this name, this config map simply parsed
			// for inventory ID from label, and then ignores this config map
			Name: fmt.Sprintf("%s-%s", "airshipit", inventoryID),
			// TODO this limits us to single namespace. research passing this as parameter from
			// somewhere higher level package that uses this module
			Namespace: DefaultNamespace,
			Labels: map[string]string{
				clicommon.InventoryLabel: inventoryID,
			},
		},
		Data: map[string]string{},
	}
	b, err := yaml.Marshal(cm)
	if err != nil {
		return nil, err
	}
	return document.NewDocumentFromBytes(b)
}

func handleError(ch chan<- events.Event, err error) {
	ch <- events.Event{
		Type: events.ErrorType,
		ErrorEvent: events.ErrorEvent{
			Error: err,
		},
	}
}
