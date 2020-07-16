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

package document_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	airapiv1 "opendev.org/airship/airshipctl/pkg/api/v1alpha1"
	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/testutil"
)

func TestDocument(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	// the easiest way to construct a bunch of documents
	// is by manufacturing a bundle
	//
	// alanmeadows(TODO): at some point
	// refactoring this so there isn't a reliance
	// on a bundle might be useful
	fSys := testutil.SetupTestFs(t, "testdata/common")
	bundle, err := document.NewBundle(fSys, "/")
	require.NoError(err, "Building Bundle Failed")
	require.NotNil(bundle)

	t.Run("GetName", func(t *testing.T) {
		docs, err := bundle.GetAllDocuments()
		require.NoError(err, "Unexpected error trying to GetAllDocuments")

		nameList := make([]string, 0, len(docs))

		for _, doc := range docs {
			nameList = append(nameList, doc.GetName())
		}

		assert.Contains(nameList, "tiller-deploy", "Could not find expected name")
	})

	t.Run("AsYAML", func(t *testing.T) {
		doc, err := bundle.GetByName("some-random-deployment-we-will-filter")
		require.NoError(err, "Unexpected error trying to GetByName for AsYAML Test")

		// see if we can marshal it while we're here for coverage
		// as this is a dependency for AsYAML
		json, err := doc.MarshalJSON()
		require.NoError(err, "Unexpected error trying to MarshalJSON()")
		assert.NotNil(json)

		// get it as yaml
		yaml, err := doc.AsYAML()
		require.NoError(err, "Unexpected error trying to AsYAML()")

		// convert the bytes into a string for comparison
		//
		// alanmeadows(NOTE): marshal can reorder things
		// in the yaml, and does not return document beginning
		// or end markers that may of been in the source so
		// the FixtureInitiallyIgnored has been altered to
		// look more or less how unmarshalling it would look
		s := string(yaml)
		fileData, err := fSys.ReadFile("/initially_ignored.yaml")
		require.NoError(err, "Unexpected error reading initially_ignored.yaml file")

		// increase the chance of a match by removing any \n suffix on both actual
		// and expected
		assert.Equal(strings.TrimRight(s, "\n"), strings.TrimRight(string(fileData), "\n"))
	})

	t.Run("ToObject", func(t *testing.T) {
		expectedObj := map[string]interface{}{
			"apiVersion": "airshipit.org/v1alpha1",
			"kind":       "Phase",
			"metadata": map[string]interface{}{
				"name": "initinfra",
			},
			"config": map[string]interface{}{
				"documentEntryPoint": "manifests/site/test-site/initinfra",
			},
		}
		doc, err := bundle.GetByName("initinfra")
		require.NoError(err)
		actualObj := make(map[string]interface{})
		err = doc.ToObject(&actualObj)
		assert.NoError(err)
		assert.Equal(expectedObj, actualObj)
	})

	t.Run("ToAPIObject", func(t *testing.T) {
		expectedObj := &airapiv1.Clusterctl{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Clusterctl",
				APIVersion: "airshipit.org/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "clusterctl-v1",
			},
			Providers: []*airapiv1.Provider{
				{
					Name: "aws",
					Type: "InfrastructureProvider",
					URL:  "/manifests/capi/infra/aws/v0.3.0",
				},
			},
		}
		sel, err := document.NewSelector().ByObject(expectedObj, airapiv1.Scheme)
		require.NoError(err)
		doc, err := bundle.SelectOne(sel)
		require.NoError(err)
		actualObj := &airapiv1.Clusterctl{}
		err = doc.ToAPIObject(actualObj, airapiv1.Scheme)
		assert.NoError(err)
		assert.Equal(expectedObj, actualObj)
	})

	t.Run("GetString", func(t *testing.T) {
		doc, err := bundle.GetByName("some-random-deployment-we-will-filter")
		require.NoError(err, "Unexpected error trying to GetByName")

		appLabelMatch, err := doc.GetString("spec.selector.matchLabels.app")
		require.NoError(err, "Unexpected error trying to GetString from document")

		assert.Equal(appLabelMatch, "some-random-deployment-we-will-filter")
	})

	t.Run("GetNamespace", func(t *testing.T) {
		doc, err := bundle.GetByName("some-random-deployment-we-will-filter")
		require.NoError(err, "Unexpected error trying to GetByName")

		assert.Equal("foobar", doc.GetNamespace())
	})

	t.Run("GetStringSlice", func(t *testing.T) {
		doc, err := bundle.GetByName("some-random-deployment-we-will-filter")
		require.NoError(err, "Unexpected error trying to GetByName")
		s := []string{"foobar"}

		gotSlice, err := doc.GetStringSlice("spec.template.spec.containers[0].args")
		require.NoError(err, "Unexpected error trying to GetStringSlice")

		assert.Equal(s, gotSlice)
	})

	t.Run("Annotate", func(t *testing.T) {
		docs, err := bundle.GetAllDocuments()
		require.NoError(err, "Unexpected error trying to GetAllDocuments")
		annotationMap := map[string]string{
			"test-annotation": "test-annotaiton-value",
		}

		for _, doc := range docs {
			doc.Annotate(annotationMap)
			annotationList := doc.GetAnnotations()
			assert.Equal(annotationList["test-annotation"], "test-annotaiton-value")
		}
	})

	t.Run("Label", func(t *testing.T) {
		docs, err := bundle.GetAllDocuments()
		require.NoError(err, "Unexpected error trying to GetAllDocuments")
		labelMap := map[string]string{
			"test-label": "test-label-value",
		}

		for _, doc := range docs {
			doc.Label(labelMap)
			labelList := doc.GetLabels()
			assert.Equal(labelList["test-label"], "test-label-value")
		}
	})

	t.Run("GetGroup", func(t *testing.T) {
		doc, err := bundle.GetByName("some-random-deployment-we-will-filter")
		require.NoError(err, "Unexpected error trying to GetByName")
		group := doc.GetGroup()
		require.NotNil(group)
		assert.Equal(group, "apps")
	})

	t.Run("GetVersion", func(t *testing.T) {
		doc, err := bundle.GetByName("some-random-deployment-we-will-filter")
		require.NoError(err, "Unexpected error trying to GetByName")
		version := doc.GetVersion()
		require.NotNil(version)
		assert.Equal(version, "v1")
	})

	t.Run("GetBool", func(t *testing.T) {
		doc, err := bundle.GetByName("some-random-deployment-we-will-filter")
		require.NoError(err, "Unexpected error trying to GetByName")
		boolValue, err := doc.GetBool("spec.template.spec.containers[0].env[0].value")
		require.NoError(err, "Unexpected error trying to GetBool from document")
		assert.Equal(boolValue, false)
	})

	t.Run("GetFloat64", func(t *testing.T) {
		doc, err := bundle.GetByName("some-random-deployment-we-will-filter")
		require.NoError(err, "Unexpected error trying to GetByName")
		floatValue, err := doc.GetFloat64("spec.template.spec.containers[0].env[1].value")
		require.NoError(err, "Unexpected error trying to GetFloat from document")
		assert.Equal(floatValue, float64(1.012))
	})

	t.Run("GetInt64", func(t *testing.T) {
		doc, err := bundle.GetByName("some-random-deployment-we-will-filter")
		require.NoError(err, "Unexpected error trying to GetByName")
		intValue, err := doc.GetInt64("spec.template.spec.containers[0].env[2].value")
		require.NoError(err, "Unexpected error trying to GetInt from document")
		assert.Equal(intValue, int64(1000))
	})

	t.Run("GetSlice", func(t *testing.T) {
		doc, err := bundle.GetByName("some-random-deployment-we-will-filter")
		require.NoError(err, "Unexpected error trying to GetByName")
		s := []interface{}{"foobar"}
		gotSlice, err := doc.GetSlice("spec.template.spec.containers[0].args")
		require.NoError(err, "Unexpected error trying to GetSlice")
		assert.Equal(s, gotSlice)
	})

	t.Run("GetMap", func(t *testing.T) {
		doc, err := bundle.GetByName("some-random-deployment-we-will-filter")
		require.NoError(err, "Unexpected error trying to GetByName")
		gotMap, err := doc.GetMap("spec.template.spec.containers[0].env[0]")
		require.NoError(err, "Unexpected error trying to GetMap")
		assert.NotNil(gotMap)
	})
}

func TestNewDocumentFromBytes(t *testing.T) {
	tests := []struct {
		name            string
		stringData      string
		expectErr       bool
		expectedDocName string
	}{
		{
			name: "ConfigMap",
			stringData: `apiVersion: v1
kind: Secret
metadata:
  name: control-0-bmc
  namespace: metal3
type: Opaque
stringData:
  username: username
  password: password`,
			expectErr:       false,
			expectedDocName: "control-0-bmc",
		},
		{
			name: "Manformed Bytes",
			stringData: `
			broken:fas -<
				fasd`,
			expectErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			doc, err := document.NewDocumentFromBytes([]byte(tt.stringData))
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, doc.GetName(), tt.expectedDocName)
			}
		})
	}
}
