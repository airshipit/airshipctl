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

	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/testutil"
)

func TestRunSetAuthInfo(t *testing.T) {
	t.Run("testAddAuthInfo", func(t *testing.T) {
		conf := testutil.DummyConfig()
		dummyAuthInfoOptions := testutil.DummyAuthInfoOptions()
		dummyAuthInfoOptions.Name = "second_user"
		dummyAuthInfoOptions.Token = ""

		modified, err := config.RunSetAuthInfo(dummyAuthInfoOptions, conf, false)
		assert.NoError(t, err)
		assert.False(t, modified)
		assert.Contains(t, conf.AuthInfos, "second_user")
	})

	t.Run("testModifyAuthInfo", func(t *testing.T) {
		conf := testutil.DummyConfig()
		dummyAuthInfoOptions := testutil.DummyAuthInfoOptions()
		dummyAuthInfoOptions.Name = "dummy_user"
		dummyAuthInfoOptions.Password = "testpassword123"
		dummyAuthInfoOptions.Token = ""

		modified, err := config.RunSetAuthInfo(dummyAuthInfoOptions, conf, false)
		assert.NoError(t, err)
		assert.True(t, modified)
		authInfo, err := conf.GetAuthInfo("dummy_user")
		assert.NoError(t, err)
		assert.Equal(t, dummyAuthInfoOptions.Password, authInfo.KubeAuthInfo().Password)
	})
}

func TestRunSetCluster(t *testing.T) {
	t.Run("testAddCluster", func(t *testing.T) {
		conf := testutil.DummyConfig()
		dummyClusterOptions := testutil.DummyClusterOptions()
		dummyClusterOptions.Name = "second_cluster"

		modified, err := config.RunSetCluster(dummyClusterOptions, conf, false)
		assert.NoError(t, err)
		assert.False(t, modified)
		assert.Contains(t, conf.Clusters, "second_cluster")
	})

	t.Run("testModifyCluster", func(t *testing.T) {
		conf := testutil.DummyConfig()
		dummyClusterOptions := testutil.DummyClusterOptions()
		dummyClusterOptions.Server = "http://123.45.67.890"

		modified, err := config.RunSetCluster(dummyClusterOptions, conf, false)
		assert.NoError(t, err)
		assert.True(t, modified)
		assert.Equal(
			t, "http://123.45.67.890",
			conf.Clusters["dummy_cluster"].ClusterTypes["ephemeral"].KubeCluster().Server)
	})
}

func TestRunSetContext(t *testing.T) {
	t.Run("testAddContext", func(t *testing.T) {
		conf := testutil.DummyConfig()
		dummyContextOptions := testutil.DummyContextOptions()
		dummyContextOptions.Name = "second_context"

		modified, err := config.RunSetContext(dummyContextOptions, conf, false)
		assert.NoError(t, err)
		assert.False(t, modified)
		assert.Contains(t, conf.Contexts, "second_context")
	})

	t.Run("testModifyContext", func(t *testing.T) {
		conf := testutil.DummyConfig()
		dummyContextOptions := testutil.DummyContextOptions()
		dummyContextOptions.Namespace = "new_namespace"

		modified, err := config.RunSetContext(dummyContextOptions, conf, false)
		assert.NoError(t, err)
		assert.True(t, modified)
		assert.Equal(t, "new_namespace", conf.Contexts["dummy_context"].KubeContext().Namespace)
	})
}

func TestRunUseContext(t *testing.T) {
	t.Run("testUseContext", func(t *testing.T) {
		conf := testutil.DummyConfig()
		err := config.RunUseContext("dummy_context", conf)
		assert.Nil(t, err)
	})

	t.Run("testUseContextDoesNotExist", func(t *testing.T) {
		conf := config.NewConfig()
		err := config.RunUseContext("foo", conf)
		assert.Error(t, err)
	})
}

func TestRunSetManifest(t *testing.T) {
	t.Run("testAddManifest", func(t *testing.T) {
		conf := testutil.DummyConfig()
		dummyManifestOptions := testutil.DummyManifestOptions()
		dummyManifestOptions.Name = "test_manifest"

		modified, err := config.RunSetManifest(dummyManifestOptions, conf, false)
		assert.NoError(t, err)
		assert.False(t, modified)
	})

	t.Run("testModifyManifest", func(t *testing.T) {
		conf := testutil.DummyConfig()
		dummyManifestOptions := testutil.DummyManifestOptions()
		dummyManifestOptions.TargetPath = "/tmp/default"

		modified, err := config.RunSetManifest(dummyManifestOptions, conf, false)
		assert.NoError(t, err)
		assert.True(t, modified)
		assert.Equal(t, "/tmp/default", conf.Manifests["dummy_manifest"].TargetPath)
	})
}
