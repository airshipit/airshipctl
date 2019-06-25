package log_test

import (
	"bytes"
	"strings"
	"testing"

	"opendev.org/airship/airshipctl/pkg/log"
)

const notEqualFmt = `Output does not match expected
GOT:      %v
Expected: %v`

func TestLoggingPrint(t *testing.T) {
	tests := []struct {
		Name     string
		Message  string
		Vals     []interface{}
		Debug    bool
		Expected string
	}{
		{
			Name:     "Print, debug set to false",
			Message:  "Print test - debug false",
			Debug:    false,
			Expected: "Print test - debug false",
		},
		{
			Name:     "Print, debug set to true",
			Message:  "Print test - debug true",
			Debug:    true,
			Expected: "Print test - debug true",
		},
		{
			Name:     "Printf, debug set to false",
			Message:  "%s - %s",
			Vals:     []interface{}{"Printf test", "debug false"},
			Debug:    false,
			Expected: "Printf test - debug false",
		},
		{
			Name:     "Printf, debug set to true",
			Message:  "%s - %s",
			Vals:     []interface{}{"Printf test", "debug true"},
			Debug:    true,
			Expected: "Printf test - debug true",
		},
	}

	for _, test := range tests {
		output := new(bytes.Buffer)
		log.Init(test.Debug, output)

		if len(test.Vals) == 0 {
			log.Print(test.Message)
		} else {
			log.Printf(test.Message, test.Vals...)
		}
		outputFields := strings.Fields(output.String())
		if len(outputFields) < 3 {
			t.Fatalf("Expected log message to have the following format: YYYY/MM/DD HH:MM:SS Message")
		}
		actual := strings.Join(outputFields[2:], " ")
		if actual != test.Expected {
			t.Errorf(notEqualFmt, actual, test.Expected)
		}
	}
}

func TestLoggingDebug(t *testing.T) {
	tests := []struct {
		Name     string
		Message  string
		Vals     []interface{}
		Debug    bool
		Expected string
	}{
		{
			Name:     "Debug, debug set to false",
			Message:  "Debug test - debug false",
			Debug:    false,
			Expected: "",
		},
		{
			Name:     "Debug, debug set to true",
			Message:  "Debug test - debug true",
			Debug:    true,
			Expected: "Debug test - debug true",
		},
		{
			Name:     "Debugf, debug set to false",
			Message:  "%s - %s",
			Vals:     []interface{}{"Debugf test", "debug false"},
			Debug:    false,
			Expected: "",
		},
		{
			Name:     "Debugf, debug set to true",
			Message:  "%s - %s",
			Vals:     []interface{}{"Debugf test", "debug true"},
			Debug:    true,
			Expected: "Debugf test - debug true",
		},
	}

	for _, test := range tests {
		output := new(bytes.Buffer)
		log.Init(test.Debug, output)

		if len(test.Vals) == 0 {
			log.Debug(test.Message)
		} else {
			log.Debugf(test.Message, test.Vals...)
		}
		var actual string
		if test.Debug {
			outputFields := strings.Fields(output.String())
			if len(outputFields) < 3 {
				t.Fatalf("Expected log message to have the following format: YYYY/MM/DD HH:MM:SS Message")
			}
			actual = strings.Join(outputFields[2:], " ")
		} else {
			actual = output.String()
		}
		if actual != test.Expected {
			t.Errorf(notEqualFmt, actual, test.Expected)
		}
	}
}
