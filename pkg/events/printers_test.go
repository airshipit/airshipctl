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
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"opendev.org/airship/airshipctl/pkg/events"
)

type fakeWriter struct {
	writeErr error
}

var _ io.Writer = fakeWriter{}

func (f fakeWriter) Write(p []byte) (n int, err error) {
	return 0, f.writeErr
}

func TestPrintEvent(t *testing.T) {
	tests := []struct {
		name          string
		formatterType string
		errString     string
		writer        io.Writer
	}{
		{
			name:          "Fail on formatter type",
			formatterType: events.YAMLPrinter,
			errString:     "Error on write",
			writer: fakeWriter{
				writeErr: fmt.Errorf("Error on write"),
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			p := events.NewGenericPrinter(tt.writer, tt.formatterType)
			e := events.NewEvent().WithIsogenEvent(events.IsogenEvent{
				Operation: events.IsogenStart,
				Message:   "starting ISO generation",
			})
			ge := events.Normalize(e)
			err := p.PrintEvent(ge)
			if tt.errString != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errString)
			}
		})
	}
}
