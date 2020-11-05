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
	"time"

	applyevent "sigs.k8s.io/cli-utils/pkg/apply/event"
	statuspollerevent "sigs.k8s.io/cli-utils/pkg/kstatus/polling/event"
)

// Type indicates type of the event
type Type int

const (
	// ApplierType is event that are produced by applier
	ApplierType Type = iota
	// ErrorType means that even is of error type
	ErrorType
	// StatusPollerType event produced by status poller
	StatusPollerType
	// WaitType is event emitted when airshipctl is waiting for something
	WaitType
	// ClusterctlType event emitted by Clusterctl executor
	ClusterctlType
	// IsogenType event emitted by Isogen executor
	IsogenType
	// BootstrapType event emitted by Bootstrap executor
	BootstrapType
	//GenericContainerType event emitted by GenericContainer
	GenericContainerType
)

// Event holds all possible events that can be produced by airship
type Event struct {
	Type                  Type
	Timestamp             time.Time
	ApplierEvent          applyevent.Event
	ErrorEvent            ErrorEvent
	StatusPollerEvent     statuspollerevent.Event
	ClusterctlEvent       ClusterctlEvent
	IsogenEvent           IsogenEvent
	BootstrapEvent        BootstrapEvent
	GenericContainerEvent GenericContainerEvent
}

// NewEvent create new event with timestamp
func NewEvent() Event {
	return Event{
		Timestamp: time.Now(),
	}
}

// ErrorEvent is produced when error is encountered
type ErrorEvent struct {
	Error error
}

// WithErrorEvent sets type and actual error event
func (e Event) WithErrorEvent(event ErrorEvent) Event {
	e.Type = ErrorType
	e.ErrorEvent = event
	return e
}

// ClusterctlOperation type
type ClusterctlOperation int

const (
	// ClusterctlInitStart operation
	ClusterctlInitStart ClusterctlOperation = iota
	// ClusterctlInitEnd operation
	ClusterctlInitEnd
	// ClusterctlMoveStart operation
	ClusterctlMoveStart
	// ClusterctlMoveEnd operation
	ClusterctlMoveEnd
)

// ClusterctlEvent is produced by clusterctl executor
type ClusterctlEvent struct {
	Operation ClusterctlOperation
	Message   string
}

// WithClusterctlEvent sets type and actual clusterctl event
func (e Event) WithClusterctlEvent(concreteEvent ClusterctlEvent) Event {
	e.Type = ClusterctlType
	e.ClusterctlEvent = concreteEvent
	return e
}

// IsogenOperation type
type IsogenOperation int

const (
	// IsogenStart operation
	IsogenStart IsogenOperation = iota
	// IsogenValidation operation
	IsogenValidation
	// IsogenEnd operation
	IsogenEnd
)

// IsogenEvent needs to to track events in isogen executor
type IsogenEvent struct {
	Operation IsogenOperation
	Message   string
}

// WithIsogenEvent sets type and actual isogen event
func (e Event) WithIsogenEvent(concreteEvent IsogenEvent) Event {
	e.Type = IsogenType
	e.IsogenEvent = concreteEvent
	return e
}

// BootstrapOperation type
type BootstrapOperation int

const (
	// BootstrapStart operation
	BootstrapStart BootstrapOperation = iota
	// BootstrapDryRun operation
	BootstrapDryRun
	// BootstrapValidation operation
	BootstrapValidation
	// BootstrapRun operation
	BootstrapRun
	// BootstrapEnd operation
	BootstrapEnd
)

// BootstrapEvent needs to to track events in bootstrap executor
type BootstrapEvent struct {
	Operation BootstrapOperation
	Message   string
}

// WithBootstrapEvent sets type and actual bootstrap event
func (e Event) WithBootstrapEvent(concreteEvent BootstrapEvent) Event {
	e.Type = BootstrapType
	e.BootstrapEvent = concreteEvent
	return e
}

// GenericContainerOperation type
type GenericContainerOperation int

const (
	// GenericContainerStart operation
	GenericContainerStart GenericContainerOperation = iota
	// GenericContainerStop operation
	GenericContainerStop
)

// GenericContainerEvent needs to to track events in GenericContainer executor
type GenericContainerEvent struct {
	Operation GenericContainerOperation
	Message   string
}

// WithGenericContainerEvent sets type and actual GenericContainer event
func (e Event) WithGenericContainerEvent(concreteEvent GenericContainerEvent) Event {
	e.Type = GenericContainerType
	e.GenericContainerEvent = concreteEvent
	return e
}
