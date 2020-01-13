package util

import (
	"io/ioutil"
	"os"
)

// WriteFiles write multiple files described in a map
func WriteFiles(fls map[string][]byte, mode os.FileMode) error {
	for fileName, data := range fls {
		if err := ioutil.WriteFile(fileName, data, mode); err != nil {
			return err
		}
	}
	return nil
}
