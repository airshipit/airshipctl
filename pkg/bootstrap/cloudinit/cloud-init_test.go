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
)

type selectors struct {
	userDataSelector      document.Selector
	userDataKey           string
	networkConfigSelector document.Selector
	networkConfigKey      string
}

var (
	emptySelectors = selectors{
		userDataSelector:      document.NewSelector(),
		networkConfigSelector: document.NewSelector(),
	}
	validSelectors = selectors{
		userDataSelector: document.NewSelector().
			ByKind("Secret").
			ByLabel("airshipit.org/ephemeral-user-data in (True, true)"),
		userDataKey: defaultUserDataKey,
		networkConfigSelector: document.NewSelector().
			ByKind("BareMetalHost").
			ByLabel("airshipit.org/ephemeral-node in (True, true)"),
		networkConfigKey: defaultNetworkConfigKey,
	}
)

func TestGetCloudData(t *testing.T) {
	bundle, err := document.NewBundleByPath("testdata")
	require.NoError(t, err, "Building Bundle Failed")

	tests := []struct {
		name string
		selectors
		labelFilter      string
		expectedUserData []byte
		expectedNetData  []byte
		expectedErr      error
	}{
		{
			name:             "Default selectors",
			labelFilter:      "test=validdocset",
			selectors:        emptySelectors,
			expectedUserData: []byte("cloud-init"),
			expectedNetData:  []byte("net-config"),
			expectedErr:      nil,
		},
		{
			name:             "BareMetalHost document not found",
			labelFilter:      "test=ephemeralmissing",
			selectors:        validSelectors,
			expectedUserData: nil,
			expectedNetData:  nil,
			expectedErr: document.ErrDocNotFound{
				Selector: document.NewSelector().
					ByLabel(document.EphemeralHostSelector).
					ByKind("BareMetalHost"),
			},
		},
		{
			name:             "BareMetalHost document duplication",
			labelFilter:      "test=ephemeralduplicate",
			selectors:        validSelectors,
			expectedUserData: nil,
			expectedNetData:  nil,
			expectedErr: document.ErrMultiDocsFound{
				Selector: document.NewSelector().
					ByLabel(document.EphemeralHostSelector).
					ByKind("BareMetalHost"),
			},
		},
		{
			name:             "Bad network data document reference",
			labelFilter:      "test=networkdatabadpointer",
			selectors:        validSelectors,
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
			name:             "Bad network data document structure",
			labelFilter:      "test=networkdatamalformed",
			selectors:        validSelectors,
			expectedUserData: nil,
			expectedNetData:  nil,
			expectedErr: document.ErrDocumentDataKeyNotFound{
				DocName: "networkdatamalformed-malformed",
				Key:     defaultNetworkConfigKey,
			},
		},
		{
			name:             "Bad user data document structure",
			labelFilter:      "test=userdatamalformed",
			selectors:        validSelectors,
			expectedUserData: nil,
			expectedNetData:  nil,
			expectedErr: document.ErrDocumentDataKeyNotFound{
				DocName: "userdatamalformed-somesecret",
				Key:     defaultUserDataKey,
			},
		},
		{
			name:             "User data document not found",
			labelFilter:      "test=userdatamissing",
			selectors:        validSelectors,
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
		t.Run(tt.name, func(t *testing.T) {
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
		})
	}
}
