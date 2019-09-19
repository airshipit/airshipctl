package remote

import (
	"testing"

	"github.com/stretchr/testify/assert"

	redfish "opendev.org/airship/airshipctl/pkg/remote/redfish"
)

func TestUnknownRemoteType(t *testing.T) {

	rdCfg := RemoteDirectConfig{
		RemoteType:      "new-remote",
		RemoteURL:       "http://localhost:8000",
		EphemeralNodeId: "test-node",
		IsoPath:         "/test.iso",
	}

	err := DoRemoteDirect(rdCfg)

	_, ok := err.(*RemoteDirectError)
	assert.True(t, ok)
}

func TestRedfishRemoteDirectWithBogusConfig(t *testing.T) {

	rdCfg := RemoteDirectConfig{
		RemoteType:      "redfish",
		RemoteURL:       "http://nolocalhost:8888",
		EphemeralNodeId: "test-node",
		IsoPath:         "/test.iso",
	}

	err := DoRemoteDirect(rdCfg)

	_, ok := err.(*redfish.RedfishClientError)
	assert.True(t, ok)
}

func TestRedfishRemoteDirectWithEmptyURL(t *testing.T) {

	rdCfg := RemoteDirectConfig{
		RemoteType:      "redfish",
		RemoteURL:       "",
		EphemeralNodeId: "test-node",
		IsoPath:         "/test.iso",
	}

	err := DoRemoteDirect(rdCfg)

	_, ok := err.(*redfish.RedfishConfigError)
	assert.True(t, ok)
}

func TestRedfishRemoteDirectWithEmptyNodeID(t *testing.T) {

	rdCfg := RemoteDirectConfig{
		RemoteType:      "redfish",
		RemoteURL:       "http://nolocalhost:8888",
		EphemeralNodeId: "",
		IsoPath:         "/test.iso",
	}

	err := DoRemoteDirect(rdCfg)

	_, ok := err.(*redfish.RedfishConfigError)
	assert.True(t, ok)
}

func TestRedfishRemoteDirectWithEmptyIsoPath(t *testing.T) {

	rdCfg := RemoteDirectConfig{
		RemoteType:      "redfish",
		RemoteURL:       "http://nolocalhost:8888",
		EphemeralNodeId: "123",
		IsoPath:         "",
	}

	err := DoRemoteDirect(rdCfg)

	_, ok := err.(*redfish.RedfishConfigError)
	assert.True(t, ok)
}
