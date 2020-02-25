package initinfra_test

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/kustomize/v3/pkg/fs"

	"opendev.org/airship/airshipctl/pkg/cluster/initinfra"
	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/environment"
	"opendev.org/airship/airshipctl/pkg/k8s/client"
	"opendev.org/airship/airshipctl/pkg/k8s/kubectl"
	"opendev.org/airship/airshipctl/testutil/k8sutils"
)

type TestClient struct {
	MockKubectl   func() kubectl.Interface
	MockClientset func() kubernetes.Interface
}

func (tc TestClient) ClientSet() kubernetes.Interface { return tc.MockClientset() }
func (tc TestClient) Kubectl() kubectl.Interface      { return tc.MockKubectl() }

type TestKubectl struct {
	MockApply        func() error
	MockApplyOptions func() (*kubectl.ApplyOptions, error)
}

func (tk TestKubectl) Apply(docs []document.Document, ao *kubectl.ApplyOptions) error {
	return tk.MockApply()
}
func (tk TestKubectl) ApplyOptions() (*kubectl.ApplyOptions, error) {
	return tk.MockApplyOptions()
}

const (
	kubeconfigPath    = "testdata/kubeconfig.yaml"
	filenameRC        = "testdata/replicationcontroller.yaml"
	airshipConfigFile = "testdata/config.yaml"
)

var (
	DynamicClientError = errors.New("DynamicClientError")
)

func TestNewInfra(t *testing.T) {
	rs := makeNewFakeRootSettings(t, kubeconfigPath, airshipConfigFile)
	infra := initinfra.NewInfra(rs)

	assert.NotNil(t, infra.RootSettings)
}

func TestDeploy(t *testing.T) {
	rs := makeNewFakeRootSettings(t, kubeconfigPath, airshipConfigFile)
	tf := k8sutils.NewFakeFactoryForRC(t, filenameRC)
	defer tf.Cleanup()

	infra := initinfra.NewInfra(rs)
	infra.ClusterType = "ephemeral"
	infra.DryRun = true

	infra.FileSystem = kubectl.Buffer{FileSystem: fs.MakeRealFS()}

	kctl := kubectl.NewKubectl(tf)
	tc := TestClient{
		MockKubectl: func() kubectl.Interface { return kctl },
	}

	tests := []struct {
		theInfra      *initinfra.Infra
		client        client.Interface
		prune         bool
		expectedError error
	}{
		{
			client: TestClient{
				MockKubectl: func() kubectl.Interface {
					return kubectl.NewKubectl(k8sutils.
						NewMockKubectlFactory().
						WithDynamicClientByError(nil, DynamicClientError))
				},
			},
			expectedError: DynamicClientError,
		},
		{
			expectedError: nil,
			prune:         false,
			client:        tc,
		},
		{
			expectedError: nil,
			prune:         true,
			client:        tc,
		},
	}

	for _, test := range tests {
		infra.Prune = test.prune
		infra.Client = test.client
		actualErr := infra.Deploy()
		assert.Equal(t, test.expectedError, actualErr)
	}
}

// MakeNewFakeRootSettings takes kubeconfig path and directory path to fixture dir as argument.
func makeNewFakeRootSettings(t *testing.T, kp string, dir string) *environment.AirshipCTLSettings {
	t.Helper()
	rs := &environment.AirshipCTLSettings{}

	akp, err := filepath.Abs(kp)
	require.NoError(t, err)

	adir, err := filepath.Abs(dir)
	require.NoError(t, err)

	rs.SetAirshipConfigPath(adir)
	rs.SetKubeConfigPath(akp)

	rs.InitConfig()
	return rs
}
