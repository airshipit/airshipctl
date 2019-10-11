package yaml

import (
	"io"

	"sigs.k8s.io/yaml"
)

// WriteOut dumps any yaml competible document to writer, adding yaml separator `---`
// at the beginning of the document, and `...` at the end
func WriteOut(dst io.Writer, src interface{}) error {

	yamlOut, err := yaml.Marshal(src)
	if err != nil {
		return err
	}
	// add separator for each document
	_, err = dst.Write([]byte("---\n"))
	if err != nil {
		return err
	}
	_, err = dst.Write(yamlOut)
	if err != nil {
		return err
	}
	// add separator for each document
	_, err = dst.Write([]byte("...\n"))
	if err != nil {
		return err
	}
	return nil
}
