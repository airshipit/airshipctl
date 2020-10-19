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

package resetsatoken_test

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
	kfake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	ktesting "k8s.io/client-go/testing"

	"opendev.org/airship/airshipctl/pkg/cluster/resetsatoken"
	"opendev.org/airship/airshipctl/pkg/k8s/client/fake"
	"opendev.org/airship/airshipctl/testutil"
)

type testCase struct {
	name             string
	existingSecrets  []*v1.Secret
	existingPods     []*v1.Pod
	secretName       string
	secretNamespace  string
	numPodDeletes    int
	numSecretDeletes int
	expectErr        bool
}

var testCases = []testCase{
	{
		name:      "no-pods-secrets",
		expectErr: true,
	},
	{
		name:             "valid-secret-no-pod",
		secretName:       "valid-secret",
		secretNamespace:  "valid-namespace",
		existingSecrets:  []*v1.Secret{getSecret()},
		numSecretDeletes: 1,
	},
	{
		name:             "valid-secret-no-pod-empty-filter",
		secretNamespace:  "valid-namespace",
		existingSecrets:  []*v1.Secret{getSecret()},
		numSecretDeletes: 1,
	},
	{
		name:            "invalid-secret-no-pod",
		secretName:      "invalid-secret",
		existingSecrets: []*v1.Secret{getSecret()},
		secretNamespace: "valid-namespace",
	},
	{
		name:            "unmatched-secret-pod",
		secretName:      "invalid-secret",
		secretNamespace: "valid-namespace",
		existingPods:    []*v1.Pod{getPod()},
		existingSecrets: []*v1.Secret{getSecret()},
	},
	{
		name:             "matched-secret-pod",
		secretName:       "valid-secret",
		secretNamespace:  "valid-namespace",
		existingPods:     []*v1.Pod{getPod()},
		existingSecrets:  []*v1.Secret{getSecret()},
		numPodDeletes:    1,
		numSecretDeletes: 1,
	},
}

func TestResetSaToken(t *testing.T) {
	for _, testCase := range testCases {
		cfg, _ := testutil.InitConfig(t)

		var objects []runtime.Object
		for _, pod := range testCase.existingPods {
			objects = append(objects, pod)
		}
		for _, secret := range testCase.existingSecrets {
			objects = append(objects, secret)
		}
		ra := fake.WithTypedObjects(objects...)
		kclient := fake.NewClient(ra)

		assert.NotEmpty(t, kclient)
		assert.NotEmpty(t, cfg)

		clientset := kclient.ClientSet()
		manager, err := resetsatoken.NewTokenManager(clientset)
		assert.NoError(t, err)

		err = manager.RotateToken(testCase.secretNamespace, testCase.secretName)
		if testCase.expectErr {
			assert.Error(t, err)
			continue
		}

		actions := clientset.(*kfake.Clientset).Actions()

		podDeleteActions := filterActions(actions, "pods", "delete")
		assert.Len(t, podDeleteActions, testCase.numPodDeletes)

		secretDeleteActions := filterActions(actions, "secrets", "delete")
		assert.Len(t, secretDeleteActions, testCase.numSecretDeletes)
	}
}

func getSecret() *v1.Secret {
	object := readObjectFromFile("testdata/secret.yaml")
	if secret, ok := object.(*v1.Secret); ok {
		return secret
	}
	return nil
}

func getPod() *v1.Pod {
	object := readObjectFromFile("testdata/pod.yaml")
	if pod, ok := object.(*v1.Pod); ok {
		return pod
	}
	return nil
}

func readObjectFromFile(fileName string) runtime.Object {
	contents, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil
	}
	jsonContents, err := yaml.ToJSON(contents)
	if err != nil {
		return nil
	}

	object, err := runtime.Decode(scheme.Codecs.UniversalDeserializer(), jsonContents)
	if err != nil {
		return nil
	}
	return object
}

func filterActions(actions []ktesting.Action, resource string, verb string) []ktesting.Action {
	var result []ktesting.Action
	for _, action := range actions {
		if action.GetVerb() == verb && action.GetResource().Resource == resource {
			result = append(result, action)
		}
	}
	return result
}
