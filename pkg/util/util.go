package util

import (
	"os"
)

// IsReadable returns nil if the file at path is readable. An error is returned otherwise.
func IsReadable(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err.(*os.PathError).Err
	}
	return f.Close()
}
