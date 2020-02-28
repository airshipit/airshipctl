package remote

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/environment"

	"opendev.org/airship/airshipctl/pkg/remote/redfish"
)

func initSettings(t *testing.T, rd *config.RemoteDirect, testdata string) *environment.AirshipCTLSettings {
	settings := &environment.AirshipCTLSettings{}
	settings.SetConfig(config.DummyConfig())
	bi, err := settings.Config().CurrentContextBootstrapInfo()
	require.NoError(t, err)
	bi.RemoteDirect = rd
	cm, err := settings.Config().CurrentContextManifest()
	require.NoError(t, err)
	cm.TargetPath = "testdata/" + testdata
	return settings
}

func TestUnknownRemoteType(t *testing.T) {
	s := initSettings(
		t,
		&config.RemoteDirect{
			RemoteType: "new-remote",
			IsoURL:     "/test.iso",
		},
		"base",
	)

	err := DoRemoteDirect(s)

	_, ok := err.(*GenericError)
	assert.True(t, ok)
}

func TestRedfishRemoteDirectWithEmptyURL(t *testing.T) {
	s := initSettings(
		t,
		&config.RemoteDirect{
			RemoteType: "redfish",
			IsoURL:     "/test.iso",
		},
		"emptyurl",
	)

	err := DoRemoteDirect(s)

	_, ok := err.(redfish.ErrRedfishMissingConfig)
	assert.True(t, ok)
}

func TestRedfishRemoteDirectWithEmptyIsoPath(t *testing.T) {
	s := initSettings(
		t,
		&config.RemoteDirect{
			RemoteType: "redfish",
			IsoURL:     "",
		},
		"base",
	)

	err := DoRemoteDirect(s)

	_, ok := err.(redfish.ErrRedfishMissingConfig)
	assert.True(t, ok)
}

func TestBootstrapRemoteDirectMissingConfigOpts(t *testing.T) {
	s := initSettings(
		t,
		nil,
		"base",
	)

	err := DoRemoteDirect(s)

	_, ok := err.(config.ErrMissingConfig)
	assert.True(t, ok)
}
