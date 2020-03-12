package cloudinit

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/v3/pkg/types"

	"opendev.org/airship/airshipctl/pkg/document"
)

func TestGetCloudData(t *testing.T) {
	bundle, err := document.NewBundleByPath("testdata")
	require.NoError(t, err, "Building Bundle Failed")

	tests := []struct {
		labelFilter      string
		expectedUserData []byte
		expectedNetData  []byte
		expectedErr      error
	}{
		{
			labelFilter:      "test=validdocset",
			expectedUserData: []byte("cloud-init"),
			expectedNetData:  []byte("net-config"),
			expectedErr:      nil,
		},
		{
			labelFilter:      "test=ephemeralmissing",
			expectedUserData: nil,
			expectedNetData:  nil,
			expectedErr: document.ErrDocNotFound{
				Selector: document.NewSelector().
					ByLabel(document.EphemeralHostSelector).
					ByKind("BareMetalHost"),
			},
		},
		{
			labelFilter:      "test=ephemeralduplicate",
			expectedUserData: nil,
			expectedNetData:  nil,
			expectedErr: document.ErrMultipleDocsFound{
				Selector: document.NewSelector().
					ByLabel(document.EphemeralHostSelector).
					ByKind("BareMetalHost"),
			},
		},
		{
			labelFilter:      "test=networkdatabadpointer",
			expectedUserData: nil,
			expectedNetData:  nil,
			expectedErr: document.ErrDocNotFound{
				Selector: document.NewSelector().
					ByKind("Secret").
					ByNamespace("networkdatabadpointer-missing").
					ByName("networkdatabadpointer-missing"),
			},
		},
		{
			labelFilter:      "test=networkdatamalformed",
			expectedUserData: nil,
			expectedNetData:  nil,
			expectedErr:      ErrDataNotSupplied{DocName: "networkdatamalformed-malformed", Key: networkDataKey},
		},
		{
			labelFilter:      "test=networkdatamissing",
			expectedUserData: nil,
			expectedNetData:  nil,
			expectedErr:      types.NoFieldError{Field: "spec.networkData.name"},
		},
		{
			labelFilter:      "test=userdatamalformed",
			expectedUserData: nil,
			expectedNetData:  nil,
			expectedErr:      ErrDataNotSupplied{DocName: "userdatamalformed-somesecret", Key: userDataKey},
		},
		{
			labelFilter:      "test=userdatamissing",
			expectedUserData: nil,
			expectedNetData:  nil,
			expectedErr: document.ErrDocNotFound{
				Selector: document.NewSelector().
					ByKind("Secret").
					ByLabel(document.EphemeralUserDataSelector),
			},
		},
	}

	for _, tt := range tests {
		// prune the bundle down using the label filter for the specific test
		selector := document.NewSelector().ByLabel(tt.labelFilter)
		filteredBundle, err := bundle.SelectBundle(selector)
		require.NoError(t, err, "Building filtered bundle for %s failed", tt.labelFilter)

		// ensure each test case filter has at least one document
		docs, err := filteredBundle.GetAllDocuments()
		require.NoError(t, err, "GetAllDocuments failed")
		require.NotZero(t, docs)

		actualUserData, actualNetData, actualErr := GetCloudData(filteredBundle)

		assert.Equal(t, tt.expectedUserData, actualUserData)
		assert.Equal(t, tt.expectedNetData, actualNetData)
		assert.Equal(t, tt.expectedErr, actualErr)
	}
}
