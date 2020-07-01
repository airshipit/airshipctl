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
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"opendev.org/airship/airshipctl/pkg/events"
	"opendev.org/airship/airshipctl/pkg/k8s/utils"
	k8stest "opendev.org/airship/airshipctl/testutil/k8sutils"
)

func TestDefaultProcessor(t *testing.T) {
	proc := events.NewDefaultProcessor(utils.Streams())
	tests := []struct {
		name      string
		events    []events.Event
		errString string
	}{
		{
			name:   "success",
			events: successEvents(),
		},
		{
			name:      "error event",
			events:    errEvents(),
			errString: "somerror",
		},
		{
			name:      "error event",
			events:    errApplyEvents(),
			errString: "apply-error",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			ch := make(chan events.Event, len(tt.events))
			for _, e := range tt.events {
				ch <- e
			}
			close(ch)
			err := proc.Process(ch)
			if tt.errString != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errString)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func successEvents() []events.Event {
	applyEvents := k8stest.SuccessEvents()
	airEvents := []events.Event{}
	for _, e := range applyEvents {
		airEvents = append(airEvents, events.Event{
			Type:         events.ApplierType,
			ApplierEvent: e,
		})
	}
	return airEvents
}

func errEvents() []events.Event {
	return []events.Event{
		{
			Type: events.ErrorType,
			ErrorEvent: events.ErrorEvent{
				Error: fmt.Errorf("somerror"),
			},
		},
	}
}

func errApplyEvents() []events.Event {
	errApplyEvents := k8stest.ErrorEvents()
	airEvents := []events.Event{}
	for _, e := range errApplyEvents {
		airEvents = append(airEvents, events.Event{
			Type:         events.ApplierType,
			ApplierEvent: e,
		})
	}
	return airEvents
}
