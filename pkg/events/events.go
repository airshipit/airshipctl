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
)

// Event holds all possible events that can be produced by airship
type Event struct {
	Type              Type
	ApplierEvent      applyevent.Event
	ErrorEvent        ErrorEvent
	StatusPollerEvent statuspollerevent.Event
}

// ErrorEvent is produced when error is encountered
type ErrorEvent struct {
	Error error
}
