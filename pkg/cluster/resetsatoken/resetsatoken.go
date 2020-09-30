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

package resetsatoken

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"opendev.org/airship/airshipctl/pkg/log"
)

const (
	replicaSetKind = "ReplicaSet"
)

// TokenManager manages service account rotation
type TokenManager struct {
	kclient kubernetes.Interface
}

// NewTokenManager returns an instance of a TokenManager
func NewTokenManager(kclient kubernetes.Interface) (*TokenManager, error) {
	return &TokenManager{
		kclient: kclient,
	}, nil
}

// RotateToken - rotates token > 1. Deletes the secret and 2. Deletes its pod
// Deleting the SA Secret recreates a new secret with a new token information
// However, the pods referencing to the old secret needs to be refreshed
// manually and hence deleting the pod to allow it to get recreated with new
// secret reference
func (manager TokenManager) RotateToken(ns string, secretName string) error {
	if secretName == "" {
		return manager.rotateAllTokens(ns)
	}
	return manager.rotateSingleToken(ns, secretName)
}

// deleteSecret- deletes the secret
func (manager TokenManager) deleteSecret(secretName string, ns string) error {
	return manager.kclient.CoreV1().Secrets(ns).Delete(secretName, &metav1.DeleteOptions{})
}

// deletePod - identifies the secret relationship with pods and deletes corresponding pods
// if its part of replicaset
func (manager TokenManager) deletePod(secretName string, ns string) error {
	pods, err := manager.kclient.CoreV1().Pods(ns).List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, pod := range pods.Items {
		for _, volume := range pod.Spec.Volumes {
			if volume.Name == secretName {
				if manager.isReplicaSet(pod.OwnerReferences) {
					log.Printf("Deleting pod - %s in %s", pod.Name, ns)
					if deleteErr := manager.kclient.CoreV1().Pods(ns).Delete(pod.Name,
						&metav1.DeleteOptions{}); deleteErr != nil {
						log.Printf("Failed to delete pod: %v", err.Error())
					}
				}
			}
		}
	}
	return nil
}

// rotateAllTokens rotates all the tokens in the given namespace
func (manager TokenManager) rotateAllTokens(ns string) error {
	tokenTypeFieldSelector := fmt.Sprintf("type=%s", corev1.SecretTypeServiceAccountToken)
	listOptions := metav1.ListOptions{FieldSelector: tokenTypeFieldSelector}

	secrets, err := manager.kclient.CoreV1().Secrets(ns).List(listOptions)
	if err != nil {
		return err
	}

	if len(secrets.Items) == 0 {
		return ErrNoSATokenFound{namespace: ns}
	}

	for _, secret := range secrets.Items {
		err := manager.rotate(secret.Name, secret.Namespace)
		if err != nil {
			return err
		}
	}
	return nil
}

// rotateSingleToken rotates a given token in the given ns
func (manager TokenManager) rotateSingleToken(ns string, secretName string) error {
	secret, err := manager.kclient.CoreV1().Secrets(ns).Get(secretName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if secret.Type != corev1.SecretTypeServiceAccountToken {
		return ErrNotSAToken{secretName: secretName}
	}
	return manager.rotate(secretName, ns)
}

// rotate performs delete action for secrets and its pods
func (manager TokenManager) rotate(secretName string, secretNamespace string) error {
	log.Printf("Rotating token - %s in %s", secretName, secretNamespace)
	err := manager.deleteSecret(secretName, secretNamespace)
	if err != nil {
		return err
	}

	return manager.deletePod(secretName, secretNamespace)
}

// isReplicaSet checks if the pod is controlled by a ReplicaSet making it safe to delete
func (manager TokenManager) isReplicaSet(ownerReferences []metav1.OwnerReference) bool {
	for _, ownerRef := range ownerReferences {
		if ownerRef.Kind == replicaSetKind {
			return true
		}
	}
	return false
}
