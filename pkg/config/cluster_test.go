/*
Copyright 2014 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/testutil"
)

func TestPrettyString(t *testing.T) {
	fSys := testutil.SetupTestFs(t, "testdata")
	data, err := fSys.ReadFile("/prettycluster-string.yaml")
	require.NoError(t, err)

	cluster := testutil.DummyCluster()
	assert.EqualValues(t, cluster.PrettyString(), string(data))
}

func TestValidClusterTypeFail(t *testing.T) {
	err := config.ValidClusterType("Fake")
	assert.Error(t, err)
}

func TestGetCluster(t *testing.T) {
	conf, cleanup := testutil.InitConfig(t)
	defer cleanup(t)

	cluster, err := conf.GetCluster("def", config.Ephemeral)
	require.NoError(t, err)

	// Test Positives
	assert.EqualValues(t, cluster.NameInKubeconf, "def_ephemeral")
	assert.EqualValues(t, cluster.KubeCluster().Server, "http://5.6.7.8")

	// Test Wrong Cluster
	_, err = conf.GetCluster("unknown", config.Ephemeral)
	assert.Error(t, err)

	// Test Wrong Cluster Type
	_, err = conf.GetCluster("def", "Unknown")
	assert.Error(t, err)
}

func TestAddCluster(t *testing.T) {
	conf, cleanup := testutil.InitConfig(t)
	defer cleanup(t)

	co := testutil.DummyClusterOptions()
	cluster, err := conf.AddCluster(co)
	require.NoError(t, err)

	assert.EqualValues(t, conf.Clusters[co.Name].ClusterTypes[co.ClusterType], cluster)
}
