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

package kubectl_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"

	"opendev.org/airship/airshipctl/pkg/k8s/kubectl"
	k8stest "opendev.org/airship/airshipctl/testutil/k8sutils"
)

var (
	filenameRC = "testdata/replicationcontroller.yaml"

	testStreams           = genericclioptions.NewTestIOStreamsDiscard()
	ErrToDiscoveryError   = errors.New("ErrToDiscoveryError")
	ErrDynamicClientError = errors.New("ErrDynamicClientError")
	ErrValidateError      = errors.New("ErrValidateError")
	ErrToRESTMapperError  = errors.New("ErrToRESTMapperError")
	ErrNamespaceError     = errors.New("ErrNamespaceError")
)

func TestNewApplyOptionsFactoryFailures(t *testing.T) {
	tests := []struct {
		f             cmdutil.Factory
		expectedError error
	}{
		{
			f:             k8stest.NewMockKubectlFactory().WithToDiscoveryClientByError(nil, ErrToDiscoveryError),
			expectedError: ErrToDiscoveryError,
		},
		{
			f:             k8stest.NewMockKubectlFactory().WithDynamicClientByError(nil, ErrDynamicClientError),
			expectedError: ErrDynamicClientError,
		},
		{
			f:             k8stest.NewMockKubectlFactory().WithValidatorByError(nil, ErrValidateError),
			expectedError: ErrValidateError,
		},
		{
			f:             k8stest.NewMockKubectlFactory().WithToRESTMapperByError(nil, ErrToRESTMapperError),
			expectedError: ErrToRESTMapperError,
		},
		{
			f: k8stest.NewMockKubectlFactory().
				WithToRawKubeConfigLoaderByError(k8stest.
					NewMockClientConfig().
					WithNamespace("", false, ErrNamespaceError)),
			expectedError: ErrNamespaceError,
		},
	}
	for _, test := range tests {
		_, err := kubectl.NewApplyOptions(test.f, testStreams)
		assert.Equal(t, err, test.expectedError)
	}
}
