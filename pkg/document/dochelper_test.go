package document_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/testutil"
)

func TestDocHelpers(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	fSys := testutil.SetupTestFs(t, "testdata/dochelper")
	bundle, err := document.NewBundle(fSys, "/", "/")
	require.NoError(err, "Building Bundle Failed")
	require.NotNil(bundle)

	t.Run("GetBMHNetworkData", func(t *testing.T) {
		// retrieve our single bmh in the dataset
		selector := document.NewSelector().ByKind("BareMetalHost")
		doc, err := bundle.SelectOne(selector)
		require.NoError(err)

		networkData, err := document.GetBMHNetworkData(doc, bundle)
		require.NoError(err, "Unexpected error trying to GetBMHNetworkData")
		assert.Equal(networkData, "some network data")
	})

	t.Run("GetBMHBMCAddress", func(t *testing.T) {
		// retrieve our single bmh in the dataset
		selector := document.NewSelector().ByKind("BareMetalHost")
		doc, err := bundle.SelectOne(selector)
		require.NoError(err)

		bmcAddress, err := document.GetBMHBMCAddress(doc)
		require.NoError(err, "Unexpected error trying to GetBMHBMCAddress")
		assert.Equal(bmcAddress, "redfish+https://192.168.111.1/v1/Redfish/Foo/Bar")
	})

	t.Run("GetBMHBMCCredentials", func(t *testing.T) {
		// retrieve our single bmh in the dataset
		selector := document.NewSelector().ByKind("BareMetalHost")
		doc, err := bundle.SelectOne(selector)
		require.NoError(err)

		bmcUsername, bmcPassword, err := document.GetBMHBMCCredentials(doc, bundle)
		require.NoError(err, "Unexpected error trying to GetBMHBMCCredentials")
		assert.Equal(bmcUsername, "username")
		assert.Equal(bmcPassword, "password")
	})
}
