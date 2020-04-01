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

package cloudinit

import (
	"fmt"
)

// ErrDataNotSupplied error returned of no user-data or network configuration
// in the Secret
type ErrDataNotSupplied struct {
	DocName string
	Key     string
}

// ErrDuplicateNetworkDataDocuments error returned if multiple network documents
// were found with the same name in the same namespace
type ErrDuplicateNetworkDataDocuments struct {
	DocName   string
	Namespace string
}

func (e ErrDataNotSupplied) Error() string {
	return fmt.Sprintf("Document %s has no key %s", e.DocName, e.Key)
}

func (e ErrDuplicateNetworkDataDocuments) Error() string {
	return fmt.Sprintf("Found more than one document with the name %s in namespace %s", e.DocName, e.Namespace)
}
