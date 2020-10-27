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

package phase

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

// ErrExecutorNotFound is returned if phase executor was not found in executor
// registry map
type ErrExecutorNotFound struct {
	GVK schema.GroupVersionKind
}

func (e ErrExecutorNotFound) Error() string {
	return fmt.Sprintf("executor identified by '%s' is not found", e.GVK)
}

// ErrExecutorRegistration is a wrapper for executor registration errors
type ErrExecutorRegistration struct {
	ExecutorName string
	Err          error
}

func (e ErrExecutorRegistration) Error() string {
	return fmt.Sprintf("failed to register executor %s, registration function returned %s", e.ExecutorName, e.Err.Error())
}

// ErrDocumentEntrypointNotDefined returned when phase has no entrypoint defined and phase needs it
type ErrDocumentEntrypointNotDefined struct {
	PhaseName      string
	PhaseNamespace string
}

func (e ErrDocumentEntrypointNotDefined) Error() string {
	return fmt.Sprintf("documentEntryPoint is not defined for the phase '%s' in namespace '%s'",
		e.PhaseName,
		e.PhaseNamespace)
}

// ErrExecutorRefNotDefined returned when phase has no entrypoint defined and phase needs it
type ErrExecutorRefNotDefined struct {
	PhaseName      string
	PhaseNamespace string
}

func (e ErrExecutorRefNotDefined) Error() string {
	return fmt.Sprintf("Phase name '%s', namespace '%s' must have executorRef field defined in config",
		e.PhaseName,
		e.PhaseNamespace)
}
