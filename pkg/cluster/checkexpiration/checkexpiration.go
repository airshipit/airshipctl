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
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/k8s/client"
)

// CertificateExpirationStore is the customized client store
type CertificateExpirationStore struct {
	Kclient             client.Interface
	Settings            config.Factory
	ExpirationThreshold int
}

// NewStore returns an instance of a CertificateExpirationStore
func NewStore(cfgFactory config.Factory, clientFactory client.Factory,
	kubeconfig, _ string, expirationThreshold int) (CertificateExpirationStore, error) {
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
		Kclient:             kclient,
		Settings:            cfgFactory,
		ExpirationThreshold: expirationThreshold,
	}, nil
}

// GetExpiringTLSCertificates returns the list of TLS certificates whose expiration date
// falls within the given expirationThreshold
func (store CertificateExpirationStore) GetExpiringTLSCertificates() ([]TLSSecret, error) {
	secrets, err := store.getAllTLSCertificates()
	if err != nil {
		return nil, err
	}

	tlsData := make([]TLSSecret, 0)
	for _, secret := range secrets.Items {
		expiringCertificates := store.getExpiringCertificates(secret)
		if len(expiringCertificates) > 0 {
			tlsData = append(tlsData, TLSSecret{
				Name:                 secret.Name,
				Namespace:            secret.Namespace,
				ExpiringCertificates: expiringCertificates,
			})
		}
	}
	return tlsData, nil
}

// getAllTLSCertificates juist returns all the k8s secrets with tyoe as TLS
func (store CertificateExpirationStore) getAllTLSCertificates() (*corev1.SecretList, error) {
	secretTypeFieldSelector := fmt.Sprintf("type=%s", corev1.SecretTypeTLS)
	listOptions := metav1.ListOptions{FieldSelector: secretTypeFieldSelector}
	return store.Kclient.ClientSet().CoreV1().Secrets("").List(listOptions)
}

// getExpiringCertificates skims through all the TLS certificates and returns the ones
// lesser than threshold
func (store CertificateExpirationStore) getExpiringCertificates(secret corev1.Secret) map[string]string {
	expiringCertificates := map[string]string{}
	for _, certName := range []string{corev1.TLSCertKey, corev1.ServiceAccountRootCAKey} {
		if cert, found := secret.Data[certName]; found {
			expirationDate, err := extractExpirationDateFromCertificate(cert)
			if err != nil {
				log.Printf("Unable to parse certificate for %s in secret %s in namespace %s: %v",
					certName, secret.Name, secret.Namespace, err)
				continue
			}

			if isWithinDuration(expirationDate, store.ExpirationThreshold) {
				expiringCertificates[certName] = expirationDate.String()
			}
		}
	}
	return expiringCertificates
}

// isWithinDuration checks if the certificate expirationDate is within the duration (input)
func isWithinDuration(expirationDate time.Time, duration int) bool {
	if duration < 0 {
		return true
	}
	daysUntilExpiration := int(time.Until(expirationDate).Hours() / 24)
	return 0 <= daysUntilExpiration && daysUntilExpiration < duration
}

// extractExpirationDateFromCertificate parses the certificate and returns the expiration date
func extractExpirationDateFromCertificate(certData []byte) (time.Time, error) {
	block, _ := pem.Decode(certData)
	if block == nil {
		return time.Time{}, ErrPEMFail{Context: "decode", Err: "no PEM data could be found"}
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return time.Time{}, ErrPEMFail{Context: "parse", Err: err.Error()}
	}
	return cert.NotAfter, nil
}
