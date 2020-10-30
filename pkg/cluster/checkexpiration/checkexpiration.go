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
	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/errors"
	"opendev.org/airship/airshipctl/pkg/k8s/client"
)

// CertificateExpirationStore is the customized client store
type CertificateExpirationStore struct {
	Kclient  client.Interface
	Settings config.Factory
}

// Expiration captures expiration information of all expirable entities in the cluster
type Expiration struct{}

// NewStore returns an instance of a CertificateExpirationStore
func NewStore(cfgFactory config.Factory, clientFactory client.Factory,
	kubeconfig string, _ string) (CertificateExpirationStore, error) {
	airshipconfig, err := cfgFactory()
	if err != nil {
		return CertificateExpirationStore{}, err
	}

	// TODO (guhan) Allow kube context to be passed to client Factory
	// 4th argument in NewStore takes kube context and is ignored for now.
	// To be modified post #388. Refer to
	// https://review.opendev.org/#/c/760501/7/pkg/cluster/checkexpiration/command.go@31
	kclient, err := clientFactory(airshipconfig.LoadedConfigPath(), kubeconfig)
	if err != nil {
		return CertificateExpirationStore{}, err
	}

	return CertificateExpirationStore{
		Kclient:  kclient,
		Settings: cfgFactory,
	}, nil
}

// GetExpiringCertificates checks for the expiration data
// NOT IMPLEMENTED (guhan)
// TODO (guhan) check for TLS certificates, workload kubeconfig and node certificates
func (store CertificateExpirationStore) GetExpiringCertificates(expirationThreshold int) (Expiration, error) {
	return Expiration{}, errors.ErrNotImplemented{What: "check certificate expiration logic"}
}
