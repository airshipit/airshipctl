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
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"opendev.org/airship/airshipctl/pkg/document"
	testdoc "opendev.org/airship/airshipctl/testutil/document"
)

const (
	bmhConfig = `apiVersion: metal3.io/v1alpha1
kind: BareMetalHost
metadata:
  labels:
    airshipit.org/ephemeral-node: "true"
  name: master-0
spec:
  online: true
  bootMACAddress: 00:3b:8b:0c:ec:8b
  bmc:
    address: redfish+https://192.168.111.1/v1/Redfish/Foo/Bar
    credentialsName: master-0-bmc
  networkData:
    name: master-0-networkdata
    namespace: metal3
`
	networkConfig = `apiVersion: v1
kind: Secret
metadata:
  name: master-0-networkdata
  namespace: metal3
type: Opaque
data:
  networkData: c29tZSBuZXR3b3JrIGRhdGEK
`
	credsConfig = `apiVersion: v1
kind: Secret
metadata:
  name: master-0-bmc
  namespace: metal3
type: Opaque
stringData:
  username: username
  password: password
`
	bmhConfigInvalid = `apiVersion: metal3.io/v1alpha1
kind: BareMetalHost
metadata:
  labels:
    airshipit.org/ephemeral-node: "true"
  name: master-0
spec:
  online: true
  bootMACAddress: 00:3b:8b:0c:ec:8b
  bmc:
    credentialsName: master-0-bmc
  networkData:
    name: master-0-networkdata
    namespace: metal3
`
	networkConfigInvalid = `apiVersion: v1
kind: Secret
metadata:
  name: master-0-networkdata
  namespace: metal3
type: Opaque
`
	credsConfigInvalid = `apiVersion: v1
kind: Secret
metadata:
  name: master-0-bmc-invalid
  namespace: metal3
type: Opaque
stringData:
  username: username
  password: password
`
)

func testBMHValidBundle(t *testing.T) document.Bundle {
	bundle := &testdoc.MockBundle{}
	bmhConfigDoc, err := document.NewDocumentFromBytes([]byte(bmhConfig))
	require.NoError(t, err)
	networkSecretDoc, err := document.NewDocumentFromBytes([]byte(networkConfig))
	require.NoError(t, err)
	credsSecretDoc, err := document.NewDocumentFromBytes([]byte(credsConfig))
	require.NoError(t, err)

	bundle.On("SelectOne", mock.MatchedBy(func(selector document.Selector) bool {
		return selector.Kind == "BareMetalHost"
	})).
		Return(bmhConfigDoc, nil)
	bundle.On("SelectOne", mock.MatchedBy(func(selector document.Selector) bool {
		return selector.Kind == "Secret" && selector.Name == "master-0-networkdata"
	})).
		Return(networkSecretDoc, nil)
	bundle.On("SelectOne", mock.MatchedBy(func(selector document.Selector) bool {
		return selector.Kind == "Secret" && selector.Name == "master-0-bmc"
	})).
		Return(credsSecretDoc, nil)

	return bundle
}

func testBMHInvalidBundle(t *testing.T) document.Bundle {
	bundle := &testdoc.MockBundle{}
	bmhConfigDocInvalid, err := document.NewDocumentFromBytes([]byte(bmhConfigInvalid))
	require.NoError(t, err)
	networkSecretDocInvalid, err := document.NewDocumentFromBytes([]byte(networkConfigInvalid))
	require.NoError(t, err)
	credsSecretDocInvalid, err := document.NewDocumentFromBytes([]byte(credsConfigInvalid))
	require.NoError(t, err)

	bundle.On("SelectOne", mock.MatchedBy(func(selector document.Selector) bool {
		return selector.Kind == "BareMetalHost"
	})).
		Return(bmhConfigDocInvalid, nil)
	bundle.On("SelectOne", mock.MatchedBy(func(selector document.Selector) bool {
		return selector.Kind == "Secret" && selector.Name == "master-0-networkdata"
	})).
		Return(networkSecretDocInvalid, nil)
	bundle.On("SelectOne", mock.MatchedBy(func(selector document.Selector) bool {
		return selector.Kind == "Secret" && selector.Name == "master-0-bmc"
	})).
		Return(credsSecretDocInvalid, errors.New("error"))

	return bundle
}

func TestDocHelpers(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	bundle := testBMHValidBundle(t)
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

	bundle := testBMHInvalidBundle(t)
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
