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
	"errors"
	"fmt"
	"io/ioutil"
)

// Validate that the arguments are correct
func (o *ClusterOptions) Validate() error {
	if len(o.Name) == 0 {
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
	caPath := o.CertificateAuthority
	if caPath == "" {
		return fmt.Errorf("you must specify a --%s to embed", FlagCAFile)
	}
	if _, err := ioutil.ReadFile(caPath); err != nil {
		return fmt.Errorf("could not read %s data from %s: %v", FlagCAFile, caPath, err)
	}
	return nil
}

func (o *ContextOptions) Validate() error {
	if len(o.Name) == 0 {
		return errors.New("you must specify a non-empty context name")
	}
	// Expect ClusterType only when this is not setting currentContext
	if o.ClusterType != "" {
		err := ValidClusterType(o.ClusterType)
		if err != nil {
			return err
		}
	}
	// TODO Manifest, Cluster could be validated against the existing config maps
	return nil
}
