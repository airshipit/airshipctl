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
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/testutil"
)

func TestNewBundle(t *testing.T) {
	require := require.New(t)

	bundle := testutil.NewTestBundle(t, "testdata/common")
	require.NotNil(bundle)
}

func TestBundleDocumentFiltering(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	bundle := testutil.NewTestBundle(t, "testdata/common")

	t.Run("GetByGvk", func(t *testing.T) {
		docs, err := bundle.GetByGvk("apps", "v1", "Deployment")
		require.NoError(err)

		assert.Len(docs, 3)
	})

	t.Run("GetByAnnotation", func(t *testing.T) {
		docs, err := bundle.GetByAnnotation("airshipit.org/clustertype=ephemeral")
		require.NoError(err, "Error trying to GetByAnnotation")

		assert.Len(docs, 4)
	})

	t.Run("GetByLabel", func(t *testing.T) {
		docs, err := bundle.GetByLabel("app=workflow-controller")
		require.NoError(err, "Error trying to GetByLabel")

		assert.Len(docs, 1)
	})

	t.Run("SelectGvk", func(t *testing.T) {
		// Select* tests test the Kustomize selector, which requires we build Kustomize
		// selector objects which is useful for advanced searches that
		// need to stack filters
		selector := document.NewSelector().ByGvk("apps", "v1", "Deployment")
		docs, err := bundle.Select(selector)
		require.NoError(err, "Error trying to select resources")

		assert.Len(docs, 3)
	})

	t.Run("SelectAnnotations", func(t *testing.T) {
		// Find documents with a particular annotation, namely airshipit.org/clustertype=ephemeral
		selector := document.NewSelector().ByAnnotation("airshipit.org/clustertype=ephemeral")
		docs, err := bundle.Select(selector)
		require.NoError(err, "Error trying to select annotated resources")

		assert.Len(docs, 4)
	})

	t.Run("SelectLabels", func(t *testing.T) {
		// Find documents with a particular label, namely app=workflow-controller
		// note how this will only find resources labeled at the top most level (metadata.labels)
		// and not spec templates with this label (spec.template.metadata.labels)
		selector := document.NewSelector().ByLabel("app=workflow-controller")
		docs, err := bundle.Select(selector)
		require.NoError(err, "Error trying to select labeled resources")

		assert.Len(docs, 1)
	})

	t.Run("SelectByLabelAndName", func(t *testing.T) {
		// Find documents with a particular label and name,
		// namely app=workflow-controller and name workflow-controller
		selector := document.NewSelector().ByName("workflow-controller").ByLabel("app=workflow-controller")
		docs, err := bundle.Select(selector)
		require.NoError(err, "Error trying to select labeled with name resources")

		assert.Len(docs, 1)
	})

	t.Run("SelectByTwoLabels", func(t *testing.T) {
		// Find documents with two particular label, namely app=workflow-controller arbitrary-label=some-label
		selector := document.NewSelector().
			ByLabel("app=workflow-controller").
			ByLabel("arbitrary-label=some-label")
		docs, err := bundle.Select(selector)
		require.NoError(err, "Error trying to select by two labels")

		assert.Len(docs, 1)
		assert.Equal("workflow-controller", docs[0].GetName())
	})

	t.Run("SelectByTwoAnnotations", func(t *testing.T) {
		// Find documents with two particular annotations,
		// namely app=workflow-controller and name workflow-controller
		selector := document.NewSelector().
			ByAnnotation("airshipit.org/clustertype=target").
			ByAnnotation("airshipit.org/random-payload=random")
		docs, err := bundle.Select(selector)
		require.NoError(err, "Error trying to select by two annotations")

		assert.Len(docs, 1)
		assert.Equal("argo-ui", docs[0].GetName())
	})

	t.Run("SelectByFieldValue", func(t *testing.T) {
		// Find documents with a particular value referenced by JSON path
		filter := func(val interface{}) bool { return val == "01:3b:8b:0c:ec:8b" }
		filteredBundle, err := bundle.SelectByFieldValue("spec.bootMACAddress", filter)
		require.NoError(err, "Error trying to filter resources")
		docs, err := filteredBundle.GetAllDocuments()
		require.NoError(err)
		assert.Len(docs, 1)
		assert.Equal(docs[0].GetKind(), "BareMetalHost")
		assert.Equal(docs[0].GetName(), "master-1")
	})

	t.Run("Write", func(t *testing.T) {
		// Ensure we can write out a bundle
		//
		// alanmeadows(TODO) improve validation what we write looks correct
		var b bytes.Buffer

		err := bundle.Write(&b)
		require.NoError(err, "Failed to write bundle out")

		// b.String() will contain all kustomize built YAML
		// so we for now we just search for an expected string
		// obviously, this should be improved
		assert.Contains(b.String(), "workflow-controller")
	})

	t.Run("BundleAppend", func(t *testing.T) {
		doc, err := document.NewDocumentFromBytes([]byte(`apiVersion: v1
kind: Secret
metadata:
  name: append-secret
  namespace: metal3
type: Opaque
stringData:
  username: append-username
  password: append-password`))
		require.NoError(err)
		err = bundle.Append(doc)
		require.NoError(err)
		secretDoc, err := bundle.SelectOne(document.NewSelector().ByKind("Secret").ByName("append-secret"))
		require.NoError(err)
		require.NotNil(t, secretDoc)
		password, err := secretDoc.GetString("stringData.password")
		require.NoError(err)
		username, err := secretDoc.GetString("stringData.username")
		require.NoError(err)
		assert.Equal(password, "append-password")
		assert.Equal(username, "append-username")
	})
}

func TestBundleOrder(t *testing.T) {
	bundle := testutil.NewTestBundle(t, "testdata/order")

	docs, err := bundle.GetAllDocuments()
	require.NoError(t, err)
	require.Len(t, docs, 3)

	// first must be namespace argo-namespace
	doc := docs[0]
	require.NotNil(t, doc)
	assert.Equal(t, "Namespace", doc.GetKind())
	assert.Equal(t, "argo-namespace", doc.GetName())

	// second must be CRD named workflows.argoproj.io
	doc = docs[1]
	require.NotNil(t, doc)
	assert.Equal(t, "CustomResourceDefinition", doc.GetKind())
	assert.Equal(t, "workflows.argoproj.io", doc.GetName())

	// second must be CR workflow-controller
	doc = docs[2]
	require.NotNil(t, doc)
	assert.Equal(t, "Deployment", doc.GetKind())
	assert.Equal(t, "workflow-controller", doc.GetName())
}
