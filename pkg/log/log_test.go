package log_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/ian-howell/airshipadm/pkg/environment"
	"github.com/ian-howell/airshipadm/pkg/log"
)

const notEqualFmt = `Output does not match expected
GOT:      %v
Expected: %v`

func TestLoggingWithoutDebug(t *testing.T) {
	settings := environment.AirshipADMSettings{
		Debug: false,
	}

	output := new(bytes.Buffer)

	log.Init(&settings, output)

	log.Print("Print test - debug false")
	expected := "Print test - debug false"
	outputFields := strings.Fields(output.String())
	if len(outputFields) < 3 {
		t.Fatalf("Expected log message to have the following format: YYYY/MM/DD HH:MM:SS Message")
	}
	outputMessage := strings.Join(outputFields[2:], " ")
	if outputMessage != expected {
		t.Errorf(notEqualFmt, outputMessage, expected)
	}

	output.Reset()

	log.Printf("%s - %s", "Printf test", "debug false")
	expected = "Printf test - debug false"
	outputFields = strings.Fields(output.String())
	if len(outputFields) < 3 {
		t.Fatalf("Expected log message to have the following format: YYYY/MM/DD HH:MM:SS Message")
	}
	outputMessage = strings.Join(outputFields[2:], " ")
	if outputMessage != expected {
		t.Errorf(notEqualFmt, outputMessage, expected)
	}

	output.Reset()
	log.Debug("Debug test - debug false")
	log.Debugf("%s - %s", "Debugf test", "debug false")
	if len(output.Bytes()) > 0 {
		t.Errorf("Unexpected output: %s", output)
	}
}

func TestLoggingWithDebug(t *testing.T) {
	settings := environment.AirshipADMSettings{
		Debug: true,
	}

	output := new(bytes.Buffer)

	log.Init(&settings, output)

	log.Debug("Debug test - debug true")
	expected := "Debug test - debug true"
	outputFields := strings.Fields(output.String())
	if len(outputFields) < 3 {
		t.Fatalf("Expected log message to have the following format: YYYY/MM/DD HH:MM:SS Message")
	}
	outputMessage := strings.Join(outputFields[2:], " ")
	if outputMessage != expected {
		t.Errorf(notEqualFmt, outputMessage, expected)
	}

	output.Reset()

	log.Debugf("%s - %s", "Debugf test", "debug true")
	expected = "Debugf test - debug true"
	outputFields = strings.Fields(output.String())
	if len(outputFields) < 3 {
		t.Fatalf("Expected log message to have the following format: YYYY/MM/DD HH:MM:SS Message")
	}
	outputMessage = strings.Join(outputFields[2:], " ")
	if outputMessage != expected {
		t.Errorf(notEqualFmt, outputMessage, expected)
	}
}
