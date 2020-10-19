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

package resetsatoken

import (
	"fmt"
)

// ErrNoSATokenFound is returned if there are no SA tokens found in the provided namespace
type ErrNoSATokenFound struct {
	namespace string
}

// ErrNotSAToken is returned if the user input is not an SA token
type ErrNotSAToken struct {
	secretName string
}

// ErrRotateTokenFail is called when there is a failure in rotating the SA token
type ErrRotateTokenFail struct {
	Err string
}

func (e ErrNoSATokenFound) Error() string {
	return fmt.Sprintf("no service account tokens found in namespace %s", e.namespace)
}

func (e ErrNotSAToken) Error() string {
	return fmt.Sprintf("%s is not a Service Account Token", e.secretName)
}

func (e ErrRotateTokenFail) Error() string {
	return fmt.Sprintf("failed to rotate token: %s", e.Err)
}
