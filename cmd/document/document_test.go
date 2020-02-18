package document_test

import (
	"testing"

	"opendev.org/airship/airshipctl/cmd/document"
	"opendev.org/airship/airshipctl/testutil"
)

func TestDocument(t *testing.T) {
	tests := []*testutil.CmdTest{
		{
			Name:    "document-with-defaults",
			CmdLine: "",
			Cmd:     document.NewDocumentCommand(nil),
		},
		{
			Name:    "document-render-with-help",
			CmdLine: "-h",
			Cmd:     document.NewRenderCommand(nil),
		},
	}
	for _, tt := range tests {
		testutil.RunTest(t, tt)
	}
}
