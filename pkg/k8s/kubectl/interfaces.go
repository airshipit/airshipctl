package kubectl

import (
	"k8s.io/kubectl/pkg/cmd/apply"

	"opendev.org/airship/airshipctl/pkg/document"
)

type Interface interface {
	Apply(docs []document.Document, ao *apply.ApplyOptions) error
}
