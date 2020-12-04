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
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"

	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/k8s/client"
)

const (
	kubeconfigIdentifierSuffix   = "-kubeconfig"
	timeFormat                   = "Jan 02, 2006 15:04 MST"
	nodeCertExpirationAnnotation = "cert-expiration"
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
	return store.getSecrets(listOptions)
}

// getSecrets returns the secret list based on the listOptions
func (store CertificateExpirationStore) getSecrets(listOptions metav1.ListOptions) (*corev1.SecretList, error) {
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

// GetExpiringKubeConfigs - fetches all the '-kubeconfig' secrets and identifies expiration
func (store CertificateExpirationStore) GetExpiringKubeConfigs() ([]Kubeconfig, error) {
	kubeconfigs, err := store.getKubeconfSecrets()

	if err != nil {
		return nil, err
	}

	kSecretData := make([]Kubeconfig, 0)

	for _, kubeconfig := range kubeconfigs {
		kubecontent, err := clientcmd.Load(kubeconfig.Data["value"])
		if err != nil {
			log.Printf("Failed to read kubeconfig from %s in %s-"+
				"it maybe malformed : %v", kubeconfig.Name, kubeconfig.Namespace, err.Error())
			continue
		}

		expiringClusters := store.getExpiringClusterCertificates(kubecontent)

		expiringUsers := store.getExpiringUserCertificates(kubecontent)

		if len(expiringClusters) > 0 || len(expiringUsers) > 0 {
			kSecretData = append(kSecretData, Kubeconfig{
				SecretName:      kubeconfig.Name,
				SecretNamespace: kubeconfig.Namespace,
				Cluster:         expiringClusters,
				User:            expiringUsers,
			})
		}
	}
	return kSecretData, nil
}

// filterKubeConfigs identifies the kubeconfig secrets based on the kubeconfigIdentifierSuffix
func filterKubeConfigs(secrets []corev1.Secret) []corev1.Secret {
	filteredSecrets := []corev1.Secret{}
	for _, secret := range secrets {
		if strings.HasSuffix(secret.Name, kubeconfigIdentifierSuffix) {
			filteredSecrets = append(filteredSecrets, secret)
		}
	}
	return filteredSecrets
}

// filterOwners allows only the secrets with Ownerreferences matching ownerKind
func filterOwners(secrets []corev1.Secret, ownerKind string) []corev1.Secret {
	filteredSecrets := []corev1.Secret{}
	for _, secret := range secrets {
		for _, ownerRef := range secret.OwnerReferences {
			if ownerRef.Kind == ownerKind {
				filteredSecrets = append(filteredSecrets, secret)
			}
		}
	}
	return filteredSecrets
}

func (store CertificateExpirationStore) getExpiringClusterCertificates(
	kubeconfig *clientcmdapi.Config) []kubeconfData {
	expiringClusterCertificates := make([]kubeconfData, 0)

	// Iterate through each Cluster and identify expiration
	for clusterName, clusterData := range kubeconfig.Clusters {
		expirationDate, err := extractExpirationDateFromCertificate(clusterData.CertificateAuthorityData)
		if err != nil {
			log.Printf("Unable to parse certificate for %s : %v", clusterName, err)
			continue
		}

		if isWithinDuration(expirationDate, store.ExpirationThreshold) {
			expiringClusterCertificates = append(expiringClusterCertificates, kubeconfData{
				Name:            clusterName,
				CertificateName: "CertificateAuthorityData",
				ExpirationDate:  expirationDate.String(),
			})
		}
	}
	return expiringClusterCertificates
}

func (store CertificateExpirationStore) getExpiringUserCertificates(
	kubeconfig *clientcmdapi.Config) []kubeconfData {
	expiringUserCertificates := make([]kubeconfData, 0)

	// Iterate through each User and identify expiration
	for userName, userData := range kubeconfig.AuthInfos {
		expirationDate, err := extractExpirationDateFromCertificate(userData.ClientCertificateData)
		if err != nil {
			log.Printf("Unable to parse certificate for %s : %v", userName, err)
			continue
		}

		if isWithinDuration(expirationDate, store.ExpirationThreshold) {
			expiringUserCertificates = append(expiringUserCertificates, kubeconfData{
				Name:            userName,
				CertificateName: "ClientCertificateData",
				ExpirationDate:  expirationDate.String(),
			})
		}
	}
	return expiringUserCertificates
}

// getKubeconfSecrets filters the kubeconf secrets
func (store CertificateExpirationStore) getKubeconfSecrets() ([]corev1.Secret, error) {
	secrets, err := store.getSecrets(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	kubeconfigs := filterKubeConfigs(secrets.Items)
	kubeconfigs = filterOwners(kubeconfigs, "KubeadmControlPlane")
	return kubeconfigs, nil
}

// GetExpiringNodeCertificates runs through all the nodes and identifies expiration
func (store CertificateExpirationStore) GetExpiringNodeCertificates() ([]NodeCert, error) {
	// Node will be updated with an annotation with the expiry content (Activity
	// of HostConfig Operator - 'check-expiry' CR Object) every day (Cron like
	// activity is performed by reconcile tag in the Operator) Below code is
	// implemented to just read the annotation, parse it, identify expirable
	// content and report back

	// Expected Annotation Format:
	// "cert-expiration": "{ admin.conf: Aug 06, 2021 12:36 UTC },
	// 					   { apiserver: Aug 06, 2021 12:36 UTC },
	//                     { apiserver-etcd-client: Aug 06, 2021 12:36 UTC },
	//                     { apiserver-kubelet-client: Aug 06, 2021 12:36 UTC },
	//                     { controller-manager.conf: Aug 06, 2021 12:36 UTC },
	//                     { etcd-healthcheck-client: Aug 06, 2021 12:36 UTC },
	//                     { etcd-peer: Aug 06, 2021 12:36 UTC },
	//                     { etcd-server: Aug 06, 2021 12:36 UTC },
	//                     { front-proxy-client: Aug 06, 2021 12:36 UTC },
	//                     { scheduler.conf: Aug 06, 2021 12:36 UTC },
	//                     { ca: Aug 04, 2030 12:36 UTC },
	//                     { etcd-ca: Aug 04, 2030 12:36 UTC },
	//                     { front-proxy-ca: Aug 04, 2030 12:36 UTC }"

	nodes, err := store.getNodes(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	nodeData := make([]NodeCert, 0)
	for _, node := range nodes.Items {
		expiringNodeCertificates := store.getExpiringNodeCertificates(node)
		if len(expiringNodeCertificates) > 0 {
			nodeData = append(nodeData, NodeCert{
				Name:                 node.Name,
				Namespace:            node.Namespace,
				ExpiringCertificates: expiringNodeCertificates,
			})
		}
	}
	return nodeData, nil
}

// getSecrets returns the Nodes list based on the listOptions
func (store CertificateExpirationStore) getNodes(listOptions metav1.ListOptions) (*corev1.NodeList, error) {
	return store.Kclient.ClientSet().CoreV1().Nodes().List(listOptions)
}

// getExpiringNodeCertificates skims through all the node certificates and returns
// the ones lesser than threshold
func (store CertificateExpirationStore) getExpiringNodeCertificates(node corev1.Node) map[string]string {
	if cert, found := node.ObjectMeta.Annotations[nodeCertExpirationAnnotation]; found {
		certificateList := splitAsList(cert)

		expiringCertificates := map[string]string{}
		for _, certificate := range certificateList {
			certificateName, expirationDate := identifyCertificateNameAndExpirationDate(certificate)
			if certificateName != "" && isWithinDuration(expirationDate, store.ExpirationThreshold) {
				expiringCertificates[certificateName] = expirationDate.String()
			}
		}
		return expiringCertificates
	}
	log.Printf("%s annotation missing for node %s in %s", nodeCertExpirationAnnotation,
		node.Name, node.Namespace)
	return nil
}

// splitAsList performes the required string manipulations and returns list of items
func splitAsList(value string) []string {
	return strings.Split(strings.ReplaceAll(value, "{", ""), "},")
}

// identifyCertificateNameAndExpirationDate performs string manipulations and returns
// certificate name and its expiration date
func identifyCertificateNameAndExpirationDate(certificate string) (string, time.Time) {
	certificateName := strings.TrimSpace(strings.Split(certificate, ":")[0])
	expirationDate := strings.TrimSpace(strings.Split(certificate, ":")[1]) +
		":" +
		strings.TrimSpace(strings.ReplaceAll(strings.Split(certificate, ":")[2], "}", ""))

	formattedExpirationDate, err := time.Parse(timeFormat, expirationDate)
	if err != nil {
		log.Printf(err.Error())
		return "", time.Time{}
	}
	return certificateName, formattedExpirationDate
}
