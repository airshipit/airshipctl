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

package document

import (
	"fmt"
)

// ErrDocNotFound returned if desired document not found
type ErrDocNotFound struct {
	Selector Selector
}

// ErrDocumentDataKeyNotFound returned if desired key within a document not found
type ErrDocumentDataKeyNotFound struct {
	DocName string
	Key     string
}

// ErrDocumentMalformed returned if the document is structurally malformed
// (e.g. missing required low level keys)
type ErrDocumentMalformed struct {
	DocName string
	Message string
}

// ErrMultipleDocsFound returned if desired document not found
type ErrMultipleDocsFound struct {
	Selector Selector
}

func (e ErrDocNotFound) Error() string {
	return fmt.Sprintf("Document filtered by selector %v found no documents", e.Selector)
}

func (e ErrDocumentDataKeyNotFound) Error() string {
	return fmt.Sprintf("Document %q cannot retrieve data key %q", e.DocName, e.Key)
}

func (e ErrDocumentMalformed) Error() string {
	return fmt.Sprintf("Document %q is malformed: %q", e.DocName, e.Message)
}

func (e ErrMultipleDocsFound) Error() string {
	return fmt.Sprintf("Document filtered by selector %v found more than one document", e.Selector)
}
