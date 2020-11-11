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

package checkexpiration

import (
	"encoding/json"
	"io"
	"strings"

	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/k8s/client"
	"opendev.org/airship/airshipctl/pkg/util/yaml"
)

// CheckFlags flags given for checking the expiration
type CheckFlags struct {
	Threshold   int
	FormatType  string
	Kubeconfig  string
	KubeContext string
}

// CheckCommand check expiration command
type CheckCommand struct {
	Options       CheckFlags
	CfgFactory    config.Factory
	ClientFactory client.Factory
}

// RunE is the implementation of check command
func (c *CheckCommand) RunE(w io.Writer) error {
	if !strings.EqualFold(c.Options.FormatType, "json") && !strings.EqualFold(c.Options.FormatType, "yaml") {
		return ErrInvalidFormat{RequestedFormat: c.Options.FormatType}
	}

	secretStore, err := NewStore(c.CfgFactory, c.ClientFactory, c.Options.Kubeconfig, c.Options.KubeContext)
	if err != nil {
		return err
	}

	expirationInfo, err := secretStore.GetExpiringCertificates(c.Options.Threshold)
	if err != nil {
		return err
	}
	if c.Options.FormatType == "yaml" {
		err = yaml.WriteOut(w, expirationInfo)
		if err != nil {
			return err
		}
	} else {
		buffer, err := json.MarshalIndent(expirationInfo, "", "    ")
		if err != nil {
			return err
		}
		_, err = w.Write(buffer)
		if err != nil {
			return err
		}
	}
	return nil
}
