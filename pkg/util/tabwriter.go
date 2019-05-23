package util

import (
	"io"
	"text/tabwriter"
)

const (
	tabwriterMinWidth = 6
	tabwriterWidth    = 4
	tabwriterPadding  = 3
	tabwriterPadChar  = ' '
	tabwriterFlags    = 0
)

// NewTabWriter returns a tabwriter that translates tabbed columns in input into properly aligned text.
func NewTabWriter(out io.Writer) *tabwriter.Writer {
	return tabwriter.NewWriter(out, tabwriterMinWidth, tabwriterWidth, tabwriterPadding, tabwriterPadChar, tabwriterFlags)
}
