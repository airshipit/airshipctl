package document_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/testutil"
)

func TestSelectorsPositive(t *testing.T) {
	bundle := testutil.NewTestBundle(t, "testdata/selectors/valid")

	t.Run("TestEphemeralCloudDataSelector", func(t *testing.T) {
		doc, err := bundle.Select(document.NewEphemeralCloudDataSelector())
		require.NoError(t, err)
		assert.Len(t, doc, 1)
	})

	t.Run("TestEphemeralNetworkDataSelector", func(t *testing.T) {
		docs, err := bundle.Select(document.NewEphemeralBMHSelector())
		require.NoError(t, err)
		assert.Len(t, docs, 1)
		bmhDoc := docs[0]
		selector, err := document.NewNetworkDataSelector(bmhDoc)
		require.NoError(t, err)
		assert.Equal(t, "validName", selector.Name)
	})

	t.Run("TestEphemeralCloudDataSelector", func(t *testing.T) {
		doc, err := bundle.Select(document.NewEphemeralCloudDataSelector())
		require.NoError(t, err)
		assert.Len(t, doc, 1)
	})
}

func TestSelectorsNegative(t *testing.T) {
	// These two tests take bundle with two malformed documents
	// each of the documents will fail at different locations providing higher
	// test coverage
	bundle := testutil.NewTestBundle(t, "testdata/selectors/invalid")

	t.Run("TestNewNetworkDataSelectorErr", func(t *testing.T) {
		docs, err := bundle.Select(document.NewEphemeralBMHSelector())
		require.NoError(t, err)
		assert.Len(t, docs, 2)
		bmhDoc := docs[0]
		_, err = document.NewNetworkDataSelector(bmhDoc)
		assert.Error(t, err)
	})

	t.Run("TestEphemeralNetworkDataSelectorErr", func(t *testing.T) {
		docs, err := bundle.Select(document.NewEphemeralBMHSelector())
		require.NoError(t, err)
		assert.Len(t, docs, 2)
		bmhDoc := docs[1]
		_, err = document.NewNetworkDataSelector(bmhDoc)
		assert.Error(t, err)
	})
}

func TestSelectorsSkip(t *testing.T) {
	// These two tests take bundle with two malformed documents
	// each of the documents will fail at different locations providing higher
	// test coverage
	bundle := testutil.NewTestBundle(t, "testdata/selectors/exclude-from-k8s")

	t.Run("TestNewNetworkDataSelectorErr", func(t *testing.T) {
		selector := document.NewDeployToK8sSelector()
		docs, err := bundle.Select(selector)
		require.NoError(t, err)
		assert.Len(t, docs, 5)
		for _, doc := range docs {
			assert.NotEqual(t, "ignore-namespace", doc.GetName())
			assert.NotEqual(t, "ignore-bmh", doc.GetName())
		}
	})
}

func TestSelectorString(t *testing.T) {
	tests := []struct {
		name     string
		selector document.Selector
		expected string
	}{
		{
			name:     "unconditional",
			selector: document.Selector{},
			expected: "No selection conditions specified",
		},
		{
			name:     "by-name",
			selector: document.NewSelector().ByName("foo"),
			expected: `[Name="foo"]`,
		},
		{
			name: "by-all",
			selector: document.NewSelector().
				ByGvk("testGroup", "testVersion", "testKind").
				ByNamespace("testNamespace").
				ByName("testName").
				ByAnnotation("testAnnotation=true").
				ByLabel("testLabel=true"),
			expected: `[Group="testGroup", Version="testVersion", Kind="testKind", ` +
				`Namespace="testNamespace", Name="testName", ` +
				`Annotations="testAnnotation=true", Labels="testLabel=true"]`,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.selector.String())
		})
	}
}
