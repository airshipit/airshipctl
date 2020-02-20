package kubectl

import (
	"opendev.org/airship/airshipctl/pkg/document"
)

type Interface interface {
	Apply(docs []document.Document, ao *ApplyOptions) error
	ApplyOptions() (*ApplyOptions, error)
}
