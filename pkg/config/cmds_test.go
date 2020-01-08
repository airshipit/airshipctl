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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateCluster(t *testing.T) {
	co := DummyClusterOptions()

	// Assert that the initial dummy config is valid
	err := co.Validate()
	assert.NoError(t, err)

	// Validate with Embedded Data
	// Empty CA
	co.EmbedCAData = true
	co.CertificateAuthority = ""
	err = co.Validate()
	assert.Error(t, err)

	// Lets add a CA
	co.CertificateAuthority = "testdata/ca.crt"
	err = co.Validate()
	assert.NoError(t, err)

	// Lets add a CA but garbage
	co.CertificateAuthority = "garbage"
	err = co.Validate()
	assert.Error(t, err)

	// Lets change the Insecure mode
	co.InsecureSkipTLSVerify = true
	err = co.Validate()
	assert.Error(t, err)

	// Invalid Cluter Type
	co.ClusterType = "Invalid"
	err = co.Validate()
	assert.Error(t, err)

	// Empty Cluster Name case
	co.Name = ""
	err = co.Validate()
	assert.Error(t, err)
}

func TestValidateContext(t *testing.T) {
	co := DummyContextOptions()
	// Valid Data case
	err := co.Validate()
	assert.NoError(t, err)
}
