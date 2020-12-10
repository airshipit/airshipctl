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

package events

import (
	"encoding/json"
	"io"

	"sigs.k8s.io/yaml"

	"opendev.org/airship/airshipctl/pkg/log"
)

const (
	// YAMLPrinter format for event printer output
	YAMLPrinter = "yaml"
	// JSONPrinter format for event printer output
	JSONPrinter = "json"
)

// NewGenericPrinter returns event printer
func NewGenericPrinter(writer io.Writer, formatterType string) GenericPrinter {
	var formatter func(o interface{}) ([]byte, error)
	switch formatterType {
	case YAMLPrinter:
		formatter = yaml.Marshal
	case JSONPrinter:
		formatter = json.Marshal
	default:
		log.Fatal("Event printer received wrong type of event formatter")
	}
	return GenericPrinter{
		formatter: formatter,
		writer:    writer,
	}
}

// GenericPrinter object represents event printer
type GenericPrinter struct {
	formatter func(interface{}) ([]byte, error)
	writer    io.Writer
}

// PrintEvent write event details
func (p GenericPrinter) PrintEvent(ge GenericEvent) error {
	data, err := p.formatter(map[string]interface{}{
		"Type":      ge.Type,
		"Operation": ge.Operation,
		"Message":   ge.Message,
		"Timestamp": ge.Timestamp,
	})
	if err != nil {
		return err
	}
	data = append(data, '\n')
	_, err = p.writer.Write(data)
	return err
}
