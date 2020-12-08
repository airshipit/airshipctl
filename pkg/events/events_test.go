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

package events_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"opendev.org/airship/airshipctl/pkg/events"
)

func TestNormalize(t *testing.T) {
	tests := []struct {
		name          string
		sourceEvent   events.Event
		expectedEvent events.GenericEvent
	}{
		{
			name:        "Unknow event type",
			sourceEvent: events.NewEvent().WithErrorEvent(events.ErrorEvent{}),
			expectedEvent: events.GenericEvent{
				Type: "Unknown event type: ErrorType",
			},
		},
		{
			name: "Clusterctl event type",
			sourceEvent: events.NewEvent().WithClusterctlEvent(events.ClusterctlEvent{
				Operation: events.ClusterctlInitStart,
				Message:   "Clusterctl init start",
			}),
			expectedEvent: events.GenericEvent{
				Type:    "ClusterctlEvent",
				Message: "Clusterctl init start",
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			ge := events.Normalize(tt.sourceEvent)
			assert.Equal(t, tt.expectedEvent.Type, ge.Type)
			if tt.expectedEvent.Type != "Unknown event type: ErrorType" {
				assert.Equal(t, tt.expectedEvent.Message, ge.Message)
			}
		})
	}
}
