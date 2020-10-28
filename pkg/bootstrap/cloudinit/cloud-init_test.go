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

package cloudinit

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"opendev.org/airship/airshipctl/pkg/document"

	"sigs.k8s.io/kustomize/api/resid"
	"sigs.k8s.io/kustomize/api/types"
)

func TestGetCloudData(t *testing.T) {
	bundle, err := document.NewBundleByPath("testdata")
	require.NoError(t, err, "Building Bundle Failed")

	tests := []struct {
		labelFilter           string
		userDataSelector      types.Selector
		userDataKey           string
		networkConfigSelector types.Selector
		networkConfigKey      string
		expectedUserData      []byte
		expectedNetData       []byte
		expectedErr           error
	}{
		{
			labelFilter:           "test=validdocset",
			userDataSelector:      types.Selector{},
			networkConfigSelector: types.Selector{},
			expectedUserData:      []byte("cloud-init"),
			expectedNetData:       []byte("net-config"),
			expectedErr:           nil,
		},
		{
			labelFilter: "test=ephemeralmissing",
			userDataSelector: types.Selector{
				Gvk:           resid.Gvk{Kind: "Secret"},
				LabelSelector: "airshipit.org/ephemeral-user-data in (True, true)",
			},
			userDataKey: "userData",
			networkConfigSelector: types.Selector{
				Gvk:           resid.Gvk{Kind: "BareMetalHost"},
				LabelSelector: "airshipit.org/ephemeral-node in (True, true)",
			},
			networkConfigKey: "networkData",
			expectedUserData: nil,
			expectedNetData:  nil,
			expectedErr: document.ErrDocNotFound{
				Selector: document.NewSelector().
					ByLabel(document.EphemeralHostSelector).
					ByKind("BareMetalHost"),
			},
		},
		{
			labelFilter: "test=ephemeralduplicate",
			userDataSelector: types.Selector{
				Gvk:           resid.Gvk{Kind: "Secret"},
				LabelSelector: "airshipit.org/ephemeral-user-data in (True, true)",
			},
			userDataKey: "userData",
			networkConfigSelector: types.Selector{
				Gvk:           resid.Gvk{Kind: "BareMetalHost"},
				LabelSelector: "airshipit.org/ephemeral-node in (True, true)",
			},
			networkConfigKey: "networkData",
			expectedUserData: nil,
			expectedNetData:  nil,
			expectedErr: document.ErrMultiDocsFound{
				Selector: document.NewSelector().
					ByLabel(document.EphemeralHostSelector).
					ByKind("BareMetalHost"),
			},
		},
		{
			labelFilter: "test=networkdatabadpointer",
			userDataSelector: types.Selector{
				Gvk:           resid.Gvk{Kind: "Secret"},
				LabelSelector: "airshipit.org/ephemeral-user-data in (True, true)",
			},
			userDataKey: "userData",
			networkConfigSelector: types.Selector{
				Gvk:           resid.Gvk{Kind: "BareMetalHost"},
				LabelSelector: "airshipit.org/ephemeral-node in (True, true)",
			},
			networkConfigKey: "networkData",
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
			labelFilter: "test=networkdatamalformed",
			userDataSelector: types.Selector{
				Gvk:           resid.Gvk{Kind: "Secret"},
				LabelSelector: "airshipit.org/ephemeral-user-data in (True, true)",
			},
			userDataKey: "userData",
			networkConfigSelector: types.Selector{
				Gvk:           resid.Gvk{Kind: "BareMetalHost"},
				LabelSelector: "airshipit.org/ephemeral-node in (True, true)",
			},
			networkConfigKey: "networkData",
			expectedUserData: nil,
			expectedNetData:  nil,
			expectedErr: document.ErrDataNotSupplied{
				DocName: "networkdatamalformed-malformed",
				Key:     networkConfigKeyDefault,
			},
		},
		{
			labelFilter: "test=userdatamalformed",
			userDataSelector: types.Selector{
				Gvk:           resid.Gvk{Kind: "Secret"},
				LabelSelector: "airshipit.org/ephemeral-user-data in (True, true)",
			},
			userDataKey: "userData",
			networkConfigSelector: types.Selector{
				Gvk:           resid.Gvk{Kind: "BareMetalHost"},
				LabelSelector: "airshipit.org/ephemeral-node in (True, true)",
			},
			networkConfigKey: "networkData",
			expectedUserData: nil,
			expectedNetData:  nil,
			expectedErr: document.ErrDataNotSupplied{
				DocName: "userdatamalformed-somesecret",
				Key:     userDataKeyDefault,
			},
		},
		{
			labelFilter: "test=userdatamissing",
			userDataSelector: types.Selector{
				Gvk:           resid.Gvk{Kind: "Secret"},
				LabelSelector: "airshipit.org/ephemeral-user-data in (True, true)",
			},
			userDataKey: "userData",
			networkConfigSelector: types.Selector{
				Gvk:           resid.Gvk{Kind: "BareMetalHost"},
				LabelSelector: "airshipit.org/ephemeral-node in (True, true)",
			},
			networkConfigKey: "networkData",
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

		actualUserData, actualNetData, actualErr := GetCloudData(
			filteredBundle,
			tt.userDataSelector,
			tt.userDataKey,
			tt.networkConfigSelector,
			tt.networkConfigKey,
		)

		assert.Equal(t, tt.expectedUserData, actualUserData)
		assert.Equal(t, tt.expectedNetData, actualNetData)
		assert.Equal(t, tt.expectedErr, actualErr)
	}
}
