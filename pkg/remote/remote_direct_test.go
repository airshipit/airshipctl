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

package remote

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"opendev.org/airship/airshipctl/testutil/redfishutils"
)

const (
	redfishURL = "redfish+https://localhost:2344/Systems/System.Embedded.1"
	username   = "admin"
	password   = "password"
)

func TestDoRemoteDirectMissingConfigOpts(t *testing.T) {
	rMock, err := redfishutils.NewClient(redfishURL, false, false, username, password)
	assert.NoError(t, err)

	ephemeralHost := baremetalHost{
		rMock,
		redfishURL,
		"doc-name",
		username,
		password,
	}

	settings := initSettings(t, withTestDataPath("noremote"))
	// there must be document.ErrDocNotFound
	err = ephemeralHost.DoRemoteDirect(settings)
	expectedErrorMessage := `document filtered by selector [Group="airshipit.org", Version="v1alpha1", ` +
		`Kind="RemoteDirectConfiguration"] found no documents`
	assert.Equal(t, expectedErrorMessage, fmt.Sprintf("%s", err))
	assert.Error(t, err)
}

func TestDoRemoteDirectRedfish(t *testing.T) {
	rMock, err := redfishutils.NewClient(redfishURL, false, false, username, password)
	require.NoError(t, err)

	rMock.On("RemoteDirect").Times(1).Return(nil)

	ephemeralHost := baremetalHost{
		rMock,
		redfishURL,
		"doc-name",
		username,
		password,
	}

	settings := initSettings(t, withTestDataPath("base"))
	err = ephemeralHost.DoRemoteDirect(settings)
	assert.NoError(t, err)
}

func TestDoRemoteDirectError(t *testing.T) {
	rMock, err := redfishutils.NewClient(redfishURL, false, false, username, password)
	require.NoError(t, err)

	expectedErr := fmt.Errorf("remote direct error")
	rMock.On("RemoteDirect").Times(1).Return(expectedErr)

	ephemeralHost := baremetalHost{
		rMock,
		redfishURL,
		"doc-name",
		username,
		password,
	}

	settings := initSettings(t, withTestDataPath("base"))
	actualErr := ephemeralHost.DoRemoteDirect(settings)
	assert.Equal(t, expectedErr, actualErr)
}
