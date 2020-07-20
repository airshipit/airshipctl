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

package apply_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/environment"
	"opendev.org/airship/airshipctl/pkg/k8s/applier"
	"opendev.org/airship/airshipctl/pkg/phase/apply"
	"opendev.org/airship/airshipctl/testutil"
	"opendev.org/airship/airshipctl/testutil/k8sutils"
)

const (
	kubeconfigPath    = "testdata/kubeconfig.yaml"
	airshipConfigFile = "testdata/config.yaml"
)

func TestDeploy(t *testing.T) {
	bundle := testutil.NewTestBundle(t, "testdata/primary/site/test-site/ephemeral/initinfra")
	replicationController, err := bundle.SelectOne(document.NewSelector().ByKind("ReplicationController"))
	require.NoError(t, err)
	b, err := replicationController.AsYAML()
	require.NoError(t, err)
	f := k8sutils.FakeFactory(t,
		[]k8sutils.ClientHandler{
			&k8sutils.InventoryObjectHandler{},
			&k8sutils.NamespaceHandler{},
			&k8sutils.GenericHandler{
				Obj:       &corev1.ReplicationController{},
				Bytes:     b,
				URLPath:   "/namespaces/%s/replicationcontrollers",
				Namespace: replicationController.GetNamespace(),
			},
		})
	defer f.Cleanup()
	tests := []struct {
		name                string
		expectedErrorString string
		cliApplier          *applier.Applier
		clusterPurposes     map[string]*config.ClusterPurpose
		phaseName           string
	}{
		{
			name:                "success",
			expectedErrorString: "",
			cliApplier: applier.NewFakeApplier(genericclioptions.IOStreams{
				In:     os.Stdin,
				Out:    os.Stdout,
				ErrOut: os.Stderr,
			}, k8sutils.SuccessEvents(), f),
		},
		{
			name:                "missing clusters",
			expectedErrorString: "At least one cluster needs to be defined",
			clusterPurposes:     map[string]*config.ClusterPurpose{},
		},
		{
			name:                "missing phase",
			expectedErrorString: "Phase document 'missingPhase' was not found",
			phaseName:           "missingPhase",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			rs := makeNewFakeRootSettings(t, kubeconfigPath, airshipConfigFile)
			ao := &apply.Options{
				RootSettings: rs,
			}
			ao.Initialize()
			ao.PhaseName = "initinfra"
			ao.DryRun = true
			if tt.cliApplier != nil {
				ao.Applier = tt.cliApplier
			}
			if tt.clusterPurposes != nil {
				ao.RootSettings.Config.Clusters = tt.clusterPurposes
			}
			if tt.phaseName != "" {
				ao.PhaseName = tt.phaseName
			}
			actualErr := ao.Run()
			if tt.expectedErrorString != "" {
				require.Error(t, actualErr)
				assert.Contains(t, actualErr.Error(), tt.expectedErrorString)
			} else {
				assert.NoError(t, actualErr)
			}
		})
	}
}

// makeNewFakeRootSettings takes kubeconfig path and directory path to fixture dir as argument.
func makeNewFakeRootSettings(t *testing.T, kp string, dir string) *environment.AirshipCTLSettings {
	t.Helper()

	akp, err := filepath.Abs(kp)
	require.NoError(t, err)

	adir, err := filepath.Abs(dir)
	require.NoError(t, err)

	settings := &environment.AirshipCTLSettings{
		AirshipConfigPath: adir,
		KubeConfigPath:    akp,
	}

	settings.InitConfig()
	return settings
}
