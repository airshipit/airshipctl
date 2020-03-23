package client_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"opendev.org/airship/airshipctl/pkg/environment"
	"opendev.org/airship/airshipctl/pkg/k8s/client"
	"opendev.org/airship/airshipctl/testutil"
)

const (
	kubeconfigPath   = "testdata/kubeconfig.yaml"
	airshipConfigDir = "testdata"
)

func TestNewClient(t *testing.T) {
	conf, cleanup := testutil.InitConfig(t)
	defer cleanup(t)

	akp, err := filepath.Abs(kubeconfigPath)
	require.NoError(t, err)

	adir, err := filepath.Abs(airshipConfigDir)
	require.NoError(t, err)

	settings := &environment.AirshipCTLSettings{
		Config:            conf,
		AirshipConfigPath: adir,
		KubeConfigPath:    akp,
	}

	client, err := client.NewClient(settings)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.NotNil(t, client.ClientSet())
	assert.NotNil(t, client.Kubectl())
}
