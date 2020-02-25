package kubectl_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"

	"opendev.org/airship/airshipctl/pkg/k8s/kubectl"
	k8stest "opendev.org/airship/airshipctl/testutil/k8sutils"
)

var (
	filenameRC = "testdata/replicationcontroller.yaml"

	testStreams        = genericclioptions.NewTestIOStreamsDiscard()
	ToDiscoveryError   = errors.New("ToDiscoveryError")
	DynamicClientError = errors.New("DynamicClientError")
	ValidateError      = errors.New("ValidateError")
	ToRESTMapperError  = errors.New("ToRESTMapperError")
	NamespaceError     = errors.New("NamespaceError")
)

func TestApplyOptionsRun(t *testing.T) {
	f := k8stest.NewFakeFactoryForRC(t, filenameRC)
	defer f.Cleanup()

	streams := genericclioptions.NewTestIOStreamsDiscard()

	aa, err := kubectl.NewApplyOptions(f, streams)
	require.NoError(t, err, "Could not build ApplyAdapter")
	aa.SetDryRun(true)
	aa.SetSourceFiles([]string{filenameRC})
	assert.NoError(t, aa.Run())
}

func TestNewApplyOptionsFactoryFailures(t *testing.T) {
	tests := []struct {
		f             cmdutil.Factory
		expectedError error
	}{
		{
			f:             k8stest.NewMockKubectlFactory().WithToDiscoveryClientByError(nil, ToDiscoveryError),
			expectedError: ToDiscoveryError,
		},
		{
			f:             k8stest.NewMockKubectlFactory().WithDynamicClientByError(nil, DynamicClientError),
			expectedError: DynamicClientError,
		},
		{
			f:             k8stest.NewMockKubectlFactory().WithValidatorByError(nil, ValidateError),
			expectedError: ValidateError,
		},
		{
			f:             k8stest.NewMockKubectlFactory().WithToRESTMapperByError(nil, ToRESTMapperError),
			expectedError: ToRESTMapperError,
		},
		{
			f: k8stest.NewMockKubectlFactory().
				WithToRawKubeConfigLoaderByError(k8stest.
					NewMockClientConfig().
					WithNamespace("", false, NamespaceError)),
			expectedError: NamespaceError,
		},
	}
	for _, test := range tests {
		_, err := kubectl.NewApplyOptions(test.f, testStreams)
		assert.Equal(t, err, test.expectedError)
	}
}
