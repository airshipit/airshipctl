package cmd_test

import (
	"testing"

	"github.com/ian-howell/airshipadm/internal/test"
)

func TestVersion(t *testing.T) {
	tests := []test.CmdTest{{
		Name:    "version",
		Command: "version",
	}}
	test.RunCmdTests(t, tests)
}
