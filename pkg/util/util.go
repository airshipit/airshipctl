package util

import (
	"os"
	"time"
)

// IsReadable returns nil if the file at path is readable. An error is returned otherwise.
func IsReadable(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err.(*os.PathError).Err
	}
	return f.Close()
}

// Clock is mostly useful for unit testing, allowing for mocking out of time
var clock *time.Time

// Now returns time.Now() if the Clock is nil. If the Clock is not nil, it is
// returned instead. This is useful for testing
func Now() time.Time {
	if clock == nil {
		return time.Now()
	}
	return *clock
}

// InitClock creates a clock for unit testing
func InitClock() {
	mockClock := time.Date(2000, time.January, 0, 0, 0, 0, 0, time.UTC)
	clock = &mockClock
}

// Clock gives access to the mocked clock
func Clock() time.Time {
	if clock == nil {
		InitClock()
	}
	return *clock
}
