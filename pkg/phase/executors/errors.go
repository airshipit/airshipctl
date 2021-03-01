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

package executors

import (
	"fmt"
)

// ErrUnknownExecutorAction is returned when unknown action or operation is requested
// from one of the executors.
type ErrUnknownExecutorAction struct {
	Action       string
	ExecutorName string
}

func (e ErrUnknownExecutorAction) Error() string {
	return fmt.Sprintf("unknown action type '%s' was requested from executor '%s'",
		e.Action, e.ExecutorName)
}

// ErrIsogenNilBundle is returned when isogen executor is not provided with bundle
type ErrIsogenNilBundle struct {
}

func (e ErrIsogenNilBundle) Error() string {
	return "Cannot build iso with empty bundle, no data source is available"
}

// ErrUnknownExecutorName is returned for unknown executor name parameter
// received by RegisterExecutor function
type ErrUnknownExecutorName struct {
	ExecutorName string
}

func (e ErrUnknownExecutorName) Error() string {
	return fmt.Sprintf("unknown executor name '%s'", e.ExecutorName)
}

// ErrUnableParseProvider is returned when it's impossible to parse provider's name and version
type ErrUnableParseProvider struct {
	Provider     string
	ProviderType string
}

func (e ErrUnableParseProvider) Error() string {
	return fmt.Sprintf("unable to parse name and version of '%s' type, '%s' provider", e.ProviderType, e.Provider)
}

// ErrNilExecutorDoc returned when the executor document is nil
type ErrNilExecutorDoc struct {
}

func (e ErrNilExecutorDoc) Error() string {
	return "executor document is nil"
}

// ErrInvalidPhase is returned if the phase is invalid
type ErrInvalidPhase struct {
	Reason string
}

func (e ErrInvalidPhase) Error() string {
	return fmt.Sprintf("invalid phase: %s", e.Reason)
}
