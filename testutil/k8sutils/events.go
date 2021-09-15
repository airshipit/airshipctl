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

package k8sutils

import (
	"opendev.org/airship/airshipctl/pkg/events"
)

// SuccessEvents returns list of events that constitute a successful cli utils apply
func SuccessEvents() []events.ApplierEvent {
	return []events.ApplierEvent{
		{
			Operation: "init",
			Message:   "success",
		},
		{
			Operation: "configured",
			Message:   "success",
		},
		{
			Operation: "completed",
			Message:   "success",
		},
	}
}

// ErrorEvents return a list of events with error
func ErrorEvents() []events.ApplierEvent {
	return []events.ApplierEvent{
		{
			Operation: "error",
			Message:   "apply-error",
		},
	}
}
