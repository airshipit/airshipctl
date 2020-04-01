/*
 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     https://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

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

// WriteOut dumps any yaml compatible document to writer, adding yaml separator `---`
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
