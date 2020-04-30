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

// Package power safely translates power information between different management clients.
package power

const (
	// StatusUnknown indicates that a baremetal host is in an unknown power state.
	StatusUnknown Status = iota
	// StatusOn indicates that a baremetal host is powered on.
	StatusOn
	// StatusOff indicates that a baremetal host is not powered on.
	StatusOff
	// StatusPoweringOn indicates that a baremetal host is powering on.
	StatusPoweringOn
	// StatusPoweringOff indicates that a baremetal host is powering off.
	StatusPoweringOff
)

// Status indicates the power status of a baremetal host e.g. on, off, unknown.
type Status int

var statusMap = [5]string{
	"UNKNOWN",
	"ON",
	"OFF",
	"POWERING ON",
	"POWERING OFF",
}

// String provides a human-readable string value for a power status.
func (s Status) String() string {
	return statusMap[s]
}
