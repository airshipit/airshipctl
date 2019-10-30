package document_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/v3/pkg/gvk"
	"sigs.k8s.io/kustomize/v3/pkg/types"

	"opendev.org/airship/airshipctl/testutil"
)

func TestNewBundle(t *testing.T) {

	bundle := testutil.NewTestBundle(t, "testdata")
	require := require.New(t)
	require.NotNil(bundle)

}

func TestBundleDocumentFiltering(t *testing.T) {

	bundle := testutil.NewTestBundle(t, "testdata")
	assert := assert.New(t)

	t.Run("GetKustomizeResourceMap", func(t *testing.T) {

		r := bundle.GetKustomizeResourceMap()

		// ensure it is populated
		assert.True(r.Size() > 0)

	})

	t.Run("GetByGvk", func(t *testing.T) {

		docs, err := bundle.GetByGvk("apps", "v1", "Deployment")
		if err != nil {
			t.Fatalf("Error trying to GetGvk: %v", err)
		}

		assert.Equal(len(docs), 3, "GetGvk returned the wrong number of resources")

	})

	t.Run("GetByAnnotation", func(t *testing.T) {

		docs, err := bundle.GetByAnnotation("airshipit.org/clustertype=ephemeral")
		if err != nil {
			t.Fatalf("Error trying to GetByAnnotation: %v", err)
		}

		assert.Equal(len(docs), 4, "GetByAnnotation returned wrong number of resources")

	})

	t.Run("GetByLabel", func(t *testing.T) {

		docs, err := bundle.GetByLabel("app=workflow-controller")
		if err != nil {
			t.Fatalf("Error trying to GetByLabel: %v", err)
		}

		assert.Equal(len(docs), 1, "GetByLabel returned wrong number of resources")

	})

	t.Run("SelectGvk", func(t *testing.T) {

		// Select* tests test the Kustomize selector, which requires we build Kustomize
		// selector objects which is useful for advanced searches that
		// need to stack filters
		g := gvk.Gvk{Group: "apps", Version: "v1", Kind: "Deployment"}
		selector := types.Selector{Gvk: g}
		docs, err := bundle.Select(selector)
		if err != nil {
			t.Fatalf("Error trying to select resources: %v", err)
		}
		assert.Equal(len(docs), 3, "SelectGvk returned wrong number of resources")

	})

	t.Run("SelectAnnotations", func(t *testing.T) {

		// Find documents with a particular annotation, namely airshipit.org/clustertype=ephemeral
		selector := types.Selector{AnnotationSelector: "airshipit.org/clustertype=ephemeral"}
		docs, err := bundle.Select(selector)
		if err != nil {
			t.Fatalf("Error trying to select annotated resources: %v", err)
		}
		assert.Equal(len(docs), 4, "SelectAnnotations returned wrong number of resources")

	})

	t.Run("SelectLabels", func(t *testing.T) {

		// Find documents with a particular label, namely app=workflow-controller
		// note how this will only find resources labeled at the top most level (metadata.labels)
		// and not spec templates with this label (spec.template.metadata.labels)
		selector := types.Selector{LabelSelector: "app=workflow-controller"}
		docs, err := bundle.Select(selector)
		if err != nil {
			t.Fatalf("Error trying to select labeled resources: %v", err)
		}
		assert.Equal(len(docs), 1, "SelectLabels returned wrong number of resources")

	})

	t.Run("Write", func(t *testing.T) {

		// Ensure we can write out a bundle
		//
		// alanmeadows(TODO) improve validation what we write looks correct
		var b bytes.Buffer

		err := bundle.Write(&b)
		if err != nil {
			t.Fatalf("Failed to write bundle out: %v", err)
		}

		// b.String() will contain all kustomize built YAML
		// so we for now we just search for an expected string
		// obviously, this should be improved
		assert.Contains(b.String(), "workflow-controller")

	})

}
