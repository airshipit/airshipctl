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

package cluster_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"opendev.org/airship/airshipctl/cmd/cluster"
	pkgcluster "opendev.org/airship/airshipctl/pkg/cluster"
	"opendev.org/airship/airshipctl/testutil"
)

func TestNewKubeConfigCommandCmd(t *testing.T) {
	tests := []*testutil.CmdTest{
		{
			Name:    "cluster-get-kubeconfig-cmd-with-help",
			CmdLine: "--help",
			Cmd:     cluster.NewGetKubeconfigCommand(nil),
		},
	}
	for _, testcase := range tests {
		testutil.RunTest(t, testcase)
	}
}

func TestGetKubeconfArgs(t *testing.T) {
	tests := []struct {
		name                 string
		args                 []string
		expectedErrStr       string
		expectedClusterNames []string
	}{
		{
			name:                 "success one cluster specified",
			args:                 []string{"cluster01"},
			expectedClusterNames: []string{"cluster01"},
		},
		{
			name: "success no cluster specified",
		},
		{
			name:                 "success two cluster specified",
			args:                 []string{"cluster01", "cluster02"},
			expectedClusterNames: []string{"cluster01", "cluster02"},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			cmd := &pkgcluster.GetKubeconfigCommand{}
			args := cluster.GetKubeconfArgs(cmd)
			err := args(cluster.NewGetKubeconfigCommand(nil), tt.args)
			if tt.expectedErrStr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrStr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedClusterNames, cmd.ClusterNames)
			}
		})
	}
}
