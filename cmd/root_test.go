package cmd_test

import (
	"testing"

	"github.com/ian-howell/airshipadm/internal/test"
)

func TestRoot(t *testing.T) {
	tests := []test.CmdTest{{
		Name:    "arishipadm root",
		Command: "",
	}}
	test.RunCmdTests(t, tests)
}
