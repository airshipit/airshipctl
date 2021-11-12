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

	"opendev.org/airship/airshipctl/pkg/api/v1alpha1"
	"opendev.org/airship/airshipctl/pkg/document"
)

type selectors struct {
	userDataSelector      v1alpha1.Selector
	userDataKey           string
	networkConfigSelector v1alpha1.Selector
	networkConfigKey      string
}

var (
	emptySelectors = selectors{
		userDataSelector:      v1alpha1.Selector{},
		networkConfigSelector: v1alpha1.Selector{},
	}
	validSelectors = selectors{
		userDataSelector: v1alpha1.Selector{
			ResID: v1alpha1.ResID{
				Gvk: v1alpha1.Gvk{
					Kind: secret,
				},
			},
			LabelSelector: ephUser,
		},
		userDataKey: defaultUserDataKey,
		networkConfigSelector: v1alpha1.Selector{
			ResID: v1alpha1.ResID{
				Gvk: v1alpha1.Gvk{
					Kind: bmh,
				},
			},
			LabelSelector: ephNode,
		},
		networkConfigKey: defaultNetworkConfigKey,
	}
)

func TestGetCloudData(t *testing.T) {
	tests := []struct {
		name string
		selectors
		labelFilter      string
		testBundle       document.Bundle
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
			testBundle: createTestBundle(t, testData{
				docsCfg:           getValidDocSet(),
				secret:            "Secret",
				bmh:               "BareMetalHost",
				ephNode:           "airshipit.org/ephemeral-node in (True, true)",
				ephUser:           "airshipit.org/ephemeral-user-data in (True, true)",
				selectorNamespace: "metal3",
				selectorName:      "master-1-networkdata",
				mockErr1:          nil,
				mockErr2:          nil,
				mockErr3:          nil,
			}),
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
					ByKind(bmh),
			},
			testBundle: createTestBundle(t, testData{
				docsCfg:           getEphemeralMissing(),
				secret:            "Secret",
				bmh:               "BareMetalHost",
				ephNode:           "airshipit.org/ephemeral-node in (True, true)",
				ephUser:           "airshipit.org/ephemeral-user-data in (True, true)",
				selectorNamespace: "",
				selectorName:      "",
				mockErr1:          nil,
				mockErr2: document.ErrDocNotFound{
					Selector: document.NewSelector().
						ByLabel(document.EphemeralHostSelector).
						ByKind(bmh),
				},
				mockErr3: nil,
			}),
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
					ByKind(bmh),
			},
			testBundle: createTestBundle(t, testData{
				docsCfg:           getEphemeralDuplicate(),
				secret:            "Secret",
				bmh:               "BareMetalHost",
				ephNode:           "airshipit.org/ephemeral-node in (True, true)",
				ephUser:           "airshipit.org/ephemeral-user-data in (True, true)",
				selectorNamespace: "",
				selectorName:      "",
				mockErr1:          nil,
				mockErr2: document.ErrMultiDocsFound{
					Selector: document.NewSelector().
						ByLabel(document.EphemeralHostSelector).
						ByKind(bmh),
				},
				mockErr3: nil,
			}),
		},
		{
			name:             "Bad network data document reference",
			labelFilter:      "test=networkdatabadpointer",
			selectors:        validSelectors,
			expectedUserData: nil,
			expectedNetData:  nil,
			expectedErr: document.ErrDocNotFound{
				Selector: document.NewSelector().
					ByKind(secret).
					ByNamespace("networkdatabadpointer-missing").
					ByName("networkdatabadpointer-missing"),
			},
			testBundle: createTestBundle(t, testData{
				docsCfg:           getNetworkBadPointer(),
				secret:            "Secret",
				bmh:               "BareMetalHost",
				ephNode:           "airshipit.org/ephemeral-node in (True, true)",
				ephUser:           "airshipit.org/ephemeral-user-data in (True, true)",
				selectorNamespace: "metal3",
				selectorName:      "master-1-networkdata",
				mockErr1:          nil,
				mockErr2:          nil,
				mockErr3: document.ErrDocNotFound{
					Selector: document.NewSelector().
						ByKind(secret).
						ByNamespace("networkdatabadpointer-missing").
						ByName("networkdatabadpointer-missing"),
				},
			}),
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
			testBundle: createTestBundle(t, testData{
				docsCfg:           getNetDataMalformed(),
				secret:            "Secret",
				bmh:               "BareMetalHost",
				ephNode:           "airshipit.org/ephemeral-node in (True, true)",
				ephUser:           "airshipit.org/ephemeral-user-data in (True, true)",
				selectorNamespace: "metal3",
				selectorName:      "master-1-networkdata",
				mockErr1:          nil,
				mockErr2:          nil,
				mockErr3:          nil,
			}),
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
			testBundle: createTestBundle(t, testData{
				docsCfg:           getUserDataMalformed(),
				secret:            "Secret",
				bmh:               "BareMetalHost",
				ephNode:           "airshipit.org/ephemeral-node in (True, true)",
				ephUser:           "airshipit.org/ephemeral-user-data in (True, true)",
				selectorNamespace: "",
				selectorName:      "",
				mockErr1:          nil,
				mockErr2:          nil,
				mockErr3:          nil,
			}),
		},
		{
			name:             "User data document not found",
			labelFilter:      "test=userdatamissing",
			selectors:        validSelectors,
			expectedUserData: nil,
			expectedNetData:  nil,
			expectedErr: document.ErrDocNotFound{
				Selector: document.NewSelector().
					ByKind(secret).
					ByLabel(document.EphemeralUserDataSelector),
			},
			testBundle: createTestBundle(t, testData{
				docsCfg:           getUserDataMissing(),
				secret:            "Secret",
				bmh:               "BareMetalHost",
				ephNode:           "airshipit.org/ephemeral-node in (True, true)",
				ephUser:           "airshipit.org/ephemeral-user-data in (True, true)",
				selectorNamespace: "",
				selectorName:      "",
				mockErr1: document.ErrDocNotFound{
					Selector: document.NewSelector().
						ByKind(secret).
						ByLabel(document.EphemeralUserDataSelector),
				},
				mockErr2: nil,
				mockErr3: nil,
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// ensure each test case filter has at least one document
			docs, err := tt.testBundle.GetAllDocuments()
			require.NoError(t, err, "GetAllDocuments failed")
			require.NotZero(t, docs)

			actualUserData, actualNetData, actualErr := GetCloudData(
				tt.testBundle,
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
