package render

import (
	"opendev.org/airship/airshipctl/pkg/environment"
)

// Settings for document rendering
type Settings struct {
	*environment.AirshipCTLSettings
	// Label filter documents by label string
	Label []string
	// Annotation filter documents by annotation string
	Annotation []string
	// GroupVersion filter documents by API version
	GroupVersion []string
	// Kind filter documents by document kind
	Kind []string
	// RawFilter contains logical expression to filter documents
	RawFilter string
}
