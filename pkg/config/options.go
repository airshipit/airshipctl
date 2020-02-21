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

package config

// TODO(howell): Switch to strongly-typed errors

import (
	"errors"
	"fmt"
	"os"
)

type AuthInfoOptions struct {
	Name              string
	ClientCertificate string
	ClientKey         string
	Token             string
	Username          string
	Password          string
	EmbedCertData     bool
}

type ContextOptions struct {
	Name           string
	ClusterType    string
	CurrentContext bool
	Cluster        string
	AuthInfo       string
	Manifest       string
	Namespace      string
	Current        bool
}

type ClusterOptions struct {
	Name                  string
	ClusterType           string
	Server                string
	InsecureSkipTLSVerify bool
	CertificateAuthority  string
	EmbedCAData           bool
}

func (o *AuthInfoOptions) Validate() error {
	if o.Token != "" && (o.Username != "" || o.Password != "") {
		return fmt.Errorf("you cannot specify more than one authentication method at the same time: --%v or --%v/--%v",
			FlagBearerToken, FlagUsername, FlagPassword)
	}

	if !o.EmbedCertData {
		return nil
	}

	if err := checkExists(FlagCertFile, o.ClientCertificate); err != nil {
		return err
	}

	if err := checkExists(FlagKeyFile, o.ClientKey); err != nil {
		return err
	}

	return nil
}

func (o *ContextOptions) Validate() error {
	if !o.Current && o.Name == "" {
		return errors.New("you must specify a non-empty context name")
	}

	if o.Current && o.Name != "" {
		return fmt.Errorf("you cannot specify context and --%s Flag at the same time", FlagCurrent)
	}

	// If the user simply wants to change the current context, no further validation is needed
	if o.CurrentContext {
		return nil
	}

	// If the cluster-type was specified, verify that it's valid
	if o.ClusterType != "" {
		if err := ValidClusterType(o.ClusterType); err != nil {
			return err
		}
	}

	// TODO Manifest, Cluster could be validated against the existing config maps
	return nil
}

func (o *ClusterOptions) Validate() error {
	if o.Name == "" {
		return errors.New("you must specify a non-empty cluster name")
	}

	err := ValidClusterType(o.ClusterType)
	if err != nil {
		return err
	}

	if o.InsecureSkipTLSVerify && o.CertificateAuthority != "" {
		return fmt.Errorf("you cannot specify a %s and %s mode at the same time", FlagCAFile, FlagInsecure)
	}

	if !o.EmbedCAData {
		return nil
	}

	if err := checkExists(FlagCAFile, o.CertificateAuthority); err != nil {
		return err
	}

	return nil
}

func checkExists(flagName, path string) error {
	if path == "" {
		return fmt.Errorf("you must specify a --%s to embed", flagName)
	}
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("could not read %s data from '%s': %v", flagName, path, err)
	}
	return nil
}
