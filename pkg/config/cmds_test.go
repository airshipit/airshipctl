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

package config

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunGetAuthInfo(t *testing.T) {
	t.Run("testNonExistentAuthInfo", func(t *testing.T) {
		conf := DummyConfig()
		dummyAuthInfoOptions := DummyAuthInfoOptions()
		dummyAuthInfoOptions.Name = "nonexistent_user"
		output := new(bytes.Buffer)
		err := RunGetAuthInfo(dummyAuthInfoOptions, output, conf)
		assert.Error(t, err)
	})

	t.Run("testSingleAuthInfo", func(t *testing.T) {
		conf := DummyConfig()
		dummyAuthInfoOptions := DummyAuthInfoOptions()
		output := new(bytes.Buffer)
		err := RunGetAuthInfo(dummyAuthInfoOptions, output, conf)
		expectedOutput := conf.AuthInfos["dummy_user"].String() + "\n"
		assert.NoError(t, err)
		assert.Equal(t, expectedOutput, output.String())
	})

	t.Run("testAllAuthInfo", func(t *testing.T) {
		conf := DummyConfig()
		secondAuthInfo := DummyAuthInfo()
		secondUserName := "second_user"
		secondAuthInfo.kAuthInfo.Username = secondUserName
		conf.AuthInfos[secondUserName] = secondAuthInfo

		dummyAuthInfoOptions := DummyAuthInfoOptions()
		dummyAuthInfoOptions.Name = ""
		output := new(bytes.Buffer)
		err := RunGetAuthInfo(dummyAuthInfoOptions, output, conf)
		expectedOutput := conf.AuthInfos["dummy_user"].String() + "\n" + conf.AuthInfos[secondUserName].String() + "\n"
		assert.NoError(t, err)
		assert.Equal(t, expectedOutput, output.String())
	})

	t.Run("testNoAuthInfos", func(t *testing.T) {
		conf := DummyConfig()
		dummyAuthInfoOptions := DummyAuthInfoOptions()
		dummyAuthInfoOptions.Name = ""
		delete(conf.AuthInfos, "dummy_user")
		output := new(bytes.Buffer)
		err := RunGetAuthInfo(dummyAuthInfoOptions, output, conf)
		expectedMessage := "No User credentials found in the configuration.\n"
		assert.NoError(t, err)
		assert.Equal(t, expectedMessage, output.String())
	})
}

func TestRunGetCluster(t *testing.T) {
	const dummyClusterEphemeralName = "dummy_cluster_ephemeral"

	t.Run("testNonExistentCluster", func(t *testing.T) {
		conf := DummyConfig()
		dummyClusterOptions := DummyClusterOptions()
		dummyClusterOptions.Name = "nonexistent_cluster"
		output := new(bytes.Buffer)
		err := RunGetCluster(dummyClusterOptions, output, conf)
		assert.Error(t, err)
	})

	t.Run("testSingleCluster", func(t *testing.T) {
		conf := DummyConfig()
		dummyClusterOptions := DummyClusterOptions()
		output := new(bytes.Buffer)
		err := RunGetCluster(dummyClusterOptions, output, conf)
		expectedCluster := DummyCluster()
		expectedCluster.NameInKubeconf = dummyClusterEphemeralName
		assert.NoError(t, err)
		assert.Equal(t, expectedCluster.PrettyString(), output.String())
	})

	t.Run("testAllClusters", func(t *testing.T) {
		conf := DummyConfig()
		dummyClusterOptions := DummyClusterOptions()
		dummyClusterOptions.Name = ""
		output := new(bytes.Buffer)
		err := RunGetCluster(dummyClusterOptions, output, conf)

		expectedClusterTarget := DummyCluster()
		expectedClusterEphemeral := DummyCluster()
		expectedClusterEphemeral.NameInKubeconf = dummyClusterEphemeralName
		expectedOutput := expectedClusterEphemeral.PrettyString() + "\n" + expectedClusterTarget.PrettyString() + "\n"

		assert.NoError(t, err)
		assert.Equal(t, expectedOutput, output.String())
	})

	t.Run("testNoClusters", func(t *testing.T) {
		conf := DummyConfig()
		dummyClusterOptions := DummyClusterOptions()
		dummyClusterOptions.Name = ""
		delete(conf.Clusters, "dummy_cluster")
		output := new(bytes.Buffer)
		err := RunGetCluster(dummyClusterOptions, output, conf)
		expectedMessage := "No clusters found in the configuration.\n"
		assert.NoError(t, err)
		assert.Equal(t, expectedMessage, output.String())
	})
}

func TestRunGetContext(t *testing.T) {
	t.Run("testNonExistentContext", func(t *testing.T) {
		conf := DummyConfig()
		dummyContextOptions := DummyContextOptions()
		dummyContextOptions.Name = "nonexistent_context"
		output := new(bytes.Buffer)
		err := RunGetContext(dummyContextOptions, output, conf)
		assert.Error(t, err)
	})

	t.Run("testSingleContext", func(t *testing.T) {
		conf := DummyConfig()
		dummyContextOptions := DummyContextOptions()
		output := new(bytes.Buffer)
		err := RunGetContext(dummyContextOptions, output, conf)
		assert.NoError(t, err)
		assert.Equal(t, conf.Contexts["dummy_context"].PrettyString(), output.String())
	})

	t.Run("testAllContext", func(t *testing.T) {
		conf := DummyConfig()
		newCtx := DummyContext()
		newCtx.NameInKubeconf = "second_context"
		conf.Contexts["second_context"] = newCtx

		dummyContextOptions := DummyContextOptions()
		dummyContextOptions.Name = ""
		output := new(bytes.Buffer)
		err := RunGetContext(dummyContextOptions, output, conf)
		expectedOutput := conf.Contexts["dummy_context"].PrettyString() + conf.Contexts["second_context"].PrettyString()
		assert.NoError(t, err)
		assert.Equal(t, expectedOutput, output.String())
	})

	t.Run("testNoContexts", func(t *testing.T) {
		conf := DummyConfig()
		delete(conf.Contexts, "dummy_context")
		dummyContextOptions := DummyContextOptions()
		dummyContextOptions.Name = ""
		output := new(bytes.Buffer)
		err := RunGetContext(dummyContextOptions, output, conf)
		expectedOutput := "No Contexts found in the configuration.\n"
		assert.NoError(t, err)
		assert.Equal(t, expectedOutput, output.String())
	})
}

func TestRunSetAuthInfo(t *testing.T) {
	t.Run("testAddAuthInfo", func(t *testing.T) {
		conf := DummyConfig()
		dummyAuthInfoOptions := DummyAuthInfoOptions()
		dummyAuthInfoOptions.Name = "second_user"
		dummyAuthInfoOptions.Token = ""

		modified, err := RunSetAuthInfo(dummyAuthInfoOptions, conf, false)
		assert.NoError(t, err)
		assert.False(t, modified)
		assert.Contains(t, conf.AuthInfos, "second_user")
	})

	t.Run("testModifyAuthInfo", func(t *testing.T) {
		conf := DummyConfig()
		dummyAuthInfoOptions := DummyAuthInfoOptions()
		dummyAuthInfoOptions.Name = "dummy_user"
		dummyAuthInfoOptions.Password = "testpassword123"
		dummyAuthInfoOptions.Token = ""

		modified, err := RunSetAuthInfo(dummyAuthInfoOptions, conf, false)
		assert.NoError(t, err)
		assert.True(t, modified)
		assert.Equal(t, dummyAuthInfoOptions.Password, conf.AuthInfos["dummy_user"].kAuthInfo.Password)
	})
}

func TestRunSetCluster(t *testing.T) {
	t.Run("testAddCluster", func(t *testing.T) {
		conf := DummyConfig()
		dummyClusterOptions := DummyClusterOptions()
		dummyClusterOptions.Name = "second_cluster"

		modified, err := RunSetCluster(dummyClusterOptions, conf, false)
		assert.NoError(t, err)
		assert.False(t, modified)
		assert.Contains(t, conf.Clusters, "second_cluster")
	})

	t.Run("testModifyCluster", func(t *testing.T) {
		conf := DummyConfig()
		dummyClusterOptions := DummyClusterOptions()
		dummyClusterOptions.Server = "http://123.45.67.890"

		modified, err := RunSetCluster(dummyClusterOptions, conf, false)
		assert.NoError(t, err)
		assert.True(t, modified)
		assert.Equal(
			t, "http://123.45.67.890",
			conf.Clusters["dummy_cluster"].ClusterTypes["ephemeral"].kCluster.Server)
	})
}

func TestRunSetContext(t *testing.T) {
	t.Run("testAddContext", func(t *testing.T) {
		conf := DummyConfig()
		dummyContextOptions := DummyContextOptions()
		dummyContextOptions.Name = "second_context"

		modified, err := RunSetContext(dummyContextOptions, conf, false)
		assert.NoError(t, err)
		assert.False(t, modified)
		assert.Contains(t, conf.Contexts, "second_context")
	})

	t.Run("testModifyContext", func(t *testing.T) {
		conf := DummyConfig()
		dummyContextOptions := DummyContextOptions()
		dummyContextOptions.Namespace = "new_namespace"

		modified, err := RunSetContext(dummyContextOptions, conf, false)
		assert.NoError(t, err)
		assert.True(t, modified)
		assert.Equal(t, "new_namespace", conf.Contexts["dummy_context"].kContext.Namespace)
	})
}
