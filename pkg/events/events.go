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
)

// Event holds all possible events that can be produced by airship
type Event struct {
	Type              Type
	ApplierEvent      applyevent.Event
	ErrorEvent        ErrorEvent
	StatusPollerEvent statuspollerevent.Event
	ClusterctlEvent   ClusterctlEvent
	IsogenEvent       IsogenEvent
}

// ErrorEvent is produced when error is encountered
type ErrorEvent struct {
	Error error
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

// IsogenOperation type
type IsogenOperation int

const (
	// IsogenStart operation
	IsogenStart IsogenOperation = iota
	// IsogenValidation opearation
	IsogenValidation
	// IsogenEnd operation
	IsogenEnd
)

// IsogenEvent needs to to track events in isogen executor
type IsogenEvent struct {
	Operation IsogenOperation
	Message   string
}
