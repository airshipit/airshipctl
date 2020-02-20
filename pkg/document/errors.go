package document

import (
	"fmt"
)

// ErrDocNotFound returned if desired document not found
type ErrDocNotFound struct {
	Selector string
	Kind     string
}

func (e ErrDocNotFound) Error() string {
	return fmt.Sprintf("Document filtered by selector %s with Kind %s not found", e.Selector, e.Kind)
}
