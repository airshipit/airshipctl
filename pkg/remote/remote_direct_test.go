package remote

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/environment"
	"opendev.org/airship/airshipctl/pkg/remote/redfish"
	"opendev.org/airship/airshipctl/testutil"
	"opendev.org/airship/airshipctl/testutil/redfishutils"
)

const (
	systemID   = "server-100"
	isoURL     = "https://localhost:8080/ubuntu.iso"
	redfishURL = "https://redfish.local"
)

func initSettings(t *testing.T, rd *config.RemoteDirect, testdata string) *environment.AirshipCTLSettings {
	t.Helper()
	settings := &environment.AirshipCTLSettings{}
	settings.SetConfig(testutil.DummyConfig())
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

	_, err := NewAdapter(s)
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

	_, err := NewAdapter(s)
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

	_, err := NewAdapter(s)
	_, ok := err.(redfish.ErrRedfishMissingConfig)
	assert.True(t, ok)
}

func TestBootstrapRemoteDirectMissingConfigOpts(t *testing.T) {
	s := initSettings(
		t,
		nil,
		"base",
	)

	_, err := NewAdapter(s)
	_, ok := err.(config.ErrMissingConfig)
	assert.True(t, ok)
}

func TestDoRemoteDirectRedfish(t *testing.T) {
	cfg := &config.RemoteDirect{
		RemoteType: redfish.ClientType,
		IsoURL:     isoURL,
	}

	// Initialize a remote direct adapter
	settings := initSettings(t, cfg, "base")
	a, err := NewAdapter(settings)
	assert.NoError(t, err)

	ctx, rMock, err := redfishutils.NewClient(systemID, isoURL, redfishURL, false, false, "admin", "password")
	assert.NoError(t, err)

	rMock.On("SetVirtualMedia", a.Context, isoURL).Times(1).Return(nil)
	rMock.On("SetEphemeralBootSourceByType", a.Context).Times(1).Return(nil)
	rMock.On("EphemeralNodeID").Times(1).Return(systemID)
	rMock.On("RebootSystem", a.Context, systemID).Times(1).Return(nil)

	// Swap the redfish client initialized by the remote direct adapter with the above mocked client
	a.Context = ctx
	a.OOBClient = rMock

	err = a.DoRemoteDirect()
	assert.NoError(t, err)
}

func TestDoRemoteDirectRedfishVirtualMediaError(t *testing.T) {
	cfg := &config.RemoteDirect{
		RemoteType: redfish.ClientType,
		IsoURL:     isoURL,
	}

	// Initialize a remote direct adapter
	settings := initSettings(t, cfg, "base")
	a, err := NewAdapter(settings)
	assert.NoError(t, err)

	ctx, rMock, err := redfishutils.NewClient(systemID, isoURL, redfishURL, false, false, "admin", "password")
	assert.NoError(t, err)

	expectedErr := redfish.ErrRedfishClient{Message: "Unable to set virtual media."}
	rMock.On("SetVirtualMedia", a.Context, isoURL).Times(1).Return(expectedErr)
	rMock.On("SetEphemeralBootSourceByType", a.Context).Times(1).Return(nil)
	rMock.On("EphemeralNodeID").Times(1).Return(systemID)
	rMock.On("RebootSystem", a.Context, systemID).Times(1).Return(nil)

	// Swap the redfish client initialized by the remote direct adapter with the above mocked client
	a.Context = ctx
	a.OOBClient = rMock

	err = a.DoRemoteDirect()
	_, ok := err.(redfish.ErrRedfishClient)
	assert.True(t, ok)
}

func TestDoRemoteDirectRedfishBootSourceError(t *testing.T) {
	cfg := &config.RemoteDirect{
		RemoteType: redfish.ClientType,
		IsoURL:     isoURL,
	}

	// Initialize a remote direct adapter
	settings := initSettings(t, cfg, "base")
	a, err := NewAdapter(settings)
	assert.NoError(t, err)

	ctx, rMock, err := redfishutils.NewClient(systemID, isoURL, redfishURL, false, false, "admin", "password")
	assert.NoError(t, err)

	rMock.On("SetVirtualMedia", a.Context, isoURL).Times(1).Return(nil)

	expectedErr := redfish.ErrRedfishClient{Message: "Unable to set boot source."}
	rMock.On("SetEphemeralBootSourceByType", a.Context).Times(1).Return(expectedErr)
	rMock.On("EphemeralNodeID").Times(1).Return(systemID)
	rMock.On("RebootSystem", a.Context, systemID).Times(1).Return(nil)

	// Swap the redfish client initialized by the remote direct adapter with the above mocked client
	a.Context = ctx
	a.OOBClient = rMock

	err = a.DoRemoteDirect()
	_, ok := err.(redfish.ErrRedfishClient)
	assert.True(t, ok)
}

func TestDoRemoteDirectRedfishRebootError(t *testing.T) {
	cfg := &config.RemoteDirect{
		RemoteType: redfish.ClientType,
		IsoURL:     isoURL,
	}

	// Initialize a remote direct adapter
	settings := initSettings(t, cfg, "base")
	a, err := NewAdapter(settings)
	assert.NoError(t, err)

	ctx, rMock, err := redfishutils.NewClient(systemID, isoURL, redfishURL, false, false, "admin", "password")
	assert.NoError(t, err)

	rMock.On("SetVirtualMedia", a.Context, isoURL).Times(1).Return(nil)
	rMock.On("SetEphemeralBootSourceByType", a.Context).Times(1).Return(nil)
	rMock.On("EphemeralNodeID").Times(1).Return(systemID)

	expectedErr := redfish.ErrRedfishClient{Message: "Unable to set boot source."}
	rMock.On("RebootSystem", a.Context, systemID).Times(1).Return(expectedErr)

	// Swap the redfish client initialized by the remote direct adapter with the above mocked client
	a.Context = ctx
	a.OOBClient = rMock

	err = a.DoRemoteDirect()
	_, ok := err.(redfish.ErrRedfishClient)
	assert.True(t, ok)
}
