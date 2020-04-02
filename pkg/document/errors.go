package document

import (
	"fmt"
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
