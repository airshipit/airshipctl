/*
Copyright 2020 The Kubernetes Authors.

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
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"opendev.org/airship/airshipctl/pkg/config"
)

func TestAuthInfoOptionsValidate(t *testing.T) {
	aRealFile, err := ioutil.TempFile("", "a-real-file")
	require.NoError(t, err)

	aRealFilename := aRealFile.Name()
	defer os.Remove(aRealFilename)

	tests := []struct {
		name        string
		testOptions config.AuthInfoOptions
		expectError bool
	}{
		{
			name: "TokenAndUserPass",
			testOptions: config.AuthInfoOptions{
				Token:    "testToken",
				Username: "testUser",
				Password: "testPassword",
			},
			expectError: true,
		},
		{
			name: "DontEmbed",
			testOptions: config.AuthInfoOptions{
				EmbedCertData: false,
			},
			expectError: false,
		},
		{
			name: "EmbedWithoutCert",
			testOptions: config.AuthInfoOptions{
				EmbedCertData: true,
			},
			expectError: true,
		},
		{
			name: "EmbedWithoutClientKey",
			testOptions: config.AuthInfoOptions{
				EmbedCertData:     true,
				ClientCertificate: aRealFilename,
			},
			expectError: true,
		},
		{
			name: "EmbedWithCertAndClientKey",
			testOptions: config.AuthInfoOptions{
				EmbedCertData:     true,
				ClientCertificate: aRealFilename,
				ClientKey:         aRealFilename,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(subTest *testing.T) {
			err := tt.testOptions.Validate()
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestContextOptionsValidate(t *testing.T) {
	tests := []struct {
		name        string
		testOptions config.ContextOptions
		expectError bool
	}{
		{
			name: "MissingName",
			testOptions: config.ContextOptions{
				Name: "",
			},
			expectError: true,
		},
		{
			name: "InvalidClusterType",
			testOptions: config.ContextOptions{
				Name:        "testContext",
				ClusterType: "badType",
			},
			expectError: true,
		},
		{
			name: "SettingCurrentContext",
			testOptions: config.ContextOptions{
				Name:           "testContext",
				CurrentContext: true,
			},
			expectError: false,
		},
		{
			name: "NoClusterType",
			testOptions: config.ContextOptions{
				Name: "testContext",
			},
			expectError: false,
		},
		{
			name: "ValidClusterType",
			testOptions: config.ContextOptions{
				Name:        "testContext",
				ClusterType: "target",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(subTest *testing.T) {
			err := tt.testOptions.Validate()
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestClusterOptionsValidate(t *testing.T) {
	aRealfile, err := ioutil.TempFile("", "a-real-file")
	require.NoError(t, err)

	aRealFilename := aRealfile.Name()
	defer os.Remove(aRealFilename)

	tests := []struct {
		name        string
		testOptions config.ClusterOptions
		expectError bool
	}{
		{
			name: "MissingName",
			testOptions: config.ClusterOptions{
				Name: "",
			},
			expectError: true,
		},
		{
			name: "InvalidClusterType",
			testOptions: config.ClusterOptions{
				Name:        "testCluster",
				ClusterType: "badType",
			},
			expectError: true,
		},
		{
			name: "InsecureSkipTLSVerifyAndCertificateAuthority",
			testOptions: config.ClusterOptions{
				Name:                  "testCluster",
				ClusterType:           "target",
				InsecureSkipTLSVerify: true,
				CertificateAuthority:  "cert_file",
			},
			expectError: true,
		},
		{
			name: "DontEmbed",
			testOptions: config.ClusterOptions{
				Name:        "testCluster",
				ClusterType: "target",
				EmbedCAData: false,
			},
			expectError: false,
		},
		{
			name: "EmbedWithoutCA",
			testOptions: config.ClusterOptions{
				Name:        "testCluster",
				ClusterType: "target",
				EmbedCAData: true,
			},
			expectError: true,
		},
		{
			name: "EmbedWithFaultyCA",
			testOptions: config.ClusterOptions{
				Name:                 "testCluster",
				ClusterType:          "target",
				EmbedCAData:          true,
				CertificateAuthority: "not-a-real-file",
			},
			expectError: true,
		},
		{
			name: "EmbedWithGoodCA",
			testOptions: config.ClusterOptions{
				Name:                 "testCluster",
				ClusterType:          "target",
				EmbedCAData:          true,
				CertificateAuthority: aRealFilename,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(subTest *testing.T) {
			err := tt.testOptions.Validate()
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
