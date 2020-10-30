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

package checkexpiration_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"k8s.io/apimachinery/pkg/runtime"

	"opendev.org/airship/airshipctl/pkg/cluster/checkexpiration"
	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/k8s/client"
	"opendev.org/airship/airshipctl/pkg/k8s/client/fake"
	"opendev.org/airship/airshipctl/testutil"
)

type testCase struct {
	name                            string
	expiryThreshold                 int
	nodeTestFile                    string
	kubeconfTestFile                string
	tlsSecretTestFile               string
	nodeExpirationYear              string
	expectedExpiringNodeCount       int
	expectedExpiringKubeConfigCount int
	expectedExpiringTLSSecretCount  int
}

var (
	testCases = []*testCase{
		{
			name:                            "empty-expect-error",
			expectedExpiringNodeCount:       0,
			expectedExpiringKubeConfigCount: 0,
			expectedExpiringTLSSecretCount:  0,
		},
		{
			name:                      "node-cert-expiring",
			nodeTestFile:              nodeFile,
			nodeExpirationYear:        "2021",
			expiryThreshold:           testThreshold, // 20 years
			expectedExpiringNodeCount: 1,
		},
		{
			name:                      "node-cert-not-expiring",
			nodeExpirationYear:        "2025",
			nodeTestFile:              nodeFile,
			expiryThreshold:           10,
			expectedExpiringNodeCount: 0,
		},
		{
			name:                            "all-certs-not-expiring",
			nodeExpirationYear:              "2025",
			nodeTestFile:                    nodeFile,
			tlsSecretTestFile:               tlsSecretFile,
			kubeconfTestFile:                kubeconfFile,
			expiryThreshold:                 1,
			expectedExpiringNodeCount:       0,
			expectedExpiringKubeConfigCount: 0,
			expectedExpiringTLSSecretCount:  0,
		},
		{
			name:                            "all-certs-expiring",
			nodeExpirationYear:              "2021",
			nodeTestFile:                    nodeFile,
			tlsSecretTestFile:               tlsSecretFile,
			kubeconfTestFile:                kubeconfFile,
			expiryThreshold:                 testThreshold,
			expectedExpiringNodeCount:       1,
			expectedExpiringKubeConfigCount: 1,
			expectedExpiringTLSSecretCount:  1,
		},
	}
)

func TestCheckExpiration(t *testing.T) {
	for _, testCase := range testCases {
		cfg, _ := testutil.InitConfig(t)
		settings := func() (*config.Config, error) {
			return cfg, nil
		}

		var objects []runtime.Object

		if testCase.nodeExpirationYear != "" && testCase.nodeTestFile != "" {
			objects = append(objects, getNodeObject(t, testCase.nodeTestFile, testCase.nodeExpirationYear))
		}

		if testCase.tlsSecretTestFile != "" {
			objects = append(objects, getSecretObject(t, testCase.tlsSecretTestFile))
		}

		if testCase.kubeconfTestFile != "" {
			objects = append(objects, getSecretObject(t, testCase.kubeconfTestFile))
		}

		ra := fake.WithTypedObjects(objects...)

		clientFactory := func(_ string, _ string) (client.Interface, error) {
			return fake.NewClient(ra), nil
		}

		store, err := checkexpiration.NewStore(settings, clientFactory, "", "", testCase.expiryThreshold)
		assert.NoError(t, err)

		expirationInfo := store.GetExpiringCertificates()

		assert.Len(t, expirationInfo.Kubeconfs, testCase.expectedExpiringKubeConfigCount)

		assert.Len(t, expirationInfo.TLSSecrets, testCase.expectedExpiringTLSSecretCount)

		assert.Len(t, expirationInfo.NodeCerts, testCase.expectedExpiringNodeCount)
	}
}
