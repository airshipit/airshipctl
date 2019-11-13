package document

import (
	"fmt"
)

// ErrDocNotFound returned if desired document not found
type ErrDocNotFound struct {
	Annotation string
	Kind       string
}

func (e ErrDocNotFound) Error() string {
	return fmt.Sprintf("Document annotated by %s with Kind %s not found", e.Annotation, e.Kind)
}
