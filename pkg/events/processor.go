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

package events

import (
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"sigs.k8s.io/cli-utils/cmd/printers"
	applyevent "sigs.k8s.io/cli-utils/pkg/apply/event"

	"opendev.org/airship/airshipctl/pkg/log"
)

// EventProcessor use to process event channels produced by executors
type EventProcessor interface {
	Process(<-chan Event) error
}

// DefaultProcessor is implementation of EventProcessor
type DefaultProcessor struct {
	errors      []error
	applierChan chan<- applyevent.Event
}

// NewDefaultProcessor returns instance of DefaultProcessor as interface Implementation
func NewDefaultProcessor(streams genericclioptions.IOStreams) EventProcessor {
	applyCh := make(chan applyevent.Event)
	go printers.GetPrinter(printers.EventsPrinter, streams).Print(applyCh, false)
	return &DefaultProcessor{
		errors:      []error{},
		applierChan: applyCh,
	}
}

// Process is implementation of EventProcessor
func (p *DefaultProcessor) Process(ch <-chan Event) error {
	for e := range ch {
		switch e.Type {
		case ApplierType:
			p.processApplierEvent(e.ApplierEvent)
		case ErrorType:
			p.errors = append(p.errors, e.ErrorEvent.Error)
		case StatusPollerType:
			log.Fatalf("Processing for status poller events are not yet implemented")
		default:
			log.Fatalf("Unknown event type received: %d", e.Type)
		}
	}
	return checkErrors(p.errors)
}

func (p *DefaultProcessor) processApplierEvent(e applyevent.Event) {
	if e.Type == applyevent.ErrorType {
		log.Printf("Received error when applying errors to kubernetes %v", e.ErrorEvent.Err)
		p.errors = append(p.errors, e.ErrorEvent.Err)
		// Don't write errors events to applier channel, as it will use os.Exit to stop program execution
		return
	}
	p.applierChan <- e
}

// Check list of errors, and verify that these errors we are able to tolerate
// currently we simply check if the list is empty or not
func checkErrors(errs []error) error {
	if len(errs) != 0 {
		return ErrEventReceived{
			errors: errs,
		}
	}
	return nil
}
