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

package kubeconfig

import "fmt"

// ErrAllSourcesFailed returned when kubeconfig path is not specified
type ErrAllSourcesFailed struct {
	ClusterName string
}

func (e *ErrAllSourcesFailed) Error() string {
	return fmt.Sprintf("all kubeconfig sources failed for cluster '%s'", e.ClusterName)
}

// ErrKubeconfigMergeFailed is returned when builder doesn't know which context to merge
type ErrKubeconfigMergeFailed struct {
	Message string
}

func (e *ErrKubeconfigMergeFailed) Error() string {
	return fmt.Sprintf("failed merging kubeconfig: %s", e.Message)
}

// IsErrAllSourcesFailedErr returns true if error is of type ErrAllSourcesFailedErr
func IsErrAllSourcesFailedErr(err error) bool {
	_, ok := err.(*ErrAllSourcesFailed)
	return ok
}

// ErrUknownKubeconfigSourceType returned type of kubeconfig source is unknown
type ErrUknownKubeconfigSourceType struct {
	Type string
}

func (e *ErrUknownKubeconfigSourceType) Error() string {
	return fmt.Sprintf("unknown source type %s", e.Type)
}
