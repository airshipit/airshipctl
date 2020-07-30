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

package cluster_test

import (
	"testing"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"

	"opendev.org/airship/airshipctl/cmd/cluster"
	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/environment"
	"opendev.org/airship/airshipctl/pkg/k8s/client"
	"opendev.org/airship/airshipctl/pkg/k8s/client/fake"
	"opendev.org/airship/airshipctl/testutil"
)

const (
	fixturesPath = "testdata/statusmap"
)

func TestNewClusterStatusCmd(t *testing.T) {
	tests := []struct {
		cmdTest   *testutil.CmdTest
		resources []runtime.Object
		CRDs      []runtime.Object
	}{
		{
			cmdTest: &testutil.CmdTest{
				Name:    "check-status-no-resources",
				CmdLine: "",
			},
		},
		{
			cmdTest: &testutil.CmdTest{
				Name:    "cluster-status-cmd-with-help",
				CmdLine: "--help",
			},
		},
		{
			cmdTest: &testutil.CmdTest{
				Name:    "check-status-with-resources",
				CmdLine: "",
			},
			resources: []runtime.Object{
				makeResource("Resource", "stable-resource", "stable"),
				makeResource("Resource", "pending-resource", "pending"),
			},
			CRDs: []runtime.Object{
				makeResourceCRD(annotationValidStatusCheck()),
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		testClientFactory := func(_ *environment.AirshipCTLSettings) (client.Interface, error) {
			return fake.NewClient(
				fake.WithDynamicObjects(tt.resources...),
				fake.WithCRDs(tt.CRDs...),
			), nil
		}
		tt.cmdTest.Cmd = cluster.NewStatusCommand(clusterStatusTestSettings(), testClientFactory)
		testutil.RunTest(t, tt.cmdTest)
	}
}

func clusterStatusTestSettings() *environment.AirshipCTLSettings {
	return &environment.AirshipCTLSettings{
		Config: &config.Config{
			Clusters:  map[string]*config.ClusterPurpose{"testCluster": nil},
			AuthInfos: map[string]*config.AuthInfo{"testAuthInfo": nil},
			Contexts: map[string]*config.Context{
				"testContext": {Manifest: "testManifest"},
			},
			Manifests: map[string]*config.Manifest{
				"testManifest": {TargetPath: fixturesPath},
			},
			CurrentContext: "testContext",
		},
	}
}

func makeResource(kind, name, state string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "example.com/v1",
			"kind":       kind,
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": "default",
			},
			"status": map[string]interface{}{
				"state": state,
			},
		},
	}
}

func annotationValidStatusCheck() map[string]string {
	return map[string]string{
		"airshipit.org/status-check": `
[
  {
    "status": "Stable",
    "condition": "@.status.state==\"stable\""
  },
  {
    "status": "Pending",
    "condition": "@.status.state==\"pending\""
  }
]`,
	}
}

func makeResourceCRD(annotations map[string]string) *apiextensionsv1.CustomResourceDefinition {
	return &apiextensionsv1.CustomResourceDefinition{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CustomResourceDefinition",
			APIVersion: "apiextensions.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        "resources.example.com",
			Annotations: annotations,
		},
		Spec: apiextensionsv1.CustomResourceDefinitionSpec{
			Group: "example.com",
			Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
				{
					Name:    "v1",
					Served:  true,
					Storage: true,
				},
			},
			// omitting the openAPIV3Schema for brevity
			Scope: "Namespaced",
			Names: apiextensionsv1.CustomResourceDefinitionNames{
				Kind:     "Resource",
				Plural:   "resources",
				Singular: "resource",
			},
		},
	}
}
