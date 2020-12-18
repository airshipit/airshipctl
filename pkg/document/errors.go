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

	"k8s.io/apimachinery/pkg/runtime"
)

// ErrDocNotFound returned if desired document not found by selector
type ErrDocNotFound struct {
	Selector Selector
}

// ErrMultiDocsFound returned if multiple documents were found by selector
type ErrMultiDocsFound struct {
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

// ErrRuntimeObjectKind returned if runtime object contains either none or
// more than one Kinds defined in schema
type ErrRuntimeObjectKind struct {
	Obj runtime.Object
}

// ErrDataNotSupplied error returned of no data in the Secret
type ErrDataNotSupplied struct {
	DocName string
	Key     string
}

// ErrDuplicateNetworkDataDocuments error returned if multiple matching documents
// were found with the same name in the same namespace
type ErrDuplicateNetworkDataDocuments struct {
	DocName   string
	Namespace string
}

// ErrBadValueFormat returned if wrong field type requested
type ErrBadValueFormat struct {
	Value    string
	Expected string
	Actual   string
}

func (e ErrDocNotFound) Error() string {
	return fmt.Sprintf("document filtered by selector %v found no documents", e.Selector)
}

func (e ErrMultiDocsFound) Error() string {
	return fmt.Sprintf("document filtered by selector %v found more than one document", e.Selector)
}

func (e ErrDocumentDataKeyNotFound) Error() string {
	return fmt.Sprintf("document %q cannot retrieve data key %q", e.DocName, e.Key)
}

func (e ErrDocumentMalformed) Error() string {
	return fmt.Sprintf("document %q is malformed: %q", e.DocName, e.Message)
}

func (e ErrRuntimeObjectKind) Error() string {
	return fmt.Sprintf("object %#v has either none or multiple kinds in scheme (expected one)", e.Obj)
}

func (e ErrDataNotSupplied) Error() string {
	return fmt.Sprintf("Document %s has no key %s", e.DocName, e.Key)
}

func (e ErrDuplicateNetworkDataDocuments) Error() string {
	return fmt.Sprintf("Found more than one document with the name %s in namespace %s", e.DocName, e.Namespace)
}

func (e ErrBadValueFormat) Error() string {
	return fmt.Sprintf("value of %s expected to have %s type, got %s", e.Value, e.Expected, e.Actual)
}
