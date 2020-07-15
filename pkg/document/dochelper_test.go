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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/testutil"
)

func TestDocHelpers(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	fSys := testutil.SetupTestFs(t, "testdata/dochelper/valid/")
	bundle, err := document.NewBundle(fSys, "/")
	require.NoError(err, "Building Bundle Failed")
	require.NotNil(bundle)

	t.Run("GetBMHNetworkData", func(t *testing.T) {
		// retrieve our single bmh in the dataset
		selector := document.NewSelector().ByKind("BareMetalHost")
		doc, err := bundle.SelectOne(selector)
		require.NoError(err)

		networkData, err := document.GetBMHNetworkData(doc, bundle)
		require.NoError(err, "Unexpected error trying to GetBMHNetworkData")
		assert.Equal(networkData, "some network data\n")
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

func TestDocHelpersNegativeCases(t *testing.T) {
	require := require.New(t)

	fSys := testutil.SetupTestFs(t, "testdata/dochelper/invalid/")
	bundle, err := document.NewBundle(fSys, "/")
	require.NoError(err, "Building Bundle Failed")
	require.NotNil(bundle)

	t.Run("GetBMHNetworkData", func(t *testing.T) {
		selector := document.NewSelector().ByKind("BareMetalHost")
		doc, err := bundle.SelectOne(selector)
		require.NoError(err)

		_, err = document.GetBMHNetworkData(doc, bundle)
		require.Error(err)
	})

	t.Run("GetBMHBMCAddress", func(t *testing.T) {
		// retrieve our single bmh in the dataset
		selector := document.NewSelector().ByKind("BareMetalHost")
		doc, err := bundle.SelectOne(selector)
		require.NoError(err)

		_, err = document.GetBMHBMCAddress(doc)
		require.Error(err)
	})

	t.Run("GetBMHBMCCredentials", func(t *testing.T) {
		// retrieve our single bmh in the dataset
		selector := document.NewSelector().ByKind("BareMetalHost")
		doc, err := bundle.SelectOne(selector)
		require.NoError(err)

		_, _, err = document.GetBMHBMCCredentials(doc, bundle)
		require.Error(err)
	})
}
