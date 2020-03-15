package util_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"opendev.org/airship/airshipctl/pkg/util"
)

func TestNewTabWriter(t *testing.T) {
	var tests = []struct {
		testname, src, expected string
	}{
		{
			"empty-string",
			"",
			"\n",
		},
		{
			"newline-test",
			"\n",
			"\n\n",
		},
		{
			"format-string",
			"airshipctl\tconfig\tinit",
			"airshipctl   config   init\n",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.testname, func(t *testing.T) {
			var buf bytes.Buffer
			out := util.NewTabWriter(&buf)
			fmt.Fprintln(out, tt.src)
			err := out.Flush()
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, buf.String())
		})
	}
}
