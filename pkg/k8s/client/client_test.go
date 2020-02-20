package client

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"opendev.org/airship/airshipctl/pkg/environment"
)

const (
	kubeconfigPath   = "testdata/kubeconfig.yaml"
	airshipConfigDir = "testdata"
)

func TestNewclient(t *testing.T) {
	settings := &environment.AirshipCTLSettings{}
	settings.InitConfig()

	akp, err := filepath.Abs(kubeconfigPath)
	require.NoError(t, err)

	adir, err := filepath.Abs(airshipConfigDir)
	require.NoError(t, err)

	settings.SetAirshipConfigPath(adir)
	settings.SetKubeConfigPath(akp)

	client, err := NewClient(settings)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.NotNil(t, client.ClientSet())
	assert.NotNil(t, client.Kubectl())
}
