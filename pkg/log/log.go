package log

import (
	"io"
	"log"

	"github.com/ian-howell/airshipctl/pkg/environment"
)

var debug = false

// Init initializes settings related to logging
func Init(settings *environment.AirshipCTLSettings, out io.Writer) {
	debug = settings.Debug
	log.SetOutput(out)
}

// Debug is a wrapper for log.Debug
func Debug(v ...interface{}) {
	if debug {
		log.Print(v...)
	}
}

// Debugf is a wrapper for log.Debugf
func Debugf(format string, v ...interface{}) {
	if debug {
		log.Printf(format, v...)
	}
}

// Print is a wrapper for log.Print
func Print(v ...interface{}) {
	log.Print(v...)
}

// Printf is a wrapper for log.Printf
func Printf(format string, v ...interface{}) {
	log.Printf(format, v...)
}

// Fatal is a wrapper for log.Fatal
func Fatal(v ...interface{}) {
	log.Fatal(v...)
}

// Fatalf is a wrapper for log.Fatalf
func Fatalf(format string, v ...interface{}) {
	log.Fatalf(format, v...)
}
