package log

import (
	"io"
	"log"
	"os"
)

var (
	debug      = false
	airshipLog = log.New(os.Stderr, "", log.LstdFlags)
)

// Init initializes settings related to logging
func Init(debugFlag bool, out io.Writer) {
	debug = debugFlag
	airshipLog.SetOutput(out)
}

// Debug is a wrapper for log.Debug
func Debug(v ...interface{}) {
	if debug {
		airshipLog.Print(v...)
	}
}

// Debugf is a wrapper for log.Debugf
func Debugf(format string, v ...interface{}) {
	if debug {
		airshipLog.Printf(format, v...)
	}
}

// Print is a wrapper for log.Print
func Print(v ...interface{}) {
	airshipLog.Print(v...)
}

// Printf is a wrapper for log.Printf
func Printf(format string, v ...interface{}) {
	airshipLog.Printf(format, v...)
}

// Fatal is a wrapper for log.Fatal
func Fatal(v ...interface{}) {
	airshipLog.Fatal(v...)
}

// Fatalf is a wrapper for log.Fatalf
func Fatalf(format string, v ...interface{}) {
	airshipLog.Fatalf(format, v...)
}

// Writer returns log output writer object
func Writer() io.Writer {
	return airshipLog.Writer()
}
