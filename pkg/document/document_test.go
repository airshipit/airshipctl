package document_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/testutil"
)

func TestDocument(t *testing.T) {

	// the easiest way to construct a bunch of documents
	// is by manufacturing a bundle
	//
	// alanmeadows(TODO): at some point
	// refactoring this so there isn't a reliance
	// on a bundle might be useful
	fSys := testutil.SetupTestFs(t, "testdata")
	bundle, err := document.NewBundle(fSys, "/", "/")
	if err != nil {
		t.Fatalf("Unexpected error building bundle: %v", err)
	}

	require := require.New(t)
	assert := assert.New(t)
	require.NotNil(bundle)

	t.Run("GetName", func(t *testing.T) {

		docs, err := bundle.GetAllDocuments()
		if err != nil {
			t.Fatalf("Unexpected error trying to GetAllDocuments: %v", err)
		}

		nameList := []string{}

		for _, doc := range docs {
			nameList = append(nameList, doc.GetName())
		}

		assert.Contains(nameList, "tiller-deploy", "Could not find expected name")

	})

	t.Run("AsYAML", func(t *testing.T) {

		doc, err := bundle.GetByName("some-random-deployment-we-will-filter")
		if err != nil {
			t.Fatalf("Unexpected error trying to GetByName for AsYAML Test: %v", err)
		}

		// see if we can marshal it while we're here for coverage
		// as this is a dependency for AsYAML
		json, err := doc.MarshalJSON()
		assert.NotNil(json)
		if err != nil {
			t.Fatalf("Unexpected error trying to MarshalJSON(): %v", err)
		}

		// get it as yaml
		yaml, err := doc.AsYAML()
		if err != nil {
			t.Fatalf("Unexpected error trying to AsYAML(): %v", err)
		}

		// convert the bytes into a string for comparison
		//
		// alanmeadows(NOTE): marshal can reorder things
		// in the yaml, and does not return document beginning
		// or end markers that may of been in the source so
		// the FixtureInitiallyIgnored has been altered to
		// look more or less how unmarshalling it would look
		s := string(yaml)
		fileData, err := fSys.ReadFile("/initially_ignored.yaml")
		if err != nil {
			t.Fatalf("Unexpected error reading initially_ignored.yaml file")
		}

		// increase the chance of a match by removing any \n suffix on both actual
		// and expected
		assert.Equal(strings.TrimSuffix(s, "\n"), strings.TrimSuffix(string(fileData), "\n"))

	})

	t.Run("GetString", func(t *testing.T) {

		doc, err := bundle.GetByName("some-random-deployment-we-will-filter")
		if err != nil {
			t.Fatalf("Unexpected error trying to GetByName for test: %v", err)
		}

		appLabelMatch, err := doc.GetString("spec.selector.matchLabels.app")
		if err != nil {
			t.Fatalf("Unexpected error trying to GetString from document")
		}
		assert.Equal(appLabelMatch, "some-random-deployment-we-will-filter")

	})

	t.Run("GetNamespace", func(t *testing.T) {

		doc, err := bundle.GetByName("some-random-deployment-we-will-filter")
		if err != nil {
			t.Fatalf("Unexpected error trying to GetByName for test: %v", err)
		}

		assert.Equal("foobar", doc.GetNamespace())

	})

	t.Run("GetStringSlice", func(t *testing.T) {

		doc, err := bundle.GetByName("some-random-deployment-we-will-filter")
		if err != nil {
			t.Fatalf("Unexpected error trying to GetByName for test: %v", err)
		}
		s := make([]string, 1)
		s[0] = "foobar"

		gotSlice, err := doc.GetStringSlice("spec.template.spec.containers[0].args")
		if err != nil {
			t.Fatalf("Unexpected error trying to GetStringSlice: %v", err)
		}

		assert.Equal(s, gotSlice)

	})

}
