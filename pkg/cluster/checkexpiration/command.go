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
	"opendev.org/airship/airshipctl/pkg/log"
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

// ExpirationStore captures expiration information of all expirable entities in the cluster
type ExpirationStore struct {
	TLSSecrets []TLSSecret  `json:"tlsSecrets,omitempty" yaml:"tlsSecrets,omitempty"`
	Kubeconfs  []Kubeconfig `json:"kubeconfs,omitempty" yaml:"kubeconfs,omitempty"`
}

// TLSSecret captures expiration information of certificates embedded in TLS secrets
type TLSSecret struct {
	Name                 string            `json:"name,omitempty" yaml:"name,omitempty"`
	Namespace            string            `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	ExpiringCertificates map[string]string `json:"certificate,omitempty" yaml:"certificate,omitempty"`
}

// Kubeconfig captures expiration information of all kubeconfigs
type Kubeconfig struct {
	SecretName      string         `json:"secretName,omitempty" yaml:"secretName,omitempty"`
	SecretNamespace string         `json:"secretNamespace,omitempty" yaml:"secretNamespace,omitempty"`
	Cluster         []kubeconfData `json:"cluster,omitempty" yaml:"cluster,omitempty"`
	User            []kubeconfData `json:"user,omitempty" yaml:"user,omitempty"`
}

// kubeconfData captures cluster ca certificate expiration information and kubeconfig's user's certificate
type kubeconfData struct {
	Name            string `json:"name,omitempty" yaml:"name,omitempty"`
	CertificateName string `json:"certificateName,omitempty" yaml:"certificateName,omitempty"`
	ExpirationDate  string `json:"expirationDate,omitempty" yaml:"expirationDate,omitempty"`
}

// RunE is the implementation of check command
func (c *CheckCommand) RunE(w io.Writer) error {
	if !strings.EqualFold(c.Options.FormatType, "json") && !strings.EqualFold(c.Options.FormatType, "yaml") {
		return ErrInvalidFormat{RequestedFormat: c.Options.FormatType}
	}

	secretStore, err := NewStore(c.CfgFactory, c.ClientFactory, c.Options.Kubeconfig,
		c.Options.KubeContext, c.Options.Threshold)
	if err != nil {
		return err
	}

	expirationInfo := secretStore.GetExpiringCertificates()

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

// GetExpiringCertificates encapsulates all the different expirable entities in the cluster
func (store CertificateExpirationStore) GetExpiringCertificates() ExpirationStore {
	expiringTLSCertificates, err := store.GetExpiringTLSCertificates()
	if err != nil {
		log.Printf(err.Error())
	}

	expiringKubeConfCertificates, err := store.GetExpiringKubeConfigs()
	if err != nil {
		log.Printf(err.Error())
	}

	return ExpirationStore{
		TLSSecrets: expiringTLSCertificates,
		Kubeconfs:  expiringKubeConfCertificates,
	}
}
