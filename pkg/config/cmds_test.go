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
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/testutil"
)

func TestRunGetAuthInfo(t *testing.T) {
	t.Run("testNonExistentAuthInfo", func(t *testing.T) {
		conf := testutil.DummyConfig()
		dummyAuthInfoOptions := testutil.DummyAuthInfoOptions()
		dummyAuthInfoOptions.Name = "nonexistent_user"
		output := new(bytes.Buffer)
		err := config.RunGetAuthInfo(dummyAuthInfoOptions, output, conf)
		assert.Error(t, err)
	})

	t.Run("testSingleAuthInfo", func(t *testing.T) {
		conf := testutil.DummyConfig()
		dummyAuthInfoOptions := testutil.DummyAuthInfoOptions()
		output := new(bytes.Buffer)
		err := config.RunGetAuthInfo(dummyAuthInfoOptions, output, conf)
		expectedOutput := conf.AuthInfos["dummy_user"].String() + "\n"
		assert.NoError(t, err)
		assert.Equal(t, expectedOutput, output.String())
	})

	t.Run("testAllAuthInfo", func(t *testing.T) {
		conf := testutil.DummyConfig()
		secondAuthInfo := testutil.DummyAuthInfo()
		secondUserName := "second_user"
		newKubeAuthInfo := testutil.DummyKubeAuthInfo()
		newKubeAuthInfo.Username = secondUserName
		secondAuthInfo.SetKubeAuthInfo(newKubeAuthInfo)
		conf.AuthInfos[secondUserName] = secondAuthInfo

		dummyAuthInfoOptions := testutil.DummyAuthInfoOptions()
		dummyAuthInfoOptions.Name = ""
		output := new(bytes.Buffer)
		err := config.RunGetAuthInfo(dummyAuthInfoOptions, output, conf)
		expectedOutput := conf.AuthInfos["dummy_user"].String() + "\n" + conf.AuthInfos[secondUserName].String() + "\n"
		assert.NoError(t, err)
		assert.Equal(t, expectedOutput, output.String())
	})

	t.Run("testNoAuthInfos", func(t *testing.T) {
		conf := testutil.DummyConfig()
		dummyAuthInfoOptions := testutil.DummyAuthInfoOptions()
		dummyAuthInfoOptions.Name = ""
		delete(conf.AuthInfos, "dummy_user")
		output := new(bytes.Buffer)
		err := config.RunGetAuthInfo(dummyAuthInfoOptions, output, conf)
		expectedMessage := "No User credentials found in the configuration.\n"
		assert.NoError(t, err)
		assert.Equal(t, expectedMessage, output.String())
	})
}

func TestRunGetCluster(t *testing.T) {
	const dummyClusterEphemeralName = "dummy_cluster_ephemeral"

	t.Run("testNonExistentCluster", func(t *testing.T) {
		conf := testutil.DummyConfig()
		dummyClusterOptions := testutil.DummyClusterOptions()
		dummyClusterOptions.Name = "nonexistent_cluster"
		output := new(bytes.Buffer)
		err := config.RunGetCluster(dummyClusterOptions, output, conf)
		assert.Error(t, err)
	})

	t.Run("testSingleCluster", func(t *testing.T) {
		conf := testutil.DummyConfig()
		dummyClusterOptions := testutil.DummyClusterOptions()
		output := new(bytes.Buffer)
		err := config.RunGetCluster(dummyClusterOptions, output, conf)
		expectedCluster := testutil.DummyCluster()
		expectedCluster.NameInKubeconf = dummyClusterEphemeralName
		assert.NoError(t, err)
		assert.Equal(t, expectedCluster.PrettyString(), output.String())
	})

	t.Run("testAllClusters", func(t *testing.T) {
		conf := testutil.DummyConfig()
		dummyClusterOptions := testutil.DummyClusterOptions()
		dummyClusterOptions.Name = ""
		output := new(bytes.Buffer)
		err := config.RunGetCluster(dummyClusterOptions, output, conf)

		expectedClusterTarget := testutil.DummyCluster()
		expectedClusterEphemeral := testutil.DummyCluster()
		expectedClusterEphemeral.NameInKubeconf = dummyClusterEphemeralName
		expectedOutput := expectedClusterEphemeral.PrettyString() + "\n" + expectedClusterTarget.PrettyString() + "\n"

		assert.NoError(t, err)
		assert.Equal(t, expectedOutput, output.String())
	})

	t.Run("testNoClusters", func(t *testing.T) {
		conf := testutil.DummyConfig()
		dummyClusterOptions := testutil.DummyClusterOptions()
		dummyClusterOptions.Name = ""
		delete(conf.Clusters, "dummy_cluster")
		output := new(bytes.Buffer)
		err := config.RunGetCluster(dummyClusterOptions, output, conf)
		expectedMessage := "No clusters found in the configuration.\n"
		assert.NoError(t, err)
		assert.Equal(t, expectedMessage, output.String())
	})
}

func TestRunGetContext(t *testing.T) {
	t.Run("testNonExistentContext", func(t *testing.T) {
		conf := testutil.DummyConfig()
		dummyContextOptions := testutil.DummyContextOptions()
		dummyContextOptions.Name = "nonexistent_context"
		output := new(bytes.Buffer)
		err := config.RunGetContext(dummyContextOptions, output, conf)
		assert.Error(t, err)
	})

	t.Run("testSingleContext", func(t *testing.T) {
		conf := testutil.DummyConfig()
		dummyContextOptions := testutil.DummyContextOptions()
		output := new(bytes.Buffer)
		err := config.RunGetContext(dummyContextOptions, output, conf)
		assert.NoError(t, err)
		assert.Equal(t, conf.Contexts["dummy_context"].PrettyString(), output.String())
	})

	t.Run("testAllContext", func(t *testing.T) {
		conf := testutil.DummyConfig()
		newCtx := testutil.DummyContext()
		newCtx.NameInKubeconf = "second_context"
		conf.Contexts["second_context"] = newCtx

		dummyContextOptions := testutil.DummyContextOptions()
		dummyContextOptions.Name = ""
		output := new(bytes.Buffer)
		err := config.RunGetContext(dummyContextOptions, output, conf)
		expectedOutput := conf.Contexts["dummy_context"].PrettyString() + conf.Contexts["second_context"].PrettyString()
		assert.NoError(t, err)
		assert.Equal(t, expectedOutput, output.String())
	})

	t.Run("testNoContexts", func(t *testing.T) {
		conf := testutil.DummyConfig()
		delete(conf.Contexts, "dummy_context")
		dummyContextOptions := testutil.DummyContextOptions()
		dummyContextOptions.Name = ""
		output := new(bytes.Buffer)
		err := config.RunGetContext(dummyContextOptions, output, conf)
		expectedOutput := "No Contexts found in the configuration.\n"
		assert.NoError(t, err)
		assert.Equal(t, expectedOutput, output.String())
	})
}

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
