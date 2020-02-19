package document

import (
	"fmt"
)

// ErrDocNotFound returned if desired document not found
type ErrDocNotFound struct {
	Selector Selector
}

// ErrMultipleDocsFound returned if desired document not found
type ErrMultipleDocsFound struct {
	Selector Selector
}

func (e ErrDocNotFound) Error() string {
	return fmt.Sprintf("Document filtered by selector %q found no documents", e.Selector)
}

func (e ErrMultipleDocsFound) Error() string {
	return fmt.Sprintf("Document filtered by selector %q found more than one document", e.Selector)
}
