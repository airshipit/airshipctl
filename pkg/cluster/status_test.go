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
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/cli-utils/pkg/kstatus/status"
	"sigs.k8s.io/cli-utils/pkg/object"

	"opendev.org/airship/airshipctl/pkg/cluster"
	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/k8s/client"
	"opendev.org/airship/airshipctl/pkg/k8s/client/fake"
	"opendev.org/airship/airshipctl/testutil"
)

func TestGetStatusMapDocs(t *testing.T) {
	tests := []struct {
		name      string
		resources []runtime.Object
		CRDs      []runtime.Object
	}{
		{
			name: "get-status-map-docs-no-resources",
		},
		{
			name: "get-status-map-docs-with-resources",
			resources: []runtime.Object{
				makeResource("stable-resource", "stable"),
				makeResource("pending-resource", "pending"),
			},
			CRDs: []runtime.Object{
				makeResourceCRD(annotationValidStatusCheck()),
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		settings := clusterStatusTestSettings()
		fakeClient := fake.NewClient(
			fake.WithDynamicObjects(tt.resources...),
			fake.WithCRDs(tt.CRDs...))
		clientFactory := func(_ string, _ string) (client.Interface, error) {
			return fakeClient, nil
		}
		statusOptions := cluster.NewStatusOptions(func() (*config.Config, error) {
			return settings, nil
		}, clientFactory, "")

		expectedSM, err := cluster.NewStatusMap(fakeClient)
		require.NoError(t, err)
		docBundle, err := document.NewBundleByPath(settings.Manifests["testManifest"].TargetPath)
		require.NoError(t, err)
		expectedDocs, err := docBundle.GetAllDocuments()
		require.NoError(t, err)

		sm, docs, err := statusOptions.GetStatusMapDocs()
		require.NoError(t, err)
		assert.Equal(t, expectedSM, sm)
		assert.Equal(t, expectedDocs, docs)
	}
}

func clusterStatusTestSettings() *config.Config {
	return &config.Config{
		Contexts: map[string]*config.Context{
			"testContext": {Manifest: "testManifest"},
		},
		Manifests: map[string]*config.Manifest{
			"testManifest": {TargetPath: "testdata/statusmap"},
		},
		CurrentContext: "testContext",
	}
}

func TestNewStatusMap(t *testing.T) {
	tests := []struct {
		name   string
		client *fake.Client
		err    error
	}{
		{
			name:   "no-failure-on-valid-status-check-annotation",
			client: fake.NewClient(fake.WithCRDs(makeResourceCRD(annotationValidStatusCheck()))),
			err:    nil,
		},
		{
			name:   "no-failure-when-missing-status-check-annotation",
			client: fake.NewClient(fake.WithCRDs(makeResourceCRD(nil))),
			err:    nil,
		},
		{
			name:   "missing-status",
			client: fake.NewClient(fake.WithCRDs(makeResourceCRD(annotationMissingStatus()))),
			err:    cluster.ErrInvalidStatusCheck{What: "missing status field"},
		},
		{
			name:   "missing-condition",
			client: fake.NewClient(fake.WithCRDs(makeResourceCRD(annotationMissingCondition()))),
			err:    cluster.ErrInvalidStatusCheck{What: "missing condition field"},
		},
		{
			name:   "malformed-status-check",
			client: fake.NewClient(fake.WithCRDs(makeResourceCRD(annotationMalformedStatusCheck()))),
			err: cluster.ErrInvalidStatusCheck{What: `unable to parse jsonpath: ` +
				`"{invalid json": invalid character 'i' looking for beginning of object key string`},
		},
	}

	for _, tt := range tests {
		tt := tt
		_, err := cluster.NewStatusMap(tt.client)
		assert.Equal(t, tt.err, err)
	}
}

func TestGetStatusForResource(t *testing.T) {
	tests := []struct {
		name           string
		selector       document.Selector
		client         *fake.Client
		expectedStatus status.Status
		err            error
	}{
		{
			name: "stable-resource-is-stable",
			selector: document.NewSelector().
				ByGvk("example.com", "v1", "Resource").
				ByName("stable-resource"),
			client: fake.NewClient(
				fake.WithCRDs(makeResourceCRD(annotationValidStatusCheck())),
				fake.WithDynamicObjects(makeResource("stable-resource", "stable")),
			),
			expectedStatus: status.Status("Stable"),
		},
		{
			name: "pending-resource-is-pending",
			selector: document.NewSelector().
				ByGvk("example.com", "v1", "Resource").
				ByName("pending-resource"),
			client: fake.NewClient(
				fake.WithCRDs(makeResourceCRD(annotationValidStatusCheck())),
				fake.WithDynamicObjects(makeResource("pending-resource", "pending")),
			),
			expectedStatus: status.Status("Pending"),
		},
		{
			name: "unknown-resource-is-unknown",
			selector: document.NewSelector().
				ByGvk("example.com", "v1", "Resource").
				ByName("unknown"),
			client: fake.NewClient(
				fake.WithCRDs(makeResourceCRD(annotationValidStatusCheck())),
				fake.WithDynamicObjects(makeResource("unknown", "unknown")),
			),
			expectedStatus: status.UnknownStatus,
		},
		{
			name: "missing-resource-returns-error",
			selector: document.NewSelector().
				ByGvk("example.com", "v1", "Missing").
				ByName("missing-resource"),
			client: fake.NewClient(),
			err:    cluster.ErrResourceNotFound{Resource: "missing-resource"},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			bundle := testutil.NewTestBundle(t, "testdata/statusmap")
			testStatusMap, err := cluster.NewStatusMap(tt.client)
			require.NoError(t, err)

			doc, err := bundle.SelectOne(tt.selector)
			require.NoError(t, err)

			actualStatus, err := testStatusMap.GetStatusForResource(doc)
			if tt.err != nil {
				assert.EqualError(t, err, tt.err.Error())
				// We expected an error - no need to check anything else
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, actualStatus)
		})
	}
}

func TestReadStatus(t *testing.T) {
	c := fake.NewClient(fake.WithCRDs(makeResourceCRD(annotationValidStatusCheck())),
		fake.WithDynamicObjects(makeResource("pending-resource", "pending")))
	statusMap, err := cluster.NewStatusMap(c)
	require.NoError(t, err)
	ctx := context.Background()
	resource := object.ObjMetadata{Namespace: "target-infra",
		Name: "pending-resource", GroupKind: schema.GroupKind{Group: "example.com", Kind: "Resource"}}
	result := statusMap.ReadStatus(ctx, resource)
	assert.Equal(t, "Pending", result.Status.String())
}

func makeResource(name, state string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "example.com/v1",
			"kind":       "Resource",
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": "target-infra",
			},
			"status": map[string]interface{}{
				"state": state,
			},
		},
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

func annotationMissingStatus() map[string]string {
	return map[string]string{
		"airshipit.org/status-check": `
[
  {
    "condition": "@.status.state==\"stable\""
  },
  {
    "condition": "@.status.state==\"pending\""
  }
]`,
	}
}

func annotationMissingCondition() map[string]string {
	return map[string]string{
		"airshipit.org/status-check": `
[
  {
    "status": "Stable"
  },
  {
    "status": "Pending"
  }
]`,
	}
}

func annotationMalformedStatusCheck() map[string]string {
	return map[string]string{"airshipit.org/status-check": "{invalid json"}
}
