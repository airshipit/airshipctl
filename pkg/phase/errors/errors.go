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

package errors

import (
	"fmt"
)

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

// ErrUnknownRenderSource returned when render command source doesn't match any known types
type ErrUnknownRenderSource struct {
	Source       string
	ValidSources []string
}

func (e ErrUnknownRenderSource) Error() string {
	return fmt.Sprintf("wrong render source '%s' specified must be one of %v",
		e.Source, e.ValidSources)
}

// ErrRenderPhaseNameNotSpecified returned when render command is called with either phase or
// executor source and phase name is not specified
type ErrRenderPhaseNameNotSpecified struct {
	Sources []string
}

func (e ErrRenderPhaseNameNotSpecified) Error() string {
	return fmt.Sprintf("must specify phase name when using %v as source",
		e.Sources)
}

// ErrInvalidPhase is returned if the phase is invalid
type ErrInvalidPhase struct {
	Reason string
}

func (e ErrInvalidPhase) Error() string {
	return fmt.Sprintf("invalid phase: %s", e.Reason)
}
