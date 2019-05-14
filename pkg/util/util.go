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

// ReadDir does the same thing as ioutil.ReadDir, but it doesn't sort the files.
func ReadDir(dirname string) ([]os.FileInfo, error) {
	f, err := os.Open(dirname)
	if err != nil {
		return []os.FileInfo{}, err
	}
	defer f.Close()
	return f.Readdir(-1)
}
