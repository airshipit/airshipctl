package yaml

import (
	"io"

	"sigs.k8s.io/yaml"
)

const (
	// DotYamlSeparator yaml separator
	DotYamlSeparator = "...\n"
	// DashYamlSeparator yaml separator
	DashYamlSeparator = "---\n"
)

// WriteOut dumps any yaml competible document to writer, adding yaml separator `---`
// at the beginning of the document, and `...` at the end
func WriteOut(dst io.Writer, src interface{}) error {

	yamlOut, err := yaml.Marshal(src)
	if err != nil {
		return err
	}
	// add separator for each document
	_, err = dst.Write([]byte(DashYamlSeparator))
	if err != nil {
		return err
	}
	_, err = dst.Write(yamlOut)
	if err != nil {
		return err
	}
	// add separator for each document
	_, err = dst.Write([]byte(DotYamlSeparator))
	if err != nil {
		return err
	}
	return nil
}
